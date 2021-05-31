package utils

func MinMax(a0 int, arr ...int) (int, int) {
	min := a0
	max := a0
	for _, a := range arr {
		if a > max {
			max = a
		}
		if a < min {
			min = a
		}
	}

	return min, max
}
