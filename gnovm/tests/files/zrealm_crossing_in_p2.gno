// PKGPATH: gno.land/r/demo/crossinpalloc
package crossinpalloc

// An object a non-realm crossing frame allocates is stamped to that
// package's realm, which is never persisted. Storing it in a realm
// global is where that shows.

import "filetests/extern/crossinp"

var kept *crossinp.Box

func main(cur realm) {
	kept = crossinp.New(cross(cur), 7)
	println(kept.N)
}

// Error:
// unexpected unreal object: type=*gnolang.HeapItemValue oid=c7a7bcc489ecca7b4965d5158f8b288b4e2ce434:0 isNewReal=false isDirty=false isNewDeleted=false refCount=0
