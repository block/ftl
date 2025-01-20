package raft_test

import (
	"encoding/binary"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/raft"
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
	if testing.Short() {
		t.SkipNow()
	}
	ctx := testContext(t)

	_, views := startClusters(ctx, t, 2, func(b *raft.Builder) eventstream.EventView[IntSumView, IntStreamEvent] {
		return raft.AddEventView[IntSumView, *IntSumView, IntStreamEvent](ctx, b, 1)
	})

	assert.NoError(t, views[0].Publish(ctx, IntStreamEvent{Value: 1}))

	view, err := views[0].View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, IntSumView{Sum: 1}, view)

	view, err = views[1].View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, IntSumView{Sum: 1}, view)
}
