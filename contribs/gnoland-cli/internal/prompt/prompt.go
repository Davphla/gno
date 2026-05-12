// Package prompt is a thin wrapper around AlecAivazis/survey/v2.
//
// All survey calls go through this package so:
//   - survey.WithStdio is plumbed in from commands.IO at one place, making
//     unit tests with commands.NewTestIO() drive the prompts predictably;
//   - if a different prompt library replaces survey, only this file changes.
package prompt

import (
	"io"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"golang.org/x/term"
)

// IsInteractive reports whether prompts can be displayed for the given IO.
// True only when the underlying stdin is a terminal. False for piped stdin
// (tests, scripts, CI). Detection relies on the In() reader being a
// *os.File whose fd is a tty.
func IsInteractive(cio commands.IO) bool {
	f, ok := cio.In().(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// AskString prompts for a string. value is updated in place. validator may
// be nil. Cancellation (Ctrl-C / EOF) returns a non-nil error.
func AskString(cio commands.IO, label string, value *string, validator survey.Validator) error {
	q := &survey.Input{Message: label, Default: *value}
	opts := stdioOpts(cio)
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}
	return survey.AskOne(q, value, opts...)
}

// AskSelect prompts the user to pick one option from a fixed list.
// Returns the selected option text.
func AskSelect(cio commands.IO, label string, options []string) (string, error) {
	q := &survey.Select{Message: label, Options: options}
	var out string
	if err := survey.AskOne(q, &out, stdioOpts(cio)...); err != nil {
		return "", err
	}
	return out, nil
}

// AskConfirm prompts for a yes/no answer. defaultYes sets the default when
// the user just hits Enter.
func AskConfirm(cio commands.IO, label string, defaultYes bool) (bool, error) {
	q := &survey.Confirm{Message: label, Default: defaultYes}
	var out bool
	if err := survey.AskOne(q, &out, stdioOpts(cio)...); err != nil {
		return false, err
	}
	return out, nil
}

func stdioOpts(cio commands.IO) []survey.AskOpt {
	in, ok := cio.In().(terminal.FileReader)
	if !ok {
		in = fileReaderShim{r: cio.In()}
	}
	out, ok := cio.Out().(terminal.FileWriter)
	if !ok {
		out = fileWriterShim{w: cio.Out()}
	}
	return []survey.AskOpt{
		survey.WithStdio(in, out, cio.Err()),
	}
}

// fileReaderShim adapts a plain io.Reader to terminal.FileReader for survey.
// Fd() returns 0 (stdin) which survey treats as non-tty; safe for tests.
type fileReaderShim struct{ r io.Reader }

func (f fileReaderShim) Read(p []byte) (int, error) { return f.r.Read(p) }
func (fileReaderShim) Fd() uintptr                  { return 0 }

type fileWriterShim struct{ w io.Writer }

func (f fileWriterShim) Write(p []byte) (int, error) { return f.w.Write(p) }
func (fileWriterShim) Fd() uintptr                   { return 1 }
