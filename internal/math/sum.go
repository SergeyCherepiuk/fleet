package math

import "golang.org/x/exp/constraints"

func Sum[T constraints.Integer | constraints.Float](slice []T) (sum T) {
	for _, value := range slice {
		sum += value
	}
	return
}
