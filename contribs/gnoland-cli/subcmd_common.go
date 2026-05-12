package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/gnolang/gno/contribs/gnoland-cli/internal/gen"
	"github.com/gnolang/gno/contribs/gnoland-cli/internal/prompt"
	"github.com/gnolang/gno/contribs/gnoland-cli/internal/validate"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// resolvePositional returns args[n-1] when len(args) == expected, otherwise
// prompts when interactive, otherwise errors with flag.ErrHelp.
func resolvePositional(io commands.IO, root *rootCfg, args []string, idx, expected int, label string, validator survey.Validator) (string, error) {
	if len(args) == expected {
		return args[idx], nil
	}
	if len(args) > expected {
		return "", flag.ErrHelp
	}
	// fewer than expected — try prompting
	if root.NoInteractive || !prompt.IsInteractive(io) {
		return "", fmt.Errorf("%s required", label)
	}
	var v string
	if err := prompt.AskString(io, label, &v, validator); err != nil {
		return "", err
	}
	return v, nil
}

// resolveString fills *dst if it's empty, either by prompting (when
// interactive) or by returning an error naming the missing flag.
func resolveString(io commands.IO, root *rootCfg, dst *string, flagName, label string, validator survey.Validator) error {
	if *dst != "" {
		return nil
	}
	if root.NoInteractive || !prompt.IsInteractive(io) {
		return fmt.Errorf("%s not specified", flagName)
	}
	return prompt.AskString(io, label, dst, validator)
}

// resolveGasWanted fills *dst if zero. PositiveInt validator + ParseInt.
func resolveGasWanted(io commands.IO, root *rootCfg, dst *int64) error {
	if *dst != 0 {
		return nil
	}
	if root.NoInteractive || !prompt.IsInteractive(io) {
		return errors.New("gas-wanted not specified")
	}
	var s string
	if err := prompt.AskString(io, "Gas wanted (positive integer)", &s, validate.PositiveInt); err != nil {
		return err
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing gas-wanted: %w", err)
	}
	*dst = n
	return nil
}

// resolveSharedGas fills both gas-wanted and gas-fee on s.
func resolveSharedGas(io commands.IO, root *rootCfg, s *gen.SharedFlags) error {
	if err := resolveGasWanted(io, root, &s.GasWanted); err != nil {
		return err
	}
	return resolveString(io, root, &s.GasFee, "gas-fee", "Gas fee (e.g. 1ugnot)", validate.Coin)
}
