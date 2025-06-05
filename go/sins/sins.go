package sins

func Ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func TernaryF[T any](cond bool, a, b func() T) T {
	if cond {
		return a()
	}
	return b()
}
