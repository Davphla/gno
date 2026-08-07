// PKGPATH: gno.land/r/demo/crossinpdemo
package crossinpdemo

// gnolang/gno#5786: a package that is not a realm may now declare a
// function whose only parameter is a realm, and a realm may cross into
// it.

import "filetests/extern/crossinp"

func main(cur realm) {
	println(crossinp.Sync(cross(cur)))
}

// Output:
// filetests/extern/crossinp
