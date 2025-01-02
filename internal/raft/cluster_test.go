package raft_test

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/internal/local"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/raft"
	"golang.org/x/sync/errgroup"
)

type IntEvent int64

func (e *IntEvent) UnmarshalBinary(data []byte) error {
	*e = IntEvent(binary.BigEndian.Uint64(data))
	return nil
}

func (e IntEvent) MarshalBinary() ([]byte, error) {
	return binary.BigEndian.AppendUint64([]byte{}, uint64(e)), nil
}

type IntStateMachine struct {
	sum int64
}

var _ raft.StateMachine[int64, int64, IntEvent, *IntEvent] = &IntStateMachine{}

func (s *IntStateMachine) Update(event IntEvent) error {
	s.sum += int64(event)
	return nil
}

func (s *IntStateMachine) Lookup(key int64) (int64, error) { return s.sum, nil }
func (s *IntStateMachine) Recover(reader io.Reader) error  { return nil }
func (s *IntStateMachine) Save(writer io.Writer) error     { return nil }
func (s *IntStateMachine) Close() error                    { return nil }

func TestCluster(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Second))
	defer cancel()

	members, err := local.FreeTCPAddresses(2)
	assert.NoError(t, err)

	builder1 := testBuilder(t, members, 1, members[0].String())
	shard1_1 := raft.AddShard(ctx, builder1, 1, &IntStateMachine{})
	shard1_2 := raft.AddShard(ctx, builder1, 2, &IntStateMachine{})
	cluster1 := builder1.Build(ctx)

	builder2 := testBuilder(t, members, 2, members[1].String())
	shard2_1 := raft.AddShard(ctx, builder2, 1, &IntStateMachine{})
	shard2_2 := raft.AddShard(ctx, builder2, 2, &IntStateMachine{})
	cluster2 := builder2.Build(ctx)

	wg, wctx := errgroup.WithContext(ctx)
	wg.Go(func() error { return cluster1.Start(wctx) })
	wg.Go(func() error { return cluster2.Start(wctx) })
	assert.NoError(t, wg.Wait())
	defer cluster1.Stop()
	defer cluster2.Stop()

	assert.NoError(t, shard1_1.Propose(ctx, IntEvent(1)))
	assert.NoError(t, shard2_1.Propose(ctx, IntEvent(2)))

	assert.NoError(t, shard1_2.Propose(ctx, IntEvent(1)))
	assert.NoError(t, shard2_2.Propose(ctx, IntEvent(1)))

	assertShardValue(ctx, t, 3, shard1_1, shard2_1)
	assertShardValue(ctx, t, 2, shard1_2, shard2_2)
}

func TestJoiningExistingCluster(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Second))
	defer cancel()

	members, err := local.FreeTCPAddresses(4)
	assert.NoError(t, err)

	builder1 := testBuilder(t, members[:2], 1, members[0].String())
	shard1 := raft.AddShard(ctx, builder1, 1, &IntStateMachine{})
	cluster1 := builder1.Build(ctx)

	builder2 := testBuilder(t, members[:2], 2, members[1].String())
	shard2 := raft.AddShard(ctx, builder2, 1, &IntStateMachine{})
	cluster2 := builder2.Build(ctx)

	wg, wctx := errgroup.WithContext(ctx)
	wg.Go(func() error { return cluster1.Start(wctx) })
	wg.Go(func() error { return cluster2.Start(wctx) })
	assert.NoError(t, wg.Wait())
	defer cluster1.Stop()
	defer cluster2.Stop()

	t.Log("join to the existing cluster as a new member")
	builder3 := testBuilder(t, nil, 3, members[2].String())
	shard3 := raft.AddShard(ctx, builder3, 1, &IntStateMachine{})
	cluster3 := builder3.Build(ctx)

	assert.NoError(t, cluster1.AddMember(ctx, 1, 3, members[2].String()))

	assert.NoError(t, cluster3.Join(ctx))
	defer cluster3.Stop()

	assert.NoError(t, shard3.Propose(ctx, IntEvent(1)))

	assertShardValue(ctx, t, 1, shard1, shard2, shard3)

	t.Log("join through the new member")
	builder4 := testBuilder(t, nil, 4, members[3].String())
	shard4 := raft.AddShard(ctx, builder4, 1, &IntStateMachine{})
	cluster4 := builder4.Build(ctx)

	assert.NoError(t, cluster3.AddMember(ctx, 1, 4, members[3].String()))
	assert.NoError(t, cluster4.Join(ctx))
	defer cluster4.Stop()

	assert.NoError(t, shard4.Propose(ctx, IntEvent(1)))

	assertShardValue(ctx, t, 2, shard1, shard2, shard3, shard4)
}

func TestLeavingCluster(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Second))
	defer cancel()

	members, err := local.FreeTCPAddresses(3)
	assert.NoError(t, err)

	builder1 := testBuilder(t, members, 1, members[0].String())
	shard1 := raft.AddShard(ctx, builder1, 1, &IntStateMachine{})
	cluster1 := builder1.Build(ctx)
	builder2 := testBuilder(t, members, 2, members[1].String())
	shard2 := raft.AddShard(ctx, builder2, 1, &IntStateMachine{})
	cluster2 := builder2.Build(ctx)
	builder3 := testBuilder(t, members, 3, members[2].String())
	shard3 := raft.AddShard(ctx, builder3, 1, &IntStateMachine{})
	cluster3 := builder3.Build(ctx)

	wg, wctx := errgroup.WithContext(ctx)
	wg.Go(func() error { return cluster1.Start(wctx) })
	wg.Go(func() error { return cluster2.Start(wctx) })
	wg.Go(func() error { return cluster3.Start(wctx) })
	assert.NoError(t, wg.Wait())
	defer cluster1.Stop() //nolint:errcheck
	defer cluster2.Stop() //nolint:errcheck
	defer cluster3.Stop() //nolint:errcheck

	t.Log("proposing event")
	assert.NoError(t, shard1.Propose(ctx, IntEvent(1)))
	assertShardValue(ctx, t, 1, shard1, shard2, shard3)

	t.Log("removing member")
	assert.NoError(t, cluster2.RemoveMember(ctx, 1, 1))
	assert.NoError(t, cluster1.Stop())

	t.Log("proposing event after removal")
	assert.NoError(t, shard2.Propose(ctx, IntEvent(1)))
	assertShardValue(ctx, t, 2, shard3, shard2)
}

func testBuilder(t *testing.T, addresses []*net.TCPAddr, id uint64, address string) *raft.Builder {
	members := make([]string, len(addresses))
	for i, member := range addresses {
		members[i] = member.String()
	}

	return raft.NewBuilder(&raft.RaftConfig{
		ReplicaID:          id,
		RaftAddress:        address,
		DataDir:            t.TempDir(),
		InitialMembers:     members,
		HeartbeatRTT:       1,
		ElectionRTT:        10,
		SnapshotEntries:    10,
		CompactionOverhead: 10,
		RTT:                10 * time.Millisecond,
		ShardReadyTimeout:  5 * time.Second,
	})
}

func assertShardValue(ctx context.Context, t *testing.T, expected int64, shards ...*raft.ShardHandle[IntEvent, int64, int64]) {
	t.Helper()

	for _, shard := range shards {
		res, err := shard.Query(ctx, 0)
		assert.NoError(t, err)
		assert.Equal(t, res, expected)
	}
}
