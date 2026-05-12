package main

import (
	"context"
	"errors"
	"flag"

	"github.com/gnolang/gno/contribs/gnoland-cli/internal/gen"
	"github.com/gnolang/gno/contribs/gnoland-cli/internal/prompt"
	"github.com/gnolang/gno/contribs/gnoland-cli/internal/validate"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type runCfg struct {
	gen.SharedFlags
	Send       string
	MaxDeposit string
}

func (c *runCfg) RegisterFlags(fs *flag.FlagSet) {
	c.SharedFlags.Register(fs)
	fs.StringVar(&c.Send, "send", "", "send amount")
	fs.StringVar(&c.MaxDeposit, "max-deposit", "", "max storage deposit")
}

func newRunCmd(root *rootCfg, io commands.IO) *commands.Command {
	cfg := &runCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "run",
			ShortUsage: "run [flags] <key-name or address> <file-or-dir-or-->",
			ShortHelp:  "emit `gnokey maketx run ...`",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execRun(root, cfg, args, io)
		},
	)
}

func execRun(root *rootCfg, cfg *runCfg, args []string, io commands.IO) error {
	// Two positionals: <key> <source>.
	var keyName, source string
	switch len(args) {
	case 2:
		keyName, source = args[0], args[1]
	case 0, 1:
		if root.NoInteractive || !prompt.IsInteractive(io) {
			return errors.New("usage: run [flags] <key> <file-or-dir-or->")
		}
		if len(args) == 1 {
			keyName = args[0]
		} else {
			if err := prompt.AskString(io, "Key name or address", &keyName, validate.Required("key")); err != nil {
				return err
			}
		}
		if err := prompt.AskString(io, "Source file, directory, or '-' for stdin", &source, validate.Required("source")); err != nil {
			return err
		}
	default:
		return flag.ErrHelp
	}

	if err := resolveSharedGas(io, root, &cfg.SharedFlags); err != nil {
		return err
	}

	b := gen.New("gnokey").Add("maketx", "run")
	cfg.SharedFlags.EmitOn(b)
	b.StringFlag("-send", cfg.Send)
	b.StringFlag("-max-deposit", cfg.MaxDeposit)
	b.Positional(keyName, source)

	return gen.Run(io, b.String(), root.Run, root.NoInteractive)
}
