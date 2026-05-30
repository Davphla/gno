package main

// Constant folding: these must be usable in const decls and as array sizes.
const (
	a = min(1, 2, 3)
	b = max(1, 2, 3)
	c = min("apple", "banana")
	d = max("apple", "banana")
	e = min(1.5, 2.5)
	f = max(1.5, 2.5)
	g = min(7) // single arg
)

// Array-size context proves it's truly a const expr.
var arr [max(3, 5, 1)]int

func main() {
	println(a)
	println(b)
	println(c)
	println(d)
	println(e)
	println(f)
	println(g)
	println(len(arr))
}

// Output:
// 1
// 3
// apple
// banana
// 1.5
// 2.5
// 7
// 5
