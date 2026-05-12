package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/contribs/gnoland-cli/internal/gen"
	"github.com/gnolang/gno/contribs/gnoland-cli/internal/validate"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type sendCfg struct {
	gen.SharedFlags
	To   string
	Send string
}

func (c *sendCfg) RegisterFlags(fs *flag.FlagSet) {
	c.SharedFlags.Register(fs)
	fs.StringVar(&c.To, "to", "", "destination address (required)")
	fs.StringVar(&c.Send, "send", "", "send amount (required, e.g. 1ugnot)")
}

func newSendCmd(root *rootCfg, io commands.IO) *commands.Command {
	cfg := &sendCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "send",
			ShortUsage: "send [flags] <key-name or address>",
			ShortHelp:  "emit `gnokey maketx send ...`",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execSend(root, cfg, args, io)
		},
	)
}

func execSend(root *rootCfg, cfg *sendCfg, args []string, io commands.IO) error {
	keyName, err := resolvePositional(io, root, args, 0, 1, "Key name or address", validate.Required("key"))
	if err != nil {
		return err
	}
	if err := resolveString(io, root, &cfg.To, "to", "Destination address", validate.Bech32); err != nil {
		return err
	}
	if err := resolveString(io, root, &cfg.Send, "send", "Send amount (e.g. 1ugnot)", validate.RequiredCoins); err != nil {
		return err
	}
	if err := resolveSharedGas(io, root, &cfg.SharedFlags); err != nil {
		return err
	}

	b := gen.New("gnokey").Add("maketx", "send")
	cfg.SharedFlags.EmitOn(b)
	b.StringFlag("-to", cfg.To)
	b.StringFlag("-send", cfg.Send)
	b.Positional(keyName)

	return gen.Run(io, b.String(), root.Run, root.NoInteractive)
}
