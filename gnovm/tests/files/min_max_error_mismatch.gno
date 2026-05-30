package main

func main() {
	var i int = 1
	var s string = "a"
	println(min(i, s))
}

// Error:
// main/min_max_error_mismatch.gno:6:10-19: invalid argument: mismatched types int and string in call to min

// TypeCheckError:
// main/min_max_error_mismatch.gno:6:17: invalid argument: mismatched types int (previous argument) and string (type of s)
