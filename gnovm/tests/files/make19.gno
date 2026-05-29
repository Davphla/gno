package main

// Oversized make(...) must produce a Gno-recoverable panic on allocation-size
// overflow, not an unrecoverable Go-host panic. Each case exercises a distinct
// allocator entry point: AllocateDataArray ([]byte), AllocateListArray
// ([]struct{} / []int), and AllocateMap (map).

func shouldPanic(f func()) {
	defer func() {
		r := recover()
		if r == nil {
			panic("not panicking")
		}
		println("recovered:", r)
	}()
	f()
}

func main() {
	const maxInt = int(^uint(0) >> 1)
	shouldPanic(func() { _ = make([]byte, maxInt) })
	shouldPanic(func() { _ = make([]int, maxInt) })
	shouldPanic(func() { _ = make([]struct{}, maxInt) })
	shouldPanic(func() { _ = make(map[int]int, maxInt) })
	println("ok")
}

// Output:
// recovered: runtime error: makeslice: len out of range
// recovered: runtime error: makeslice: len out of range
// recovered: runtime error: makeslice: len out of range
// recovered: runtime error: makeslice: len out of range
// ok
