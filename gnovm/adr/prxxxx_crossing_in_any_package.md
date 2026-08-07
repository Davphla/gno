# Any package may declare a crossing function

## Context

[gnolang/gno#5786](https://github.com/gnolang/gno/issues/5786): a `/p/`
package cannot declare a helper whose only, or first, parameter is a
`realm`, because a leading `realm` makes the function crossing and
crossing was allowed only in a realm. The workaround is a discarded
leading parameter:

```gno
func inviteMembers(_ int, rlm realm, boardID boards.ID, invites ...Invite)
inviteMembers(0, cur, boardID, invites...)
```

The restriction was written for a design where packages never crossed.
Under interrealm v2 they do, so the rule now refuses a signature that
has a meaning.

## Decision

Delete the restriction. Every package may declare a crossing function.

It lived in three places, with the two carve-outs in a fourth:

| Deleted | Was |
|---|---|
| `crossingAllowed` and its two call sites | refused a crossing declaration outside `/r/` |
| `isRealm` | its only caller |
| the `!m.Package.IsRealm()` gate in `doOpEnterCrossing` | the runtime half of the same rule |
| `crossingFromTestFile` | existed only to punch a `*_test.gno` hole in that gate |

A non-realm package has a realm to cross into.
[`fillPackage`](https://github.com/gnolang/gno/blob/abcd20dad/gnovm/pkg/gnolang/store.go#L599-L611)
gives every immutable library path one derived from its pkgpath, so
`pv.GetRealm()` on the cross path is non-nil.

## Alternatives considered

**Move the realm last.**
[#6033](https://github.com/gnolang/gno/pull/6033). Clears the placeholder
from 116 signatures and leaves 68 with nowhere to put a trailing realm.

**Read the leading parameter's name**, `cur` crossing and `rlm` not, with
`FuncType.TypeID` carrying the distinction so the two shapes are
different types. Larger, and it makes a parameter name change a type,
which Go does nowhere. It does give the callee the caller's realm, which
this does not.

## Consequences

Crossing hands the callee its own package's realm, not the caller's, so
this is not a replacement for the helpers that take a realm in order to
write the caller's storage. Two filetests pin that down.

`zrealm_crossing_in_p1.gno`, writing a caller-owned field:

```
cannot directly modify readonly tainted object (use a method or crossing function): b.N
```

`zrealm_crossing_in_p2.gno`, returning an object the frame allocated:

```
unexpected unreal object: type=*gnolang.HeapItemValue oid=c7a7bcc4... isNewReal=false isDirty=false
```

The second is raised by
[`toRefValue`](https://github.com/gnolang/gno/blob/abcd20dad/gnovm/pkg/gnolang/realm.go#L2059-L2061)
during finalization: a `/p/` realm is never persisted, so an object
allocated inside that frame stays unreal while a realm global points at
it. It is an interpreter panic, not a language error, and it is the open
question on this change. Closing it means re-stamping escaping
allocations to the caller's realm, or giving `/p/` a persisted realm.

Nothing in `./gnovm/...` or `./gno.land/...` failed when the rule was
deleted, so the rule had no runtime consequence any existing test
observed. Both consequences above are new.
