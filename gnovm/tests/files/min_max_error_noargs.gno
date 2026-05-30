package main

func main() {
	println(min())
}

// Error:
// main/min_max_error_noargs.gno:4:10-15: missing argument to min

// TypeCheckError:
// main/min_max_error_noargs.gno:4:14: invalid operation: not enough arguments for min() (expected 1, found 0)
