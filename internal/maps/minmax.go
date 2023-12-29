package maps

import "golang.org/x/exp/constraints"

func KeyWithMinValue[K comparable, V constraints.Ordered](m map[K]V) K {
	var (
		key   K
		value V
	)

	i := 0
	for k, v := range m {
		if i == 0 || v < value {
			key, value = k, v
		}
		i++
	}

	return key
}

func KeyWithMaxValue[K comparable, V constraints.Ordered](m map[K]V) K {
	var (
		key   K
		value V
	)

	i := 0
	for k, v := range m {
		if i == 0 || v > value {
			key, value = k, v
		}
		i++
	}

	return key
}
