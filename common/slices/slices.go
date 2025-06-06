package slices

import (
	"cmp"
	"iter"
	"sort"

	"github.com/alecthomas/errors"
)

func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func MapErr[T, U any](slice []T, fn func(T) (U, error)) ([]U, error) {
	result := make([]U, len(slice))
	for i, v := range slice {
		var err error
		result[i], err = fn(v)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return result, nil
}

func Filter[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// GroupBy groups the elements of a slice by the result of a function.
func GroupBy[T any, K comparable](slice []T, fn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}

func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// AppendOrReplace appends a value to a slice if the slice does not contain a
// value for which the given function returns true. If the slice does contain
// such a value, it is replaced.
func AppendOrReplace[T any](slice []T, value T, fn func(T) bool) []T {
	for i, v := range slice {
		if fn(v) {
			slice[i] = value
			return slice
		}
	}
	return append(slice, value)
}

// Sort returns a sorted clone of slice.
func Sort[T cmp.Ordered](slice []T) []T {
	out := make([]T, len(slice))
	copy(out, slice)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out
}

// Reverse reverses a slice of any type in place.
func Reverse[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func FlatMap[T, U any](slice []T, fn func(T) []U) []U {
	result := make([]U, 0, len(slice))
	for _, v := range slice {
		result = append(result, fn(v)...)
	}
	return result
}

func Find[T any](slice []T, fn func(T) bool) (T, bool) {
	for _, v := range slice {
		if fn(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindVariant finds the first element in a slice that can be cast to the given type.
func FindVariant[T any, U any](slice []U) (T, bool) {
	for _, el := range slice {
		if found, ok := any(el).(T); ok {
			return found, true
		}
	}
	var zero T
	return zero, false
}

// FilterVariants finds all elements in a slice that can be cast to the given type.
func FilterVariants[T any, U any](slice []U) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, el := range slice {
			if found, ok := any(el).(T); ok {
				if !yield(found) {
					return
				}
			}
		}
	}
}

// Unique returns a new slice containing only the unique elements of the input, with the order preserved.
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// RemoveAll removes all elements in a slice that are in another slice.
func RemoveAll[T comparable](slice []T, values []T) []T {
	seen := make(map[T]struct{})
	for _, v := range values {
		seen[v] = struct{}{}
	}

	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}
