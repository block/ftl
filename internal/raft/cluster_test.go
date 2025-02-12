package raft_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/raft/v1/raftpbconnect"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/local"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/raft"
	"github.com/block/ftl/internal/retry"
	"github.com/block/ftl/internal/rpc"
	sm "github.com/block/ftl/internal/statemachine"
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

var _ sm.Snapshotting[int64, int64, IntEvent] = &IntStateMachine{}

func (s *IntStateMachine) Publish(event IntEvent) error {
	if event == 12345 {
		// 12345 is not a valid number, so we return an error
		return raft.ErrInvalidEvent
	}

	s.sum += int64(event)
	return nil
}

func (s *IntStateMachine) Lookup(key int64) (int64, error) { return s.sum, nil }
func (s *IntStateMachine) Recover(reader io.Reader) error  { return nil }
func (s *IntStateMachine) Save(writer io.Writer) error     { return nil }
func (s *IntStateMachine) Close() error                    { return nil }

func TestClusterWith2Shards(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.SkipNow()
	}
	ctx := testContext(t)

	_, shards := startClusters(ctx, t, 2, func(b *raft.Builder) []sm.Handle[int64, int64, IntEvent] {
		return []sm.Handle[int64, int64, IntEvent]{
			raft.AddShard(ctx, b, 1, &IntStateMachine{}),
			raft.AddShard(ctx, b, 2, &IntStateMachine{}),
		}
	})

	assert.NoError(t, shards[0][0].Publish(ctx, IntEvent(1)))
	assert.NoError(t, shards[1][0].Publish(ctx, IntEvent(1)))

	assert.NoError(t, shards[0][1].Publish(ctx, IntEvent(1)))
	assert.NoError(t, shards[1][1].Publish(ctx, IntEvent(2)))

	assertShardValue(ctx, t, 2, shards[0][0], shards[1][0])
	assertShardValue(ctx, t, 3, shards[0][1], shards[1][1])
}

func TestJoiningExistingCluster(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.SkipNow()
	}
	ctx := testContext(t)

	addresses, err := local.FreeTCPAddresses(5)
	assert.NoError(t, err)
	members := addresses[:4]
	controlAddress := fmt.Sprintf("http://%s", addresses[4].String())
	controlBind, err := url.Parse(controlAddress)
	assert.NoError(t, err)

	builder1 := testBuilder(t, members[:2], members[0].String(), controlBind)
	shard1 := raft.AddShard(ctx, builder1, 1, &IntStateMachine{})
	cluster1 := builder1.Build(ctx)

	builder2 := testBuilder(t, members[:2], members[1].String(), nil)
	shard2 := raft.AddShard(ctx, builder2, 1, &IntStateMachine{})
	cluster2 := builder2.Build(ctx)

	wg, wctx := errgroup.WithContext(ctx)
	wg.Go(func() error { return cluster1.Start(wctx) })
	wg.Go(func() error { return cluster2.Start(wctx) })
	// set up the control endpoint to cluster 1
	go func() {
		assert.NoError(t, rpc.Serve(ctx, controlBind,
			rpc.GRPC(raftpbconnect.NewRaftServiceHandler, cluster1),
		))
	}()
	assert.NoError(t, wg.Wait())
	t.Cleanup(func() {
		cluster1.Stop(ctx)
		cluster2.Stop(ctx)
	})

	t.Log("join to the existing cluster as a new member")
	builder3 := testBuilder(t, nil, members[2].String(), nil)
	shard3 := raft.AddShard(ctx, builder3, 1, &IntStateMachine{})
	cluster3 := builder3.Build(ctx)

	assert.NoError(t, cluster3.Join(ctx, controlAddress))
	t.Cleanup(func() {
		cluster3.Stop(ctx)
	})

	assert.NoError(t, shard3.Publish(ctx, IntEvent(1)))

	assertShardValue(ctx, t, 1, shard1, shard2, shard3)

	t.Log("join through the new member")
	builder4 := testBuilder(t, nil, members[3].String(), nil)
	shard4 := raft.AddShard(ctx, builder4, 1, &IntStateMachine{})
	cluster4 := builder4.Build(ctx)

	assert.NoError(t, cluster4.Join(ctx, controlAddress))
	t.Cleanup(func() {
		cluster4.Stop(ctx)
	})

	assert.NoError(t, shard4.Publish(ctx, IntEvent(1)))

	assertShardValue(ctx, t, 2, shard1, shard2, shard3, shard4)
}

func TestLeavingCluster(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.SkipNow()
	}
	ctx := testContext(t)

	clusters, shards := startClusters(ctx, t, 3, func(b *raft.Builder) sm.Handle[int64, int64, IntEvent] {
		return raft.AddShard(ctx, b, 1, &IntStateMachine{})
	})

	t.Log("proposing event")
	assert.NoError(t, shards[0].Publish(ctx, IntEvent(1)))
	assertShardValue(ctx, t, 1, shards...)

	t.Log("removing member")
	clusters[0].Stop(ctx)

	t.Log("proposing event after a member has been stopped")
	assert.NoError(t, shards[1].Publish(ctx, IntEvent(1)))
	assertShardValue(ctx, t, 2, shards[1:]...)
}

func TestStateIter(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.SkipNow()
	}
	ctx := testContext(t)

	_, shards := startClusters(ctx, t, 2, func(b *raft.Builder) sm.Handle[int64, int64, IntEvent] {
		return raft.AddShard(ctx, b, 1, &IntStateMachine{})
	})

	changes, err := shards[0].StateIter(ctx, 0)
	assert.NoError(t, err)

	assert.NoError(t, shards[0].Publish(ctx, IntEvent(1)))
	assert.NoError(t, shards[1].Publish(ctx, IntEvent(1)))

	assert.True(t, iterops.Contains(changes, 2))
}

func TestInvalidEvents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.SkipNow()
	}
	ctx := testContext(t)

	_, shards := startClusters(ctx, t, 2, func(b *raft.Builder) []sm.Handle[int64, int64, IntEvent] {
		return []sm.Handle[int64, int64, IntEvent]{
			raft.AddShard(ctx, b, 1, &IntStateMachine{}),
		}
	})

	// 12345 is not actually a valid number, so we expect an error
	err := shards[0][0].Publish(ctx, IntEvent(12345))
	assert.IsError(t, err, raft.ErrInvalidEvent)

	// check that the event did not impact the state machine
	assertShardValue(ctx, t, 0, shards[0][0])
	// and we can still publish valid events
	assert.NoError(t, shards[0][0].Publish(ctx, IntEvent(1)))
	assertShardValue(ctx, t, 1, shards[0][0])
}

func testBuilder(t *testing.T, addresses []*net.TCPAddr, address string, _ *url.URL) *raft.Builder {
	members := make([]string, len(addresses))
	for i, member := range addresses {
		members[i] = member.String()
	}

	initialIDs := make([]uint64, len(addresses))
	for i := range addresses {
		initialIDs[i] = uint64(i + 1)
	}

	return raft.NewBuilder(&raft.RaftConfig{
		Address:            address,
		DataDir:            t.TempDir(),
		InitialMembers:     members,
		InitialReplicaIDs:  initialIDs,
		HeartbeatRTT:       1,
		ElectionRTT:        5,
		SnapshotEntries:    10,
		CompactionOverhead: 10,
		RTT:                50 * time.Millisecond,
		ShardReadyTimeout:  5 * time.Second,
		ChangesInterval:    5 * time.Millisecond,
		ChangesTimeout:     1 * time.Second,
		QueryTimeout:       5 * time.Second,
		Retry: retry.RetryConfig{
			Min:    50 * time.Millisecond,
			Max:    1 * time.Second,
			Factor: 2,
			Jitter: true,
		},
	})
}

func startClusters[T any](ctx context.Context, t *testing.T, count int, builderF func(b *raft.Builder) T) ([]*raft.Cluster, []T) {
	t.Helper()

	clusters := make([]*raft.Cluster, count)
	members, err := local.FreeTCPAddresses(count)
	assert.NoError(t, err)
	result := make([]T, count)

	for i := range count {
		builder := testBuilder(t, members, members[i].String(), nil)
		result[i] = builderF(builder)
		clusters[i] = builder.Build(ctx)
	}

	wg, wctx := errgroup.WithContext(ctx)
	for _, cluster := range clusters {
		wg.Go(func() error { return cluster.Start(wctx) })
	}
	assert.NoError(t, wg.Wait())

	t.Cleanup(func() {
		for _, cluster := range clusters {
			// timeout stop in 1 second, as closing the full cluster can lead
			// to a state where the last members can not update the membership shard
			// anymore
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			cluster.Stop(ctx)
			cancel()
		}
	})

	return clusters, result
}

func assertShardValue(ctx context.Context, t *testing.T, expected int64, shards ...sm.Handle[int64, int64, IntEvent]) {
	t.Helper()

	for _, shard := range shards {
		res, err := shard.Query(ctx, 0)
		assert.NoError(t, err)
		assert.Equal(t, res, expected)
	}
}

func testContext(t *testing.T) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	return ctx
}
