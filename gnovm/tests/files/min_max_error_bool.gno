package main

func main() {
	println(min(true, false))
}

// Error:
// main/min_max_error_bool.gno:4:10-26: invalid argument: cannot use bool as ordered type in call to min

// TypeCheckError:
// main/min_max_error_bool.gno:4:14: invalid argument: true (untyped bool constant) cannot be ordered
