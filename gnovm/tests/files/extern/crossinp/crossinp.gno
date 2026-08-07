package crossinp

// A crossing function declared outside a realm. Before, this was a
// compile error and the workaround was a discarded leading parameter:
// func Sync(_ int, rlm realm) string.

func Sync(cur realm) string {
	return cur.PkgPath()
}

type Box struct {
	N int
}

// Mutate writes through a pointer the caller owns. As a crossing
// function this runs with the package's own realm, not the caller's.
func Mutate(cur realm, b *Box) {
	b.N = 42
}

// New allocates inside the package's own realm frame and hands the
// object back to the caller.
func New(cur realm, n int) *Box {
	return &Box{N: n}
}
