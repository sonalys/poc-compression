package gompressor

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](values []T) T {
	var max T
	for i := range values {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}
