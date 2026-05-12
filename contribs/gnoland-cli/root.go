package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// rootCfg holds tool-level (non-gnokey) flags shared by every subcommand.
type rootCfg struct {
	Run           bool
	NoInteractive bool
}

func (c *rootCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.Run, "run", false, "execute the generated command via `bash -c` after a /!\\ warning and y/N confirmation")
	fs.BoolVar(&c.NoInteractive, "no-interactive", false, "disable interactive prompts; fail fast when a required field is missing")
}

func newRootCmd(io commands.IO) *commands.Command {
	cfg := &rootCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnoland-cli",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "helper that emits gnokey commands (and optionally runs them via bash -c)",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newMakeTxCmd(cfg, io),
	)

	return cmd
}
