package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// run drives the full command tree like main(), with buffered IO and no
// TTY. Returns trimmed stdout, trimmed stderr, and the ParseAndRun error.
func run(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(stdin))
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io.SetOut(commands.WriteNopCloser(outBuf))
	io.SetErr(commands.WriteNopCloser(errBuf))

	err = newRootCmd(io).ParseAndRun(context.Background(), args)
	return strings.TrimRight(outBuf.String(), "\n"), errBuf.String(), err
}

// ---------------------------------------------------------------------
// send
// ---------------------------------------------------------------------

func TestFunctional_Send(t *testing.T) {
	const addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

	cases := []struct {
		name           string
		args           []string
		wantStdout     string
		wantErrSubstr  string
		stdoutContains []string
	}{
		{
			name: "happy minimal",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-send", "1ugnot",
				"-gas-wanted", "1000000", "-gas-fee", "1ugnot",
				"-remote", "rpc.gno.land:443", "-chainid", "gnoland1",
				"mykey",
			},
			wantStdout: "gnokey maketx send -remote rpc.gno.land:443 -gas-wanted 1000000 -gas-fee 1ugnot -chainid gnoland1 -to " + addr + " -send 1ugnot mykey",
		},
		{
			name: "with memo, quiet, home, insecure-password-stdin",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-send", "1ugnot",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"-remote", "x", "-chainid", "x",
				"-memo", "deposit for alice",
				"-home", "/tmp/gnokey-home",
				"-quiet",
				"-insecure-password-stdin",
				"mykey",
			},
			stdoutContains: []string{
				"-home /tmp/gnokey-home",
				"-quiet=true",
				"-insecure-password-stdin=true",
				"-memo 'deposit for alice'",
			},
		},
		{
			name: "broadcast=false emitted; defaults omitted",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-send", "1ugnot",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"-broadcast=false",
				"mykey",
			},
			stdoutContains: []string{"-broadcast=false"},
		},
		{
			name: "simulate=skip emitted; default chainid dev omitted",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-send", "1ugnot",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"-simulate", "skip",
				"mykey",
			},
			stdoutContains: []string{"-simulate skip"},
		},
		{
			name: "missing -to is rejected non-interactive",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-send", "1ugnot", "-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			wantErrSubstr: "to not specified",
		},
		{
			name: "missing -send is rejected non-interactive",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			wantErrSubstr: "send not specified",
		},
		{
			name: "missing gas-wanted is rejected",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-send", "1ugnot", "-gas-fee", "1ugnot",
				"mykey",
			},
			wantErrSubstr: "gas-wanted not specified",
		},
		{
			name: "missing gas-fee is rejected",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-send", "1ugnot", "-gas-wanted", "1",
				"mykey",
			},
			wantErrSubstr: "gas-fee not specified",
		},
		{
			name: "missing key positional is rejected",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-send", "1ugnot", "-gas-wanted", "1", "-gas-fee", "1ugnot",
			},
			wantErrSubstr: "Key name or address required",
		},
		{
			name: "too many positionals returns flag.ErrHelp",
			args: []string{
				"--no-interactive", "maketx", "send",
				"-to", addr, "-send", "1ugnot", "-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey", "extra",
			},
			wantErrSubstr: "help",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := run(t, "", tt.args...)
			assertOutcome(t, tt, stdout, err)
		})
	}
}

// ---------------------------------------------------------------------
// call
// ---------------------------------------------------------------------

func TestFunctional_Call(t *testing.T) {
	cases := []struct {
		name           string
		args           []string
		wantStdout     string
		wantErrSubstr  string
		stdoutContains []string
		stdoutAbsent   []string
	}{
		{
			name: "happy with all flags and chainid override",
			args: []string{
				"--no-interactive", "maketx", "call",
				"-pkgpath", "gno.land/r/demo/foo", "-func", "Bar",
				"-args", "1", "-args", "2",
				"-send", "5ugnot", "-max-deposit", "10ugnot",
				"-gas-wanted", "1000000", "-gas-fee", "1ugnot",
				"-remote", "rpc.gno.land:443", "-chainid", "gnoland1",
				"mykey",
			},
			wantStdout: "gnokey maketx call -remote rpc.gno.land:443 -gas-wanted 1000000 -gas-fee 1ugnot -chainid gnoland1 -pkgpath gno.land/r/demo/foo -func Bar -args 1 -args 2 -send 5ugnot -max-deposit 10ugnot mykey",
		},
		{
			name: "no args is fine (no-arg function call)",
			args: []string{
				"--no-interactive", "maketx", "call",
				"-pkgpath", "gno.land/r/demo/foo", "-func", "Nothing",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			stdoutAbsent: []string{"-args"},
		},
		{
			name: "defaults (-broadcast, -simulate, -chainid) omitted",
			args: []string{
				"--no-interactive", "maketx", "call",
				"-pkgpath", "gno.land/r/demo/foo", "-func", "Bar",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			stdoutAbsent: []string{"-broadcast", "-simulate", "-chainid"},
		},
		{
			name: "args with shell metachars are quoted safely",
			args: []string{
				"--no-interactive", "maketx", "call",
				"-pkgpath", "gno.land/r/demo/foo", "-func", "Greet",
				"-args", "hello world",
				"-args", "it's me",
				"-args", "$(rm -rf /)",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			stdoutContains: []string{
				`-args 'hello world'`,
				`-args 'it'\''s me'`,
				`-args '$(rm -rf /)'`,
			},
		},
		{
			name: "missing pkgpath rejected",
			args: []string{
				"--no-interactive", "maketx", "call",
				"-func", "Bar",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			wantErrSubstr: "pkgpath not specified",
		},
		{
			name: "missing func rejected",
			args: []string{
				"--no-interactive", "maketx", "call",
				"-pkgpath", "gno.land/r/demo/foo",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			wantErrSubstr: "func not specified",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := run(t, "", tt.args...)
			assertOutcome(t, struct {
				name           string
				args           []string
				wantStdout     string
				wantErrSubstr  string
				stdoutContains []string
			}{tt.name, tt.args, tt.wantStdout, tt.wantErrSubstr, tt.stdoutContains}, stdout, err)
			for _, abs := range tt.stdoutAbsent {
				if strings.Contains(stdout, abs) {
					t.Errorf("expected stdout NOT to contain %q, got: %s", abs, stdout)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------
// addpkg
// ---------------------------------------------------------------------

func TestFunctional_AddPkg(t *testing.T) {
	tmpDir := t.TempDir()

	cases := []struct {
		name           string
		args           []string
		wantStdout     string
		wantErrSubstr  string
		stdoutContains []string
	}{
		{
			name: "happy minimal",
			args: []string{
				"--no-interactive", "maketx", "addpkg",
				"-pkgpath", "gno.land/r/demo/foo",
				"-pkgdir", tmpDir,
				"-gas-wanted", "1000000", "-gas-fee", "1ugnot",
				"mykey",
			},
			stdoutContains: []string{
				"gnokey maketx addpkg",
				"-pkgpath gno.land/r/demo/foo",
				"-pkgdir " + tmpDir,
				"mykey",
			},
		},
		{
			name: "with send and max-deposit",
			args: []string{
				"--no-interactive", "maketx", "addpkg",
				"-pkgpath", "gno.land/r/demo/foo",
				"-pkgdir", tmpDir,
				"-send", "1ugnot", "-max-deposit", "5ugnot",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			stdoutContains: []string{"-send 1ugnot", "-max-deposit 5ugnot"},
		},
		{
			name: "missing pkgpath rejected",
			args: []string{
				"--no-interactive", "maketx", "addpkg",
				"-pkgdir", tmpDir,
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			wantErrSubstr: "pkgpath not specified",
		},
		{
			name: "missing pkgdir rejected",
			args: []string{
				"--no-interactive", "maketx", "addpkg",
				"-pkgpath", "gno.land/r/demo/foo",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			wantErrSubstr: "pkgdir not specified",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := run(t, "", tt.args...)
			assertOutcome(t, tt, stdout, err)
		})
	}
}

// ---------------------------------------------------------------------
// run
// ---------------------------------------------------------------------

func TestFunctional_Run(t *testing.T) {
	cases := []struct {
		name           string
		args           []string
		wantStdout     string
		wantErrSubstr  string
		stdoutContains []string
	}{
		{
			name: "happy with two positionals",
			args: []string{
				"--no-interactive", "maketx", "run",
				"-gas-wanted", "1000000", "-gas-fee", "1ugnot",
				"-remote", "rpc.gno.land:443", "-chainid", "gnoland1",
				"mykey", "./script.gno",
			},
			wantStdout: "gnokey maketx run -remote rpc.gno.land:443 -gas-wanted 1000000 -gas-fee 1ugnot -chainid gnoland1 mykey ./script.gno",
		},
		{
			name: "stdin source positional emitted as '-'",
			args: []string{
				"--no-interactive", "maketx", "run",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey", "-",
			},
			stdoutContains: []string{"mykey -"},
		},
		{
			name: "with send and max-deposit",
			args: []string{
				"--no-interactive", "maketx", "run",
				"-send", "1ugnot", "-max-deposit", "5ugnot",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey", "./script.gno",
			},
			stdoutContains: []string{"-send 1ugnot", "-max-deposit 5ugnot"},
		},
		{
			name: "missing second positional rejected non-interactive",
			args: []string{
				"--no-interactive", "maketx", "run",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
				"mykey",
			},
			wantErrSubstr: "run [flags] <key>",
		},
		{
			name: "no positionals rejected non-interactive",
			args: []string{
				"--no-interactive", "maketx", "run",
				"-gas-wanted", "1", "-gas-fee", "1ugnot",
			},
			wantErrSubstr: "run [flags] <key>",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			stdout, _, err := run(t, "", tt.args...)
			assertOutcome(t, tt, stdout, err)
		})
	}
}

// ---------------------------------------------------------------------
// root-level / cross-cutting paths
// ---------------------------------------------------------------------

func TestFunctional_RunFlagWithNoInteractiveRejected(t *testing.T) {
	const addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	stdout, _, err := run(t, "",
		"--run", "--no-interactive", "maketx", "send",
		"-to", addr, "-send", "1ugnot", "-gas-wanted", "1", "-gas-fee", "1ugnot",
		"-remote", "x", "-chainid", "x",
		"mykey",
	)
	if err == nil || !strings.Contains(err.Error(), "--run requires interactive confirmation") {
		t.Fatalf("expected ErrNoInteractiveRun-style error, got %v", err)
	}
	// The command is still printed (the primary deliverable).
	if !strings.Contains(stdout, "gnokey maketx send") {
		t.Errorf("expected printed command in stdout, got %q", stdout)
	}
}

func TestFunctional_MemoWithShellMetachars(t *testing.T) {
	const addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	stdout, _, err := run(t, "",
		"--no-interactive", "maketx", "send",
		"-to", addr, "-send", "1ugnot",
		"-gas-wanted", "1", "-gas-fee", "1ugnot",
		"-memo", "rm -rf / && echo pwned",
		"-remote", "x", "-chainid", "x",
		"mykey",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `-memo 'rm -rf / && echo pwned'`) {
		t.Errorf("memo not quoted safely: %s", stdout)
	}
}

func TestFunctional_HelpDoesNotError(t *testing.T) {
	// `gnoland-cli --help` and `gnoland-cli maketx --help` should not return
	// an error (flag.ErrHelp is swallowed by Execute, but our tests use
	// ParseAndRun which propagates it). Just verify the help path runs.
	cases := [][]string{
		{"--help"},
		{"maketx", "--help"},
		{"maketx", "send", "--help"},
		{"maketx", "call", "--help"},
		{"maketx", "addpkg", "--help"},
		{"maketx", "run", "--help"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			_, _, err := run(t, "", args...)
			// flag.ErrHelp is expected and benign — the standard library
			// returns it after printing usage.
			if err != nil && !strings.Contains(err.Error(), "help") {
				t.Errorf("unexpected non-help error: %v", err)
			}
		})
	}
}

// ---------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------

// assertOutcome checks stdout-exact, stdout-contains, and error-substring.
//   - wantErrSubstr != "" => expect err with that substring.
//   - wantErrSubstr == "" => expect nil err, then check wantStdout (if set)
//     and stdoutContains (if any).
func assertOutcome(t *testing.T, tt struct {
	name           string
	args           []string
	wantStdout     string
	wantErrSubstr  string
	stdoutContains []string
}, stdout string, err error) {
	t.Helper()
	if tt.wantErrSubstr != "" {
		if err == nil {
			t.Fatalf("expected error containing %q, got nil (stdout: %s)", tt.wantErrSubstr, stdout)
		}
		if !strings.Contains(err.Error(), tt.wantErrSubstr) {
			t.Fatalf("expected error containing %q, got %v", tt.wantErrSubstr, err)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v\nstdout: %s", err, stdout)
	}
	if tt.wantStdout != "" && stdout != tt.wantStdout {
		t.Fatalf("stdout mismatch:\n got: %q\nwant: %q", stdout, tt.wantStdout)
	}
	for _, sub := range tt.stdoutContains {
		if !strings.Contains(stdout, sub) {
			t.Errorf("expected stdout to contain %q, got: %s", sub, stdout)
		}
	}
}
