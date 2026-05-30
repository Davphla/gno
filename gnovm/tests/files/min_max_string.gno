package main

func main() {
	// Lexicographic ordering.
	println(min("banana", "apple"))
	println(max("banana", "apple"))

	// Three args, includes substring relation.
	println(min("foo", "foobar", "fo"))
	println(max("foo", "foobar", "fo"))

	// Empty string is the smallest.
	println(min("", "x", "a"))
	println(max("", "x", "a"))

	// Typed string variables.
	a := "alpha"
	b := "beta"
	c := "gamma"
	println(min(a, b, c))
	println(max(a, b, c))

	// Single arg.
	println(min("only"))
	println(max("only"))
}

// Output:
// apple
// banana
// fo
// foobar
//
// x
// alpha
// gamma
// only
// only
