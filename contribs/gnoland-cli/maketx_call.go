package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/contribs/gnoland-cli/internal/gen"
	"github.com/gnolang/gno/contribs/gnoland-cli/internal/validate"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type callCfg struct {
	gen.SharedFlags
	PkgPath    string
	FuncName   string
	Args       commands.StringArr
	Send       string
	MaxDeposit string
}

func (c *callCfg) RegisterFlags(fs *flag.FlagSet) {
	c.SharedFlags.Register(fs)
	fs.StringVar(&c.PkgPath, "pkgpath", "", "package path (required)")
	fs.StringVar(&c.FuncName, "func", "", "function to call (required)")
	fs.Var(&c.Args, "args", "arguments to the function (repeatable)")
	fs.StringVar(&c.Send, "send", "", "send amount")
	fs.StringVar(&c.MaxDeposit, "max-deposit", "", "max storage deposit")
}

func newCallCmd(root *rootCfg, io commands.IO) *commands.Command {
	cfg := &callCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "call",
			ShortUsage: "call [flags] <key-name or address>",
			ShortHelp:  "emit `gnokey maketx call ...`",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execCall(root, cfg, args, io)
		},
	)
}

func execCall(root *rootCfg, cfg *callCfg, args []string, io commands.IO) error {
	keyName, err := resolvePositional(io, root, args, 0, 1, "Key name or address", validate.Required("key"))
	if err != nil {
		return err
	}
	if err := resolveString(io, root, &cfg.PkgPath, "pkgpath", "Package path (gno.land/...)", validate.PkgPath); err != nil {
		return err
	}
	if err := resolveString(io, root, &cfg.FuncName, "func", "Function name", validate.Required("func")); err != nil {
		return err
	}
	if err := resolveSharedGas(io, root, &cfg.SharedFlags); err != nil {
		return err
	}

	b := gen.New("gnokey").Add("maketx", "call")
	cfg.SharedFlags.EmitOn(b)
	b.StringFlag("-pkgpath", cfg.PkgPath)
	b.StringFlag("-func", cfg.FuncName)
	b.RepeatedStringFlag("-args", []string(cfg.Args))
	b.StringFlag("-send", cfg.Send)
	b.StringFlag("-max-deposit", cfg.MaxDeposit)
	b.Positional(keyName)

	return gen.Run(io, b.String(), root.Run, root.NoInteractive)
}
