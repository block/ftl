package iterops_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/tuple"
	"github.com/block/ftl/internal/iterops"
)

func TestWindowPair(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4})
	result := slices.Collect(iterops.WindowPair(input))
	assert.Equal(t, result, []tuple.Pair[int, int]{
		tuple.PairOf(1, 2),
		tuple.PairOf(2, 3),
		tuple.PairOf(3, 4),
	})
}

func TestMap(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4})
	result := slices.Collect(iterops.Map(input, func(v int) int { return v * 2 }))
	assert.Equal(t, result, []int{2, 4, 6, 8})
}

func TestFlatMap(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4})
	result := slices.Collect(iterops.FlatMap(input, func(v int) iter.Seq[int] { return iterops.Const(v, v*2) }))
	assert.Equal(t, result, []int{1, 2, 2, 4, 3, 6, 4, 8})
}

func TestConcat(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4})
	result := slices.Collect(iterops.Concat(input, input))
	assert.Equal(t, result, []int{1, 2, 3, 4, 1, 2, 3, 4})

	result = slices.Collect(iterops.Concat(
		iterops.Const(1),
		iterops.Const(2),
		iterops.Const(3),
		iterops.Const(4),
	))
	assert.Equal(t, result, []int{1, 2, 3, 4})
}

func TestConst(t *testing.T) {
	input := 1
	result := slices.Collect(iterops.Const(input))
	assert.Equal(t, result, []int{1})
}

func TestEmpty(t *testing.T) {
	result := slices.Collect(iterops.Empty[int]())
	assert.Equal(t, result, nil)

	assert.Equal(t, slices.Collect(iterops.Concat(
		iterops.Empty[int](),
		iterops.Empty[int](),
	)), nil)

	assert.Equal(t, slices.Collect(iterops.Concat(
		iterops.Const(1),
		iterops.Empty[int](),
	)), []int{1})

	assert.Equal(t, slices.Collect(iterops.Concat(
		iterops.Empty[int](),
		iterops.Const(1),
	)), []int{1})
}

func TestDedup(t *testing.T) {
	input := slices.Values([]int{1, 2, 2, 3, 3, 4, 1})
	result := slices.Collect(iterops.Dedup(input))
	assert.Equal(t, result, []int{1, 2, 3, 4, 1})
}

func TestContains(t *testing.T) {
	input := slices.Values([]int{1, 2, 2, 3, 3, 4, 1})
	assert.True(t, iterops.Contains(input, 2))
	assert.False(t, iterops.Contains(input, 5))
}
