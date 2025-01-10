package iterops

import (
	"iter"

	"github.com/alecthomas/types/tuple"
)

// ChangeExtractor extracts changes from an old and new state.
type ChangeExtractor[S, C any] func(tuple.Pair[S, S]) iter.Seq[C]

// Changes returns a stream of change events from a stream of evolving state.
func Changes[S, C any](in iter.Seq[S], extractor ChangeExtractor[S, C]) iter.Seq[C] {
	return FlatMap(WindowPair(in), extractor)
}
