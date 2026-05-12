package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// emptyCfg is a no-flag config for parent commands whose only job is to
// list subcommands. Embedding it lets RegisterFlags be a no-op.
type emptyCfg struct{}

func (emptyCfg) RegisterFlags(*flag.FlagSet) {}

func newMakeTxCmd(root *rootCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "maketx",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "compose a `gnokey maketx ...` command",
		},
		emptyCfg{},
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newSendCmd(root, io),
		newCallCmd(root, io),
		newAddPkgCmd(root, io),
		newRunCmd(root, io),
	)

	return cmd
}
