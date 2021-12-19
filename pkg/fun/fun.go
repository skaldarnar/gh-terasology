package fun

// requires go@v1.18 or later for generics support
func GroupBy[T any](xs []T, f func(T) string) map[string][]T {
	res := make(map[string][]T)
	for _, x := range xs {
		res[f(x)] = append(res[f(x)], x)
	}
	return res
}

func Map[T any, U any](xs []T, f func(T) U) []U {
	ys := make([]U, 0, len(xs))
	for i, x := range xs {
		ys[i] = f(x)
	}
	return ys
}

// in-place variant of `Map`.
func MapI[T any](xs []T, f func(T) T) []T {
	for i, x := range xs {
		xs[i] = f(x)
	}
	return xs
}

// get the keys of the given map `m` as a slice
func Keys[T comparable, S any](m map[T]S) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
