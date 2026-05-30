package main

// Untyped constants converge to the typed arg's type, just like binary < / >.
func main() {
	var x int = 10
	println(min(1, x, 5))     // untyped 1, 5 convert to int
	println(max(1, x, 5))

	var y float64 = 2.5
	println(min(1, y, 4))     // untyped int -> float64
	println(max(1, y, 4))

	var s string = "middle"
	println(min("apple", s, "zebra"))
	println(max("apple", s, "zebra"))

	// All-untyped: defaults apply (untyped int -> int, untyped float -> float64).
	x2 := min(1, 2, 3)
	println(x2) // int 1

	y2 := min(1.5, 2)
	println(y2) // float64 1.5 (mixed untyped int+float → float)
}

// Output:
// 1
// 10
// 1
// 4
// apple
// zebra
// 1
// 1.5
