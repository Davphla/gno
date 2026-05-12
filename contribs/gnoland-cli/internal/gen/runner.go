package gen

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// ErrNoInteractiveRun is returned when --run is requested but --no-interactive
// disables the confirmation prompt. Unattended execution must be done
// explicitly by the caller (e.g. `gnoland-cli ... | bash`).
var ErrNoInteractiveRun = errors.New("--run requires interactive confirmation; pipe the printed command through bash for unattended execution")

// Run prints `command` to stdout, then either returns (the user just wanted
// the string), or shows a warning and asks for explicit confirmation before
// executing it via `bash -c`.
//
// When run is true and noInteractive is true, Run returns ErrNoInteractiveRun
// without executing.
func Run(io commands.IO, command string, run, noInteractive bool) error {
	io.Println(command)

	if !run {
		return nil
	}
	if noInteractive {
		return ErrNoInteractiveRun
	}

	io.ErrPrintln("")
	io.ErrPrintln("/!\\ This will execute the command above in your shell. Review it carefully.")

	ok, err := confirmDefaultNo(io, "Run this command?")
	if err != nil {
		return err
	}
	if !ok {
		io.ErrPrintln("Aborted.")
		return nil
	}

	return execBash(command)
}

// confirmDefaultNo reads one line of input and accepts only "y" or "yes"
// (case-insensitive). Empty input (Enter) defaults to no, opposite of
// tm2's GetConfirmation, which is the safer default for destructive
// actions.
func confirmDefaultNo(cio commands.IO, prompt string) (bool, error) {
	fmt.Fprintf(cio.Err(), "%s [y/N]: ", prompt)
	line, err := bufio.NewReader(cio.In()).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}

func execBash(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return commands.ExitCodeError(ee.ExitCode())
		}
		return fmt.Errorf("running bash: %w", err)
	}
	return nil
}
