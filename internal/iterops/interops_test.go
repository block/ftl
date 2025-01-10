package iterops_test

import (
	"slices"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/tuple"
	"github.com/block/ftl/internal/iterops"
)

func TestWindowPair(t *testing.T) {
	input := slices.Values([]int{1, 2, 3, 4})
	result := slices.Collect(iterops.WindowPair(input, 0))
	assert.Equal(t, result, []tuple.Pair[int, int]{
		tuple.PairOf(0, 1),
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
	result := slices.Collect(iterops.FlatMap(input, func(v int) []int { return []int{v, v * 2} }))
	assert.Equal(t, result, []int{1, 2, 2, 4, 3, 6, 4, 8})
}
