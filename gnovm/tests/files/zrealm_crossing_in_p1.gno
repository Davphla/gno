// PKGPATH: gno.land/r/demo/crossinpwrite
package crossinpwrite

// A crossing function in a non-realm package runs with that package's
// own realm, not the caller's, so it cannot write what the caller owns.
// This is the limit of allowing crossing outside /r/: it does not
// replace the helpers that take a realm in order to act on the caller's
// storage.

import "filetests/extern/crossinp"

var b = &crossinp.Box{N: 1}

func main(cur realm) {
	crossinp.Mutate(cross(cur), b)
	println(b.N)
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): b<VPBlock(1,1)>.N
