package main

import "math"

func main() {
	// Basic float64.
	println(min(1.5, 2.5))
	println(max(1.5, 2.5))

	// Multi-arg float64.
	println(min(3.14, 1.41, 2.71, 0.57))
	println(max(3.14, 1.41, 2.71, 0.57))

	// Typed float32.
	var a float32 = 1.5
	var b float32 = -0.5
	var c float32 = 2.25
	println(min(a, b, c))
	println(max(a, b, c))

	// NaN propagation: any NaN arg → NaN result.
	nan := math.NaN()
	println(math.IsNaN(min(nan, 1.0)))
	println(math.IsNaN(min(1.0, nan)))
	println(math.IsNaN(max(nan, 1.0)))
	println(math.IsNaN(max(1.0, nan)))
	println(math.IsNaN(min(1.0, 2.0, nan, 3.0)))
	println(math.IsNaN(max(1.0, 2.0, nan, 3.0)))

	// Infinities order correctly.
	inf := math.Inf(1)
	ninf := math.Inf(-1)
	println(min(inf, ninf, 0.0))
	println(max(inf, ninf, 0.0))
}

// Output:
// 1.5
// 2.5
// 0.57
// 3.14
// -0.5
// 2.25
// true
// true
// true
// true
// true
// true
// -Inf
// +Inf
