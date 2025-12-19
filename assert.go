package main

// assert panics if the condition is false
func assert(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}

// assertMsg panics with a custom message if the condition is false
func assertMsg(condition bool, message string) {
	if !condition {
		panic("assertion failed: " + message)
	}
}
