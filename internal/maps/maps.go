package maps

func FromSlice[K comparable, V any, T any](slice []T, kv func(el T) (K, V)) map[K]V {
	out := make(map[K]V, len(slice))
	for _, el := range slice {
		k, v := kv(el)
		out[k] = v
	}
	return out
}

// MapValues transforms a map[X]Y to a map[X]Z using the provided transformation function.
func MapValues[X comparable, Y any, Z any](input map[X]Y, transform func(X, Y) Z) map[X]Z {
	output := make(map[X]Z, len(input))
	for k, v := range input {
		output[k] = transform(k, v)
	}
	return output
}

// InsertMapMap inserts a value into a map[K1]map[K2]V.
// If the key1 does not exist, it is created as a new map[K2]V.
func InsertMapMap[K1 comparable, K2 comparable, V any](m map[K1]map[K2]V, key1 K1, key2 K2, value V) {
	if _, ok := m[key1]; !ok {
		m[key1] = map[K2]V{}
	}
	m[key1][key2] = value
}
