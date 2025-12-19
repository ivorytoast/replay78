package assert

func Is(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}
