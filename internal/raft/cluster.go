package raft

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/url"
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
	raftpbconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/raft/v1/raftpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/retry"
	"github.com/block/ftl/internal/rpc"
	sm "github.com/block/ftl/internal/statemachine"
)

type RaftConfig struct {
	InitialMembers    []string          `help:"Initial members"`
	ReplicaID         uint64            `help:"Node ID" required:""`
	DataDir           string            `help:"Data directory" required:""`
	Address           string            `help:"Address to advertise to other nodes" required:""`
	ListenAddress     string            `help:"Address to listen for incoming traffic. If empty, Address will be used."`
	ControlBind       *url.URL          `help:"Address to listen for control traffic. If empty, no control listener will be started."`
	ShardReadyTimeout time.Duration     `help:"Timeout for shard to be ready" default:"5s"`
	Retry             retry.RetryConfig `help:"Connection retry configuration" prefix:"retry-" embed:""`
	ChangesInterval   time.Duration     `help:"Interval for changes to be checked" default:"10ms"`
	ChangesTimeout    time.Duration     `help:"Timeout for changes to be checked" default:"1s"`

	// Raft configuration
	RTT                time.Duration `help:"Estimated average round trip time between nodes" default:"200ms"`
	ElectionRTT        uint64        `help:"Election RTT as a multiple of RTT" default:"10"`
	HeartbeatRTT       uint64        `help:"Heartbeat RTT as a multiple of RTT" default:"1"`
	SnapshotEntries    uint64        `help:"Snapshot entries" default:"10"`
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
	config        *RaftConfig
	nh            *dragonboat.NodeHost
	shards        map[uint64]statemachine.CreateStateMachineFunc
	controlClient *http.Client

	// runningCtx is cancelled when the cluster is stopped.
	runningCtx       context.Context
	runningCtxCancel context.CancelFunc
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

	lastKnownIndex atomic.Value[uint64]

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
		if err := s.cluster.withRetry(ctx, s.shardID, s.cluster.config.ReplicaID, func(ctx context.Context) error {
			logger.Debugf("getting session for shard %d on replica %d", s.shardID, s.cluster.config.ReplicaID)
			s.session, err = s.cluster.nh.SyncGetSession(ctx, s.shardID)
			if err != nil {
				return err //nolint:wrapcheck
			}
			logger.Debugf("got clientID %d for shard %d on replica %d", s.session.ClientID, s.shardID, s.cluster.config.ReplicaID)
			return nil
		}, dragonboat.ErrShardNotReady, dragonboat.ErrTimeout, dragonboat.ErrRejected); err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
	}

	if err := s.cluster.withRetry(ctx, s.shardID, s.cluster.config.ReplicaID, func(ctx context.Context) error {
		logger.Debugf("proposing event to shard %d on replica %d", s.shardID, s.cluster.config.ReplicaID)
		_, err := s.cluster.nh.SyncPropose(ctx, s.session, msgBytes)
		if err != nil {
			return err //nolint:wrapcheck
		}
		s.session.ProposalCompleted()
		return nil
	}, dragonboat.ErrShardNotReady, dragonboat.ErrTimeout); err != nil {
		return fmt.Errorf("failed to propose event: %w", err)
	}

	return nil
}

// Query the state of the shard.
func (s *ShardHandle[Q, R, E]) Query(ctx context.Context, query Q) (R, error) {
	s.verifyReady()

	var zero R

	res, err := s.cluster.nh.SyncRead(ctx, s.shardID, query)
	if err != nil {
		return zero, fmt.Errorf("failed to query shard: %w", err)
	}

	response, ok := res.(R)
	if !ok {
		panic(fmt.Errorf("invalid response type: %T", res))
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
	if s.cluster.nh == nil {
		panic("cluster not started")
	}

	result := make(chan R, 64)
	logger := log.FromContext(ctx).Scope("raft")

	// get the last known index as the starting point
	last, err := s.getLastIndex()
	if err != nil {
		logger.Errorf(err, "failed to get last index")
	}
	s.lastKnownIndex.Store(last)

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
			case <-s.cluster.runningCtx.Done():
				logger.Infof("changes channel closed")
				close(result)

				return
			case <-timer.C:
				last, err := s.getLastIndex()
				if err != nil {
					logger.Warnf("failed to get last index: %s", err)
				} else if last > s.lastKnownIndex.Load() {
					logger.Debugf("changes detected, index: %d -> %d on (%d, %d)", s.lastKnownIndex.Load(), last, s.shardID, s.cluster.config.ReplicaID)

					s.lastKnownIndex.Store(last)

					ctx, cancel := context.WithTimeout(ctx, s.cluster.config.ChangesTimeout)
					res, err := s.Query(ctx, query)
					cancel()

					if err != nil {
						logger.Errorf(err, "failed to query shard")
					} else {
						logger.Debugf("publishing to state iterator on (%d, %d)", s.shardID, s.cluster.config.ReplicaID)
						result <- res
					}
				}
			}
		}
	}()

	// dedup, as we might get false positives due to index changes caused by membership changes
	return iterops.Dedup(channels.IterContext(ctx, result)), nil
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
	if c.nh != nil {
		panic("cluster already started")
	}

	return c.start(ctx, false)
}

// Join the cluster as a new member. Blocks until the cluster instance is ready.
func (c *Cluster) Join(ctx context.Context, controlAddress string) error {
	logger := log.FromContext(ctx).Scope("raft")

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
			ReplicaId: c.config.ReplicaID,
			Address:   c.config.Address,
		}))
		if err != nil {
			duration := retry.Duration()
			logger.Warnf("failed to join cluster: %s, retrying in %s", err, duration)
			time.Sleep(duration)
			continue
		}
		break
	}

	return c.start(ctx, true)
}

func (c *Cluster) start(ctx context.Context, join bool) error {
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
		if err := c.waitReady(ctx, shardID); err != nil {
			return fmt.Errorf("failed to wait for shard %d to be ready on replica %d: %w", shardID, c.config.ReplicaID, err)
		}
	}

	ctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	c.runningCtxCancel = cancel
	c.runningCtx = ctx

	if err := c.startControlServer(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Cluster) startShard(nh *dragonboat.NodeHost, shardID uint64, sm statemachine.CreateStateMachineFunc, join bool) error {
	cfg := config.Config{
		ReplicaID:          c.config.ReplicaID,
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
			peers[uint64(idx+1)] = peer
		}
	}

	// Start the raft node for this shard
	if err := nh.StartReplica(peers, join, sm, cfg); err != nil {
		return fmt.Errorf("failed to start replica %d for shard %d: %w", c.config.ReplicaID, shardID, err)
	}
	return nil
}

func (c *Cluster) startControlServer(ctx context.Context) error {
	logger := log.FromContext(ctx).Scope("raft")

	if c.config.ControlBind == nil {
		return nil
	}

	logger.Infof("starting control server on %s", c.config.ControlBind.String())
	go func() {
		err := rpc.Serve(ctx, c.config.ControlBind,
			rpc.GRPC(raftpbconnect.NewRaftServiceHandler, c),
			rpc.PProf())
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Errorf(err, "error serving control listener")
		}
		logger.Infof("control server stopped")
	}()
	return nil
}

// Stop the node host and all shards.
// After this call, all the shard handlers created with this cluster are invalid.
func (c *Cluster) Stop(ctx context.Context) {
	if c.nh != nil {
		logger := log.FromContext(ctx).Scope("raft")
		logger.Infof("stopping replica %d", c.config.ReplicaID)

		for shardID := range c.shards {
			c.removeShardMember(ctx, shardID, c.config.ReplicaID)
		}
		c.runningCtxCancel()
		c.nh.Close()
		c.nh = nil
		c.shards = nil
	}
}

// AddMember to the cluster. This needs to be called on an existing running cluster member,
// before the new member is started.
func (c *Cluster) AddMember(ctx context.Context, req *connect.Request[raftpb.AddMemberRequest]) (*connect.Response[raftpb.AddMemberResponse], error) {
	logger := log.FromContext(ctx).Scope("raft")

	shards := req.Msg.ShardIds
	replicaID := req.Msg.ReplicaId
	address := req.Msg.Address

	logger.Infof("adding member %s to shard %d on replica %d", address, shards, replicaID)

	for _, shardID := range shards {
		if err := c.withRetry(ctx, shardID, replicaID, func(ctx context.Context) error {
			logger.Debugf("requesting add replica to shard %d on replica %d", shardID, replicaID)
			return c.nh.SyncRequestAddReplica(ctx, shardID, replicaID, address, 0)
		}, dragonboat.ErrShardNotReady, dragonboat.ErrTimeout); err != nil {
			return nil, fmt.Errorf("failed to add member: %w", err)
		}
	}
	return connect.NewResponse(&raftpb.AddMemberResponse{}), nil
}

// removeShardMember from the given shard. This removes the given member from the membership group
// and blocks until the change has been committed
func (c *Cluster) removeShardMember(ctx context.Context, shardID uint64, replicaID uint64) {
	logger := log.FromContext(ctx).Scope("raft")
	logger.Infof("removing replica %d from shard %d", shardID, replicaID)

	if err := c.withRetry(ctx, shardID, replicaID, func(ctx context.Context) error {
		logger.Debugf("requesting delete replica from shard %d on replica %d", shardID, replicaID)
		return c.nh.SyncRequestDeleteReplica(ctx, shardID, replicaID, 0)
	}, dragonboat.ErrShardNotReady, dragonboat.ErrTimeout); err != nil {
		// This can happen if the cluster is shutting down and no longer has quorum.
		logger.Warnf("removing replica %d from shard %d failed: %s", replicaID, shardID, err)
	}
}

// Ping the cluster.
func (c *Cluster) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// withTimeout runs an async dragonboat call and blocks until it succeeds or the context is cancelled.
// the call is retried if the request is dropped, which can happen if the leader is not available.
func (c *Cluster) withRetry(ctx context.Context, shardID, replicaID uint64, f func(ctx context.Context) error, retryErrors ...error) error {
	retry := c.config.Retry.Backoff()
	logger := log.FromContext(ctx).Scope("raft")

	for {
		// Timeout for the proposal to reach the leader and reach a quorum.
		// If the leader is not available, the proposal will time out, in which case
		// we retry the operation.
		timeout := time.Duration(c.config.ElectionRTT) * c.config.RTT
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := f(ctx)
		duration := retry.Duration()

		if err != nil {
			retried := false
			for _, retryError := range retryErrors {
				if errors.Is(err, retryError) {
					logger.Debugf("got error %s, retrying in %s", err, duration)
					time.Sleep(duration)
					if _, ok := <-ctx.Done(); ok {
						return fmt.Errorf("context cancelled")
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
			logger.Debugf("waiting for shard %d to be ready on replica %d: %s", shardID, c.config.ReplicaID, wait)
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled")
			case <-time.After(wait):
			}
			continue
		}
		break
	}
	return nil
}
