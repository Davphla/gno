# PR 5216: Handle conflicting votes gracefully in tryAddVote

## Context

In `tm2/pkg/bft/consensus/state.go`, the `tryAddVote` function contained a
`panic("not yet implemented")` inside the `*types.VoteConflictingVotesError`
branch. This meant that whenever two votes from the same validator arrived for
different blocks (i.e., double-signing), every node hit an unrecoverable panic
and permanently halted consensus — even though double-signing is an expected
(if malicious) network event that should be tolerated gracefully.

Additionally, the surrounding comment incorrectly referred to `cs.mempool` as
the destination for conflicting-vote evidence, and the commented-out code used
stale API calls (`GetPubKey()` and `bytes.Equal` on `[20]byte`).

## Decision

Replace the `panic` with proper error handling:

1. **Own validator conflict** — log an error and return: "Found conflicting vote
   from ourselves. Did you unsafe_reset a validator?" with full vote details.
2. **Peer validator conflict** — log a warning: "Found conflicting vote from
   validator (double-signing)" with peer and vote details, then return the error.
3. A TODO comment notes future work: submit evidence to an evidence pool for
   validator slashing once that API is available.

The surrounding comment was corrected from "add it to the cs.mempool" to "log
it as double-signing evidence."

## Alternatives Considered

- **Submit to evidence pool immediately**: The original commented-out code called
  `cs.evpool.AddEvidence(voteErr.DuplicateVoteEvidence)`. This was kept as a
  TODO because the evidence pool API is not yet stable in this codebase.
- **Keep the panic**: Unacceptable — a single malicious validator can halt all
  nodes permanently.

## Consequences

- Consensus is no longer halted by double-signing events.
- Double-signing is logged for operator visibility.
- The error is returned to the caller (`handleMsg`) but intentionally not acted
  upon there (the `if err != nil` block at the call site remains a deliberate
  no-op suppressed with `//nolint:staticcheck`, consistent with upstream
  Tendermint's approach).
- Future work: wire in evidence pool submission when the API is ready.
