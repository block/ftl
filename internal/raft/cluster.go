package raft

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/jpillora/backoff"
	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/config"
	"github.com/lni/dragonboat/v4/statemachine"

	raftpb "github.com/block/ftl/backend/protos/xyz/block/ftl/raft/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/raft/v1/raftpbconnect"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/retry"
	"github.com/block/ftl/internal/rpc"
	sm "github.com/block/ftl/internal/statemachine"
)

// maxJoinAttempts is the maximum number of times to attempt to join the cluster.
const maxJoinAttempts = 5

type RaftConfig struct {
	InitialMembers    []string          `help:"Initial members" env:"RAFT_INITIAL_MEMBERS" and:"raft"`
	InitialReplicaIDs []uint64          `name:"initial-replica-ids" help:"Initial replica IDs" env:"RAFT_INITIAL_REPLICA_IDS" and:"raft"`
	DataDir           string            `help:"Data directory" env:"RAFT_DATA_DIR" and:"raft"`
	Address           string            `help:"Address to advertise to other nodes" env:"RAFT_ADDRESS" and:"raft"`
	ListenAddress     string            `help:"Address to listen for incoming traffic. If empty, Address will be used." env:"RAFT_LISTEN_ADDRESS"`
	ControlAddress    *url.URL          `help:"Address to connect to the control server" env:"RAFT_CONTROL_ADDRESS"`
	ShardReadyTimeout time.Duration     `help:"Timeout for shard to be ready" default:"5s"`
	Retry             retry.RetryConfig `help:"Connection retry configuration" prefix:"retry-" embed:""`
	ChangesInterval   time.Duration     `help:"Interval for changes to be checked" default:"10ms"`
	ChangesTimeout    time.Duration     `help:"Timeout for changes to be checked" default:"1s"`
	QueryTimeout      time.Duration     `help:"Timeout for queries" default:"5s"`

	// Raft configuration
	RTT                time.Duration `help:"Estimated average round trip time between nodes" default:"200ms"`
	ElectionRTT        uint64        `help:"Election RTT as a multiple of RTT" default:"10"`
	HeartbeatRTT       uint64        `help:"Heartbeat RTT as a multiple of RTT" default:"1"`
	SnapshotEntries    uint64        `help:"Snapshot entries" default:"100"`
	CompactionOverhead uint64        `help:"Compaction overhead" default:"100"`
}

// Builder for a Raft Cluster.
type Builder struct {
	config        *RaftConfig
	shards        map[uint64]statemachine.CreateStateMachineFunc
	controlClient *http.Client

	handles []*ShardHandle[any, any, sm.Marshallable]
}

func NewBuilder(cfg *RaftConfig) *Builder {
	return &Builder{
		config: cfg,
		shards: map[uint64]statemachine.CreateStateMachineFunc{},
	}
}

// WithControlClient sets the http client used to communicate with
// the control plane.
func (b *Builder) WithControlClient(client *http.Client) *Builder {
	b.controlClient = client
	return b
}

// AddShard adds a shard to the cluster Builder.
func AddShard[Q any, R any, E sm.Marshallable, EPtr sm.Unmarshallable[E]](
	ctx context.Context,
	to *Builder,
	shardID uint64,
	statemachine sm.Snapshotting[Q, R, E],
) sm.Handle[Q, R, E] {
	to.shards[shardID] = newStateMachineShim[Q, R, E, EPtr](statemachine)

	handle := &ShardHandle[Q, R, E]{
		shardID: shardID,
	}
	to.handles = append(to.handles, (*ShardHandle[any, any, sm.Marshallable])(handle))
	return handle
}

// Cluster of dragonboat nodes.
type Cluster struct {
	config           *RaftConfig
	nh               *dragonboat.NodeHost
	shards           map[uint64]statemachine.CreateStateMachineFunc
	controlClient    *http.Client
	runtimeReplicaID uint64

	// runningCtx is cancelled when the cluster is stopped.
	runningCtx       context.Context
	runningCtxCancel context.CancelCauseFunc
}

var _ raftpbconnect.RaftServiceHandler = (*Cluster)(nil)

func (b *Builder) Build(ctx context.Context) *Cluster {
	controlClient := b.controlClient
	if controlClient == nil {
		controlClient = http.DefaultClient
	}

	cluster := &Cluster{
		config:        b.config,
		shards:        b.shards,
		controlClient: controlClient,
	}

	for _, handle := range b.handles {
		handle.cluster = cluster
	}

	return cluster
}

// ShardHandle is a handle to a shard in the cluster.
// It is the interface to update and query the state of a shard.
//
// E is the event type.
// Q is the query type.
// R is the query response type.
type ShardHandle[Q any, R any, E sm.Marshallable] struct {
	shardID uint64
	cluster *Cluster
	session *client.Session

	mu sync.Mutex
}

var _ sm.Handle[any, any, sm.Marshallable] = (*ShardHandle[any, any, sm.Marshallable])(nil)

// Publish an event to the shard.
func (s *ShardHandle[Q, R, E]) Publish(ctx context.Context, msg E) error {
	// client session is not thread safe, so we need to lock
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := log.FromContext(ctx).Scope("raft")

	s.verifyReady()

	msgBytes, err := msg.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	if s.session == nil {
		if err := s.cluster.withRetry(ctx, s.shardID, s.cluster.runtimeReplicaID, func(ctx context.Context) error {
			logger.Debugf("Getting session for shard %d on replica %d", s.shardID, s.cluster.runtimeReplicaID)
			s.session, err = s.cluster.nh.SyncGetSession(ctx, s.shardID)
			if err != nil {
				return err //nolint:wrapcheck
			}
			logger.Debugf("Got clientID %d for shard %d on replica %d", s.session.ClientID, s.shardID, s.cluster.runtimeReplicaID)
			return nil
		}, dragonboat.ErrShardNotReady, dragonboat.ErrTimeout, dragonboat.ErrRejected); err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
	}

	if err := s.cluster.withRetry(ctx, s.shardID, s.cluster.runtimeReplicaID, func(ctx context.Context) error {
		logger.Debugf("Proposing event to shard %d on replica %d", s.shardID, s.cluster.runtimeReplicaID)
		res, err := s.cluster.nh.SyncPropose(ctx, s.session, msgBytes)
		if err != nil {
			return err //nolint:wrapcheck
		}
		s.session.ProposalCompleted()

		if res.Value == InvalidEventValue {
			return ErrInvalidEvent
		}
		return nil
	}, dragonboat.ErrShardNotReady, dragonboat.ErrTimeout); err != nil {
		return fmt.Errorf("failed to propose event: %w", err)
	}

	return nil
}

// Query the state of the shard.
func (s *ShardHandle[Q, R, E]) Query(ctx context.Context, query Q) (R, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cluster.config.QueryTimeout)
	defer cancel()

	s.verifyReady()

	var response R
	if err := s.cluster.withRetry(ctx, s.shardID, s.cluster.runtimeReplicaID, func(ctx context.Context) error {
		res, err := s.cluster.nh.SyncRead(ctx, s.shardID, query)
		if err != nil {
			return err //nolint:wrapcheck
		}
		r, ok := res.(R)
		if !ok {
			return fmt.Errorf("invalid response type: %T", res)
		}
		response = r
		return nil
	}, dragonboat.ErrShardNotReady); err != nil {
		return response, err
	}

	return response, nil
}

// StateIter returns an iterator that will return the result of the query when the
// shard state changes.
//
// This can only be called when the cluster is running.
//
// Note, that this is not guaranteed to receive an event for every change, but
// will always receive the latest state of the shard.
func (s *ShardHandle[Q, R, E]) StateIter(ctx context.Context, query Q) (iter.Seq[R], error) {
	s.verifyReady()

	result := make(chan R, 64)
	logger := log.FromContext(ctx).Scope("raft")

	// get the last known index as the starting point
	last, err := s.getLastIndex()
	if err != nil {
		logger.Errorf(err, "Failed to get last index")
	}

	lastKnownIndex := atomic.Value[uint64]{}
	lastKnownIndex.Store(last)

	previous, err := s.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	result <- previous

	go func() {
		// poll, as dragoboat does not have a way to listen to changes directly
		timer := time.NewTicker(s.cluster.config.ChangesInterval)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				close(result)
				return
			case <-s.cluster.runningCtx.Done():
				logger.Infof("Changes channel closed")
				close(result)
				return
			case <-timer.C:
				last, err := s.getLastIndex()
				if err != nil {
					logger.Warnf("Failed to get last index: %s", err)
				} else if last > lastKnownIndex.Load() {
					logger.Debugf("Changes detected, index: %d -> %d on (%d, %d)", lastKnownIndex.Load(), last, s.shardID, s.cluster.runtimeReplicaID)

					lastKnownIndex.Store(last)

					res, err := s.Query(ctx, query)

					if err != nil {
						logger.Errorf(err, "failed to query shard")
					} else {
						logger.Debugf("Publishing to state iterator on (%d, %d)", s.shardID, s.cluster.runtimeReplicaID)
						result <- res
					}
				}
			}
		}
	}()

	// dedup, as we might get false positives due to index changes caused by membership changes
	return iterops.Dedup(channels.IterContext(ctx, result)), nil
}

func RPCOption(cluster *Cluster) rpc.Option {
	return rpc.Options(
		rpc.StartHook(func(ctx context.Context) error {
			if err := cluster.Start(ctx); err != nil {
				return fmt.Errorf("failed to start raft cluster: %w", err)
			}
			return nil
		}),
		rpc.ShutdownHook(func(ctx context.Context) error {
			logger := log.FromContext(ctx)
			logger.Debugf("stopping raft cluster")
			cluster.Stop(ctx)
			return nil
		}),
		rpc.GRPC(raftpbconnect.NewRaftServiceHandler, cluster),
	)
}

func (s *ShardHandle[Q, R, E]) getLastIndex() (uint64, error) {
	s.verifyReady()

	reader, err := s.cluster.nh.GetLogReader(s.shardID)
	if err != nil {
		return 0, fmt.Errorf("failed to get log reader: %w", err)
	}
	_, last := reader.GetRange()
	return last, nil
}

func (s *ShardHandle[Q, R, E]) verifyReady() {
	if s.cluster == nil {
		panic("cluster not built")
	}
	if s.cluster.nh == nil {
		panic("cluster not started")
	}
}

// Start the cluster. Blocks until the cluster instance is ready.
func (c *Cluster) Start(ctx context.Context) error {
	logger := log.FromContext(ctx).Scope("raft")
	if c.nh != nil {
		panic("cluster already started")
	}

	if c.initialMemberIndex() < 0 {
		logger.Infof("joining cluster as a new member")
		return c.Join(ctx, c.config.ControlAddress.String())
	}

	r, err := c.getOrGenerateReplicaID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get or generate replica ID: %w", err)
	}
	c.runtimeReplicaID = r

	logger.Infof("joining cluster as an initial member")
	return c.start(ctx, false)
}

func (c *Cluster) initialMemberIndex() int {
	for i, member := range c.config.InitialMembers {
		if member == c.config.Address {
			return i
		}
	}
	return -1
}

func (c *Cluster) RuntimeReplicaID() uint64 {
	return c.runtimeReplicaID
}

// Join the cluster as a new member. Blocks until the cluster instance is ready.
func (c *Cluster) Join(ctx context.Context, controlAddress string) error {
	logger := log.FromContext(ctx).Scope("raft")

	r, err := c.getOrGenerateReplicaID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get or generate replica ID: %w", err)
	}
	c.runtimeReplicaID = r

	// call control server to join the cluster
	client := raftpbconnect.NewRaftServiceClient(http.DefaultClient, controlAddress)

	shardIDs := make([]uint64, 0, len(c.shards))
	for shardID := range c.shards {
		shardIDs = append(shardIDs, shardID)
	}

	retry := c.config.Retry.Backoff()
	for i := 0; true; i++ {
		_, err := client.AddMember(ctx, connect.NewRequest(&raftpb.AddMemberRequest{
			ShardIds:  shardIDs,
			ReplicaId: c.runtimeReplicaID,
			Address:   c.config.Address,
		}))
		if err != nil {
			if i >= maxJoinAttempts {
				return fmt.Errorf("failed to join cluster: %w", err)
			}

			duration := retry.Duration()
			logger.Warnf("failed to join cluster: %s, retrying in %s", err, duration)
			time.Sleep(duration)
			continue
		}
		break
	}

	return c.start(ctx, true)
}

func (c *Cluster) getOrGenerateReplicaID(ctx context.Context) (uint64, error) {
	logger := log.FromContext(ctx).Scope("raft")

	initialIndex := c.initialMemberIndex()

	if initialIndex < 0 {
		replicaIDFile := filepath.Join(c.config.DataDir, "replica_id")

		// if file does not exist, generate a random ID
		if _, err := os.Stat(replicaIDFile); os.IsNotExist(err) {
			replicaID, err := randint64()
			if err != nil {
				return 0, fmt.Errorf("failed to generate replica ID: %w", err)
			}
			logger.Infof("generated new replica ID %d", replicaID)
			return replicaID, nil
		}

		replicaIDBytes, err := os.ReadFile(replicaIDFile)
		if err != nil {
			return 0, fmt.Errorf("failed to read replica ID file: %w", err)
		}
		replicaID, err := strconv.ParseUint(string(replicaIDBytes), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse replica ID: %w", err)
		}

		logger.Infof("using existing replica ID %d", replicaID)
		return replicaID, nil
	}

	return c.config.InitialReplicaIDs[initialIndex], nil
}

func (c *Cluster) writeReplicaID(replicaID uint64) error {
	replicaIDFile := filepath.Join(c.config.DataDir, "replica_id")
	if err := os.MkdirAll(c.config.DataDir, 0750); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	if err := os.WriteFile(replicaIDFile, []byte(strconv.FormatUint(replicaID, 10)), 0600); err != nil {
		return fmt.Errorf("failed to write replica ID file: %w", err)
	}

	return nil
}

func (c *Cluster) start(ctx context.Context, join bool) error {
	logger := log.FromContext(ctx).Scope("raft")

	// Create node host config
	nhc := config.NodeHostConfig{
		WALDir:         c.config.DataDir,
		NodeHostDir:    c.config.DataDir,
		RTTMillisecond: uint64(c.config.RTT.Milliseconds()),
		RaftAddress:    c.config.Address,
		ListenAddress:  c.config.ListenAddress,
	}

	// Create node host
	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		return fmt.Errorf("failed to create node host: %w", err)
	}
	c.nh = nh

	// Start replicas for each shard
	for shardID, sm := range c.shards {
		if err := c.startShard(nh, shardID, sm, join); err != nil {
			return err
		}
	}

	// Wait for all shards to be ready
	for shardID := range c.shards {
		err := c.waitReady(ctx, shardID)
		if err != nil {
			return fmt.Errorf("failed to wait for shard %d to be ready on replica %d: %w", shardID, c.runtimeReplicaID, err)
		}
	}
	logger.Infof("All shards are ready")

	if err := c.writeReplicaID(c.runtimeReplicaID); err != nil {
		return fmt.Errorf("failed to write replica ID: %w", err)
	}

	ctx, cancel := context.WithCancelCause(context.WithoutCancel(ctx))
	c.runningCtxCancel = cancel
	c.runningCtx = ctx

	return nil
}

func (c *Cluster) startShard(nh *dragonboat.NodeHost, shardID uint64, sm statemachine.CreateStateMachineFunc, join bool) error {
	if len(c.config.InitialReplicaIDs) != len(c.config.InitialMembers) {
		return fmt.Errorf("initial replica IDs and initial members must be the same length")
	}

	cfg := config.Config{
		ReplicaID:          c.runtimeReplicaID,
		ShardID:            shardID,
		CheckQuorum:        true,
		ElectionRTT:        c.config.ElectionRTT,
		HeartbeatRTT:       c.config.HeartbeatRTT,
		SnapshotEntries:    c.config.SnapshotEntries,
		CompactionOverhead: c.config.CompactionOverhead,
		WaitReady:          true,
	}

	peers := make(map[uint64]string)
	if !join {
		for idx, peer := range c.config.InitialMembers {
			peers[c.config.InitialReplicaIDs[idx]] = peer
		}
	}

	// Start the raft node for this shard
	if err := nh.StartReplica(peers, join, sm, cfg); err != nil {
		return fmt.Errorf("failed to start replica %d for shard %d: %w", c.runtimeReplicaID, shardID, err)
	}
	return nil
}

// Stop the node host and all shards.
// After this call, all the shard handlers created with this cluster are invalid.
func (c *Cluster) Stop(ctx context.Context) {
	logger := log.FromContext(ctx).Scope("raft")
	if c.nh != nil {
		c.runningCtxCancel(fmt.Errorf("stopping raft cluster: %w", context.Canceled))
		c.nh.Close()
		c.nh = nil
		c.shards = nil
	} else {
		logger.Debugf("raft cluster already stopped")
	}
}

// withRetry runs an async dragonboat call and blocks until it succeeds or the context is cancelled.
// the call is retried if the request is dropped, which can happen if the leader is not available.
func (c *Cluster) withRetry(octx context.Context, shardID, replicaID uint64, f func(ctx context.Context) error, retryErrors ...error) error {
	retry := c.config.Retry.Backoff()
	logger := log.FromContext(octx).Scope("raft")

	for {
		// Timeout for the proposal to reach the leader and reach a quorum.
		// If the leader is not available, the proposal will time out, in which case
		// we retry the operation.
		timeout := time.Duration(c.config.ElectionRTT) * c.config.RTT
		ctx, cancel := context.WithTimeout(octx, timeout)
		defer cancel()

		err := f(ctx)
		duration := retry.Duration()

		if err != nil {
			if octx.Err() != nil {
				return fmt.Errorf("context cancelled: %w", octx.Err())
			}
			retried := false
			for _, retryError := range retryErrors {
				if errors.Is(err, retryError) {
					logger.Debugf("Got error %s, retrying in %s", err, duration)
					select {
					case <-time.After(duration):
					case <-ctx.Done():
					}
					cancel()
					retried = true
					break
				}
			}
			if retried {
				continue
			}
			return fmt.Errorf("failed to submit request to shard %d on replica %d: %w", shardID, replicaID, err)
		}
		return nil
	}
}

func (c *Cluster) waitReady(ctx context.Context, shardID uint64) error {
	retry := backoff.Backoff{
		Min:    c.config.RTT,
		Max:    c.config.ShardReadyTimeout,
		Factor: 2,
		Jitter: true,
	}
	logger := log.FromContext(ctx).Scope("raft")
	for {
		wait := retry.Duration()
		rs, err := c.nh.ReadIndex(shardID, wait)
		if err != nil || rs == nil {
			return fmt.Errorf("failed to read index: %w", err)
		}
		res := <-rs.ResultC()
		rs.Release()
		if !res.Completed() {
			logger.Debugf("Waiting for shard %d to be ready on replica %d: %s", shardID, c.runtimeReplicaID, wait)
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(wait):
			}
			continue
		}
		break
	}
	logger.Debugf("Shard %d on replica %d is ready", shardID, c.runtimeReplicaID)
	return nil
}

func randint64() (uint64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, fmt.Errorf("failed to read random bytes: %w", err)
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}
