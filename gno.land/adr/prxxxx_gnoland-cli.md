# ADR: `gnoland-cli` — opt-in helper that emits `gnokey maketx` commands

> Filename uses `prxxxx_` until the PR number is known.

## Context

Several recent UX-improvement PRs against `gnokey` and `gno.land/pkg/` were
abandoned because they tied a helper experience to production CLI surface
and to libraries reviewers consider generic:

- #5594 (`maketx` interactive wizard) embedded prompts inside
  `gno.land/pkg/keyscli/maketx_wizard.go` and modified every subcommand
  exec path. It depended on a prompt library from #5557 that never merged.
- #5596 (network registry API) was rejected as "too much yolo" — hardcoding
  network names and gnoweb URLs in low-level libraries.
- #5543 and #5557 were closed/abandoned as the dependency chain unwound.

The remaining demand is real: `gnokey maketx <send|call|addpkg|run>`
takes 6–9 required flags. New users hit the cliff every time. We want a
small helper that takes the cliff away, without modifying any production
code path or library.

## Decision

A new tool at `contribs/gnoland-cli/` (working name; will be renamed
before any upstream PR). The tool **emits a bash one-liner** that calls
`gnokey`, prints it to stdout, and — only when the user opts in with
`--run` — executes that exact string via `bash -c` after a `/!\` warning
and a default-no y/N confirmation.

The same string serves as both the printed deliverable and the executed
payload. There is no separate argv-quoting path.

### Architecture

| Layer | Responsibility |
|-------|----------------|
| `contribs/gnoland-cli/main.go` + `root.go` | Wire `tm2/pkg/commands` root command with global `--run` / `--no-interactive` flags |
| `maketx.go` | Parent `maketx` command, lists the four subcommands |
| `maketx_{send,call,addpkg,run}.go` | One file per subcommand; each holds an embedded `SharedFlags`, registers subcommand-specific flags, and assembles the emitted command |
| `subcmd_common.go` | `resolvePositional` / `resolveString` / `resolveSharedGas`: prompt or error depending on TTY + `--no-interactive` |
| `internal/gen/builder.go` | Token builder: `StringFlag`, `Int64Flag`, `BoolFlag`, `StringFlagIfChanged`, `RepeatedStringFlag`, `Positional`, `String()`. Every emit goes through `shellquote.Join` |
| `internal/gen/shared_flags.go` | `SharedFlags`: gnokey base + maketx parent flags. `Register(fs)` registers them; `EmitOn(b)` writes them in canonical order, omitting defaults |
| `internal/gen/runner.go` | Print, warn, confirm-default-no, `exec.Command("bash", "-c", str)` |
| `internal/prompt/prompt.go` | Thin wrapper over `AlecAivazis/survey/v2`; single seam for `WithStdio` plumbing |
| `internal/validate/validate.go` | Syntactic validators: bech32, coins, gas, pkgpath, dir |

### Dependencies

- `github.com/AlecAivazis/survey/v2` — interactive prompts. Battle-tested
  (used by GitHub's `gh` CLI). MIT.
- `github.com/kballard/go-shellquote` — POSIX shell-quoting. Already an
  indirect dep of `gno.land/pkg/gnoweb/tools/`. BSD-2, unchanged since 2018.
- `golang.org/x/term` — TTY detection for `IsInteractive`.
- `tm2/pkg/commands` — same CLI framework as `gnokey`, `gnobro`, `gnogenesis`.

No new dependency on `gno.land/pkg/keyscli` or any other gnokey internals.

### Security properties

1. **No execution by default.** v1 prints; opt-in execution requires both
   `--run` and a typed `y`/`yes`.
2. **Same string in/out.** What the user sees in stdout is exactly what
   gets passed to `bash -c`. No second argv array.
3. **POSIX-safe quoting via go-shellquote.** Verified by a round-trip test
   asserting `Split(Join(tokens)) == tokens` across every ASCII metachar,
   command-substitution patterns, embedded quotes/newlines, unicode.
4. **Default-no confirmation.** `tm2/pkg/commands.IO.GetConfirmation`
   defaults to yes on Enter, which is unsafe for a destructive action.
   `runner.confirmDefaultNo` accepts only explicit `y`/`yes`.
5. **No network / no keybase / no FS access** beyond `os.Stat` in the
   `Dir` validator. The tool is a string builder.
6. **No `gnokey` lookup.** The shelled-out `bash -c` honors the user's
   `PATH`, aliases, and shell wrappers transparently.

## Alternatives Considered

### `exec.Command("gnokey", argv...)` (no bash)

Earlier draft. Rejected:

- Required an explicit `gnokey` binary lookup (`exec.LookPath` or a
  `--gnokey-bin` override), inviting the "wrong gnokey on PATH" footgun.
- Doesn't compose with user aliases or shell wrappers (e.g.
  `alias gnokey='sudo -u deploy gnokey'`).
- The printed command and the executed command diverged: printed was a
  shell line, executed was an argv slice. Two quoting paths to test
  rather than one.

### Wizard mode at the `maketx` parent (the #5594 shape)

The original wizard offered a "what kind of tx?" first prompt. v1
deliberately omits it: users type one of four subcommands directly. Less
surface area, less to mis-prompt. Can be added later as a separate
subcommand if the demand is clear.

### Vendor a local prompt library

Considered for ~150 LOC inline. Rejected in favor of `survey/v2` because:

- Survey already handles raw-mode TTY, Ctrl-C/EOF, terminal resize, and
  retry-on-invalid — all the edge cases hand-rolled prompts get wrong.
- Used by GitHub's `gh` CLI; the audit surface is borrowed from there.
- Wrapping it in `internal/prompt/` means a single file to swap if
  upstream `tm2/pkg/commands/prompt.go` ever lands.

### Persistent config (`~/.config/gnoland-cli/`)

Rejected for v1. Persistent state plus opt-in execution creates a "I
forgot I enabled auto-run" footgun. `--run` is per-invocation only.

### Bundled features that pulled #5594 / #5596 / #5543 down

Deliberately **not** in v1:

- Network registry / known-network list (`dev`/`staging`/`gnoland1`/`test11`).
- Gnoweb URL display or probing.
- Gas auto-estimation via simulation.
- Air-gap signing workflow hints.
- Template / saved-profile support (planned as a separate work item).

## Consequences

### Positive

- Zero impact on `gnokey`, `keyscli`, or any production library.
- Audit surface is small: builder + runner + ~four subcommand files.
- Round-trip-safe quoting tested exhaustively.
- Survives downstream changes to `gnokey` flags via the conventional
  `[flags] <key>` shape (every subcommand mirrors gnokey's flag names).

### Negative

- Duplication of gnokey flag declarations in `SharedFlags`. If gnokey
  adds a new shared flag (e.g. `-tx-version`), `SharedFlags` needs the
  same flag added. Manageable; flagged in the `SharedFlags` doc comment.
- Survey is a new top-level dependency for the gno repo. The `internal/prompt/`
  seam means it's swappable later.

### Follow-ups (separate PRs)

- Templates: `gnoland-cli template save deploy-foo ...`,
  `gnoland-cli template run deploy-foo --key mykey`.
- Additional subcommands: `query`, `sign`, `broadcast`.
- Replace `internal/prompt/` with an upstream `tm2/pkg/commands/prompt.go`
  if/when one lands.
- Final rename from the `gnoland-cli` placeholder.

## Verification

```
cd contribs/gnoland-cli
make install   # go install .
make test      # unit + integration tests
make lint      # golangci-lint, 0 issues
```

End-to-end with a `gnokey` stub binary on `PATH`:

```
$ gnoland-cli --run maketx send -to g1... -send 1ugnot \
    -gas-wanted 1 -gas-fee 1ugnot -remote x -chainid x mykey
gnokey maketx send -remote x -gas-wanted 1 -gas-fee 1ugnot -chainid x -to g1... -send 1ugnot mykey

/!\ This will execute the command above in your shell. Review it carefully.
Run this command? [y/N]: y
ARGV maketx
ARGV send
... # stub echoes every arg, byte-identical to the printed line
```
