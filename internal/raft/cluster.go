package raft

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jpillora/backoff"
	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/config"
	"github.com/lni/dragonboat/v4/statemachine"

	"github.com/block/ftl/internal/log"
)

type RaftConfig struct {
	InitialMembers    []string      `help:"Initial members" required:""`
	ReplicaID         uint64        `help:"Node ID" required:""`
	DataDir           string        `help:"Data directory" required:""`
	RaftAddress       string        `help:"Address to advertise to other nodes" required:""`
	ListenAddress     string        `help:"Address to listen for incoming traffic. If empty, RaftAddress will be used."`
	ShardReadyTimeout time.Duration `help:"Timeout for shard to be ready" default:"5s"`
	// Raft configuration
	RTT                time.Duration `help:"Estimated average round trip time between nodes" default:"200ms"`
	ElectionRTT        uint64        `help:"Election RTT as a multiple of RTT" default:"10"`
	HeartbeatRTT       uint64        `help:"Heartbeat RTT as a multiple of RTT" default:"1"`
	SnapshotEntries    uint64        `help:"Snapshot entries" default:"10"`
	CompactionOverhead uint64        `help:"Compaction overhead" default:"100"`
}

// Builder for a Raft Cluster.
type Builder struct {
	config *RaftConfig
	shards map[uint64]statemachine.CreateStateMachineFunc

	handles []*ShardHandle[Event, any, any]
}

func NewBuilder(cfg *RaftConfig) *Builder {
	return &Builder{
		config: cfg,
		shards: map[uint64]statemachine.CreateStateMachineFunc{},
	}
}

// AddShard adds a shard to the cluster Builder.
func AddShard[Q any, R any, E Event, EPtr Unmarshallable[E]](
	ctx context.Context,
	to *Builder,
	shardID uint64,
	sm StateMachine[Q, R, E, EPtr],
) *ShardHandle[E, Q, R] {
	to.shards[shardID] = newStateMachineShim[Q, R, E, EPtr](sm)

	handle := &ShardHandle[E, Q, R]{
		shardID: shardID,
	}
	to.handles = append(to.handles, (*ShardHandle[Event, any, any])(handle))
	return handle
}

// Cluster of dragonboat nodes.
type Cluster struct {
	config *RaftConfig
	nh     *dragonboat.NodeHost
	shards map[uint64]statemachine.CreateStateMachineFunc
}

func (b *Builder) Build(ctx context.Context) *Cluster {
	cluster := &Cluster{
		config: b.config,
		shards: b.shards,
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
type ShardHandle[E Event, Q any, R any] struct {
	shardID uint64
	cluster *Cluster
	session *client.Session
}

// Propose an event to the shard.
func (s *ShardHandle[E, Q, R]) Propose(ctx context.Context, msg E) error {
	s.verifyReady()

	msgBytes, err := msg.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	if s.session == nil {
		// use a no-op session for now. This means that a retry on timeout could result into duplicate events.
		s.session = s.cluster.nh.GetNoOPSession(s.shardID)
	}

	if err := s.cluster.withRetry(ctx, s.shardID, s.cluster.config.ReplicaID, func() error {
		_, err := s.cluster.nh.SyncPropose(ctx, s.session, msgBytes)
		return err //nolint:wrapcheck
	}); err != nil {
		return fmt.Errorf("failed to propose event: %w", err)
	}

	return nil
}

// Query the state of the shard.
func (s *ShardHandle[E, Q, R]) Query(ctx context.Context, query Q) (R, error) {
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

func (s *ShardHandle[E, Q, R]) verifyReady() {
	if s.cluster == nil {
		panic("cluster not built")
	}
	if s.cluster.nh == nil {
		panic("cluster not started")
	}
}

// Start the cluster. Blocks until the cluster instance is ready.
func (c *Cluster) Start(ctx context.Context) error {
	return c.start(ctx, false)
}

// Join the cluster as a new member. Blocks until the cluster instance is ready.
func (c *Cluster) Join(ctx context.Context) error {
	return c.start(ctx, true)
}

func (c *Cluster) start(ctx context.Context, join bool) error {
	// Create node host config
	nhc := config.NodeHostConfig{
		WALDir:         c.config.DataDir,
		NodeHostDir:    c.config.DataDir,
		RTTMillisecond: uint64(c.config.RTT.Milliseconds()),
		RaftAddress:    c.config.RaftAddress,
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
	}

	// Wait for all shards to be ready
	for shardID := range c.shards {
		if err := c.waitReady(ctx, shardID); err != nil {
			return fmt.Errorf("failed to wait for shard %d to be ready on replica %d: %w", shardID, c.config.ReplicaID, err)
		}
	}

	return nil
}

// Stop the node host and all shards.
func (c *Cluster) Stop() error {
	if c.nh == nil {
		return nil
	}

	c.nh.Close()
	c.nh = nil

	return nil
}

// AddMember to the cluster. This needs to be called on an existing running cluster member,
// before the new member is started.
func (c *Cluster) AddMember(ctx context.Context, shardID uint64, replicaID uint64, address string) error {
	logger := log.FromContext(ctx).Scope("raft")
	logger.Infof("adding member %s to shard %d on replica %d", address, shardID, replicaID)

	if err := c.withRetry(ctx, shardID, replicaID, func() error {
		return c.nh.SyncRequestAddReplica(ctx, shardID, replicaID, address, 0)
	}); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

// RemoveMember from the cluster. This needs to be called on an existing running cluster member,
// before the member is stopped.
func (c *Cluster) RemoveMember(ctx context.Context, shardID uint64, replicaID uint64) error {
	logger := log.FromContext(ctx).Scope("raft")
	logger.Infof("removing member from shard %d on replica %d", shardID, replicaID)

	if err := c.withRetry(ctx, shardID, replicaID, func() error {
		return c.nh.SyncRequestDeleteReplica(ctx, shardID, replicaID, 0)
	}); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	// if we removed the leader, we wait for the shard to be ready again
	if err := c.waitReady(ctx, shardID); err != nil {
		return fmt.Errorf("failed to wait for shard %d to be ready on replica %d: %w", shardID, replicaID, err)
	}
	return nil
}

// withTimeout runs an async dragonboat call and blocks until it succeeds or the context is cancelled.
// the call is retried if the request is dropped, which can happen if the leader is not available.
func (c *Cluster) withRetry(ctx context.Context, shardID, replicaID uint64, f func() error) error {
	retry := backoff.Backoff{
		Min:    c.config.RTT,
		Max:    c.config.ShardReadyTimeout,
		Factor: 2,
		Jitter: true,
	}
	logger := log.FromContext(ctx).Scope("raft")

	for {
		err := f()
		if errors.Is(err, dragonboat.ErrShardNotReady) {
			logger.Debugf("shard not ready, retrying in %s", retry.Duration())
			time.Sleep(retry.Duration())
			if _, ok := <-ctx.Done(); ok {
				return fmt.Errorf("context cancelled")
			}
			continue
		}
		if err != nil {
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
