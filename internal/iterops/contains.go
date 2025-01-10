package iterops

import "iter"

// Contains returns true if the sequence contains the value.
//
// This consumes the iterator to the first occurrence of the value.
func Contains[T comparable](seq iter.Seq[T], value T) bool {
	for v := range seq {
		if v == value {
			return true
		}
	}
	return false
}
