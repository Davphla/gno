package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/contribs/gnoland-cli/internal/gen"
	"github.com/gnolang/gno/contribs/gnoland-cli/internal/validate"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type addpkgCfg struct {
	gen.SharedFlags
	PkgPath    string
	PkgDir     string
	Send       string
	MaxDeposit string
}

func (c *addpkgCfg) RegisterFlags(fs *flag.FlagSet) {
	c.SharedFlags.Register(fs)
	fs.StringVar(&c.PkgPath, "pkgpath", "", "package path (required)")
	fs.StringVar(&c.PkgDir, "pkgdir", "", "path to package files on disk (required)")
	fs.StringVar(&c.Send, "send", "", "send amount")
	fs.StringVar(&c.MaxDeposit, "max-deposit", "", "max storage deposit")
}

func newAddPkgCmd(root *rootCfg, io commands.IO) *commands.Command {
	cfg := &addpkgCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "addpkg",
			ShortUsage: "addpkg [flags] <key-name or address>",
			ShortHelp:  "emit `gnokey maketx addpkg ...`",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execAddPkg(root, cfg, args, io)
		},
	)
}

func execAddPkg(root *rootCfg, cfg *addpkgCfg, args []string, io commands.IO) error {
	keyName, err := resolvePositional(io, root, args, 0, 1, "Key name or address", validate.Required("key"))
	if err != nil {
		return err
	}
	if err := resolveString(io, root, &cfg.PkgPath, "pkgpath", "Package path (gno.land/...)", validate.PkgPath); err != nil {
		return err
	}
	if err := resolveString(io, root, &cfg.PkgDir, "pkgdir", "Package directory on disk", validate.Dir); err != nil {
		return err
	}
	if err := resolveSharedGas(io, root, &cfg.SharedFlags); err != nil {
		return err
	}

	b := gen.New("gnokey").Add("maketx", "addpkg")
	cfg.SharedFlags.EmitOn(b)
	b.StringFlag("-pkgpath", cfg.PkgPath)
	b.StringFlag("-pkgdir", cfg.PkgDir)
	b.StringFlag("-send", cfg.Send)
	b.StringFlag("-max-deposit", cfg.MaxDeposit)
	b.Positional(keyName)

	return gen.Run(io, b.String(), root.Run, root.NoInteractive)
}
