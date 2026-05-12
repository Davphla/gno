package gen

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newTestIO returns a commands.IO backed by buffers, plus the three buffers.
func newTestIO(stdin string) (commands.IO, *bytes.Buffer, *bytes.Buffer) {
	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(stdin))
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	io.SetOut(commands.WriteNopCloser(out))
	io.SetErr(commands.WriteNopCloser(errBuf))
	return io, out, errBuf
}

func TestRun_PrintOnly(t *testing.T) {
	io, out, errBuf := newTestIO("")
	err := Run(io, "echo hello", false /* run */, false /* noInteractive */)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimRight(out.String(), "\n"); got != "echo hello" {
		t.Errorf("stdout: got %q, want %q", got, "echo hello")
	}
	if errBuf.Len() != 0 {
		t.Errorf("expected empty stderr, got %q", errBuf.String())
	}
}

func TestRun_NoInteractiveWithRunRejected(t *testing.T) {
	io, out, _ := newTestIO("")
	err := Run(io, "echo hello", true /* run */, true /* noInteractive */)
	if !errors.Is(err, ErrNoInteractiveRun) {
		t.Fatalf("expected ErrNoInteractiveRun, got %v", err)
	}
	// Even when refusing, the command must still be printed (it's the
	// primary deliverable; the user can pipe it elsewhere).
	if !strings.Contains(out.String(), "echo hello") {
		t.Errorf("expected printed command in stdout, got %q", out.String())
	}
}

func TestRun_ConfirmDecline(t *testing.T) {
	// Decline at the y/N prompt — command is printed, warning shown, but
	// no exec happens.
	io, out, errBuf := newTestIO("n\n")
	err := Run(io, "false", true, false)
	if err != nil {
		t.Fatalf("expected nil on decline, got %v", err)
	}
	if !strings.Contains(out.String(), "false") {
		t.Errorf("expected printed command in stdout")
	}
	if !strings.Contains(errBuf.String(), "/!\\") {
		t.Errorf("expected /!\\ warning in stderr, got %q", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Aborted") {
		t.Errorf("expected Aborted in stderr, got %q", errBuf.String())
	}
}

func TestRun_ConfirmAcceptExecutes(t *testing.T) {
	// Accept at the prompt — command runs via bash -c. Use `true` so the
	// child exits 0; the runner inherits stdio, so we cannot capture child
	// stdout via our buffer, but we can verify the runner returns nil.
	io, _, _ := newTestIO("y\n")
	err := Run(io, "true", true, false)
	if err != nil {
		t.Fatalf("expected nil on accept of `true`, got %v", err)
	}
}

func TestRun_ConfirmEnterDefaultsToNo(t *testing.T) {
	// Enter with no answer must NOT execute. The destructive-action prompt
	// defaults to no, opposite of tm2's GetConfirmation default.
	io, _, errBuf := newTestIO("\n")
	err := Run(io, "false", true, false)
	if err != nil {
		t.Fatalf("expected nil on Enter (default-no), got %v", err)
	}
	if !strings.Contains(errBuf.String(), "[y/N]") {
		t.Errorf("expected default-no prompt format [y/N], got %q", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "Aborted") {
		t.Errorf("expected Aborted on default-no, got %q", errBuf.String())
	}
}

func TestRun_ConfirmAcceptForwardsExitCode(t *testing.T) {
	io, _, _ := newTestIO("y\n")
	err := Run(io, "exit 42", true, false)
	var ec commands.ExitCodeError
	if !errors.As(err, &ec) {
		t.Fatalf("expected ExitCodeError, got %T %v", err, err)
	}
	if int(ec) != 42 {
		t.Errorf("expected exit code 42, got %d", int(ec))
	}
}
