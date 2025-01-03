package raft_test

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/internal/local"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/raft"
	"golang.org/x/sync/errgroup"
)

type IntStreamEvent struct {
	Value int
}

func (e IntStreamEvent) Handle(view IntSumView) (IntSumView, error) {
	return IntSumView{Sum: view.Sum + e.Value}, nil
}

func (e IntStreamEvent) MarshalBinary() ([]byte, error) {
	return binary.BigEndian.AppendUint64([]byte{}, uint64(e.Value)), nil
}

func (e *IntStreamEvent) UnmarshalBinary(data []byte) error {
	e.Value = int(binary.BigEndian.Uint64(data))
	return nil
}

type IntSumView struct {
	Sum int
}

func (v IntSumView) MarshalBinary() ([]byte, error) {
	return binary.BigEndian.AppendUint64([]byte{}, uint64(v.Sum)), nil
}

func (v *IntSumView) UnmarshalBinary(data []byte) error {
	v.Sum = int(binary.BigEndian.Uint64(data))
	return nil
}

func TestEventView(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(60*time.Second))
	t.Cleanup(cancel)

	members, err := local.FreeTCPAddresses(2)
	assert.NoError(t, err)

	builder1 := testBuilder(t, members, 1, members[0].String(), nil)
	view1 := raft.AddEventView[IntSumView, *IntSumView, IntStreamEvent](ctx, builder1, 1)
	cluster1 := builder1.Build(ctx)

	builder2 := testBuilder(t, members, 2, members[1].String(), nil)
	view2 := raft.AddEventView[IntSumView, *IntSumView, IntStreamEvent](ctx, builder2, 1)
	cluster2 := builder2.Build(ctx)

	eg, wctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return cluster1.Start(wctx) })
	eg.Go(func() error { return cluster2.Start(wctx) })
	assert.NoError(t, eg.Wait())
	t.Cleanup(func() {
		cluster1.Stop(ctx)
		cluster2.Stop(ctx)
	})

	assert.NoError(t, view1.Publish(ctx, IntStreamEvent{Value: 1}))

	view, err := view1.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, IntSumView{Sum: 1}, view)

	view, err = view2.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, IntSumView{Sum: 1}, view)
}
