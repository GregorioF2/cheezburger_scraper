package math

func MaxInt(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func MinInt(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

func GCD(a, b int) int {
	l := MaxInt(a, b)
	r := MinInt(a, b)
	for r != 0 {
		tmp := r
		r = l % r
		l = tmp
	}
	return l
}
