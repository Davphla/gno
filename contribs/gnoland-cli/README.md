# `gnoland-cli`

> Working name — to be renamed before merge.

`gnoland-cli` is a small helper that composes [`gnokey maketx`](../../gno.land/cmd/gnokey)
commands from a mix of flags and interactive prompts, prints the resulting bash
one-liner to stdout, and (only when explicitly asked) executes it.

It does **not** modify `gnokey`. It does **not** import any `gnokey` internals.
It generates bash; with `--run`, it runs that bash via `bash -c` after an
interactive confirmation.

## Usage

```
gnoland-cli maketx send -to <addr> -send 1ugnot \
  -gas-wanted 1000000 -gas-fee 1ugnot \
  -remote rpc.gno.land:443 -chainid gnoland1 mykey
```

Prints:

```
gnokey maketx send -to <addr> -send 1ugnot -gas-wanted 1000000 -gas-fee 1ugnot -remote rpc.gno.land:443 -chainid gnoland1 mykey
```

When a required flag is missing and stdin is a TTY, the tool prompts for it
interactively. Pass `--no-interactive` to disable prompts and fail fast on any
missing required field.

### Subcommands

- `gnoland-cli maketx send`     — emit a `gnokey maketx send ...` command
- `gnoland-cli maketx call`     — emit a `gnokey maketx call ...` command
- `gnoland-cli maketx addpkg`   — emit a `gnokey maketx addpkg ...` command
- `gnoland-cli maketx run`      — emit a `gnokey maketx run ...` command

### Executing the generated command

Default behavior is print-only. To execute, pass `--run`:

```
gnoland-cli maketx send ... --run mykey
```

The tool prints the bash command, prints a `/!\` warning, and asks for
explicit y/N confirmation before invoking `bash -c "<printed command>"`.

`--run --no-interactive` is rejected: unattended execution must be done
explicitly by piping the printed output through `bash`.

## Security notes

- No execution by default.
- The printed command and the executed command are the **same string**.
- All user-supplied values are POSIX shell-quoted on emission.
- No network access. No keybase reads. No RPC calls.
- The tool does not look up `gnokey` on `PATH`; the generated command runs
  through the user's shell, which respects aliases, wrappers, and `PATH`.
