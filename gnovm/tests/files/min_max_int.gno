package main

func main() {
	// Basic min/max with 2 args.
	println(min(1, 2))
	println(max(1, 2))

	// More than 2 args.
	println(min(5, 1, 3, 2, 4))
	println(max(5, 1, 3, 2, 4))

	// Single arg.
	println(min(42))
	println(max(42))

	// Negative numbers.
	println(min(-5, -10, 0))
	println(max(-5, -10, 0))

	// Typed integer kinds.
	var i8 int8 = -3
	var j8 int8 = 7
	println(min(i8, j8))
	println(max(i8, j8))

	var u32 uint32 = 100
	var v32 uint32 = 200
	println(min(u32, v32))
	println(max(u32, v32))

	var i64 int64 = 1 << 40
	var j64 int64 = 1 << 30
	println(min(i64, j64))
	println(max(i64, j64))
}

// Output:
// 1
// 2
// 1
// 5
// 42
// 42
// -10
// 0
// -3
// 7
// 100
// 200
// 1073741824
// 1099511627776
