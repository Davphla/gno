package gen

import "flag"

// gnokey CLI defaults — keep in sync with tm2/pkg/crypto/keys/client/maketx.go.
const (
	DefaultBroadcast = true
	DefaultSimulate  = "test"
	DefaultChainID   = "dev"
)

// SharedFlags carries every gnokey base flag plus every `maketx` parent
// flag. Embedded into each maketx subcommand cfg so the emit code can read
// any field by name.
type SharedFlags struct {
	// Base (root-level gnokey flags).
	Home                  string
	Remote                string
	Quiet                 bool
	InsecurePasswordStdin bool
	Config                string

	// MakeTx (parent-level gnokey maketx flags).
	GasWanted int64
	GasFee    string
	Memo      string
	Broadcast bool
	Simulate  string
	ChainID   string
}

// Register registers every shared flag on fs with gnokey-matching names and
// defaults. Subcommand cfgs call this from their own RegisterFlags before
// adding subcommand-specific flags.
func (s *SharedFlags) Register(fs *flag.FlagSet) {
	// Base.
	fs.StringVar(&s.Home, "home", "", "home directory")
	fs.StringVar(&s.Remote, "remote", "", "remote node URL")
	fs.BoolVar(&s.Quiet, "quiet", false, "suppress output during execution")
	fs.BoolVar(&s.InsecurePasswordStdin, "insecure-password-stdin", false, "WARNING! take password from stdin")
	fs.StringVar(&s.Config, "config", "", "config file (optional)")

	// MakeTx.
	fs.Int64Var(&s.GasWanted, "gas-wanted", 0, "gas requested for tx")
	fs.StringVar(&s.GasFee, "gas-fee", "", "gas payment fee")
	fs.StringVar(&s.Memo, "memo", "", "any descriptive text")
	fs.BoolVar(&s.Broadcast, "broadcast", DefaultBroadcast, "sign, simulate and broadcast")
	fs.StringVar(&s.Simulate, "simulate", DefaultSimulate, "select how to simulate the transaction (test|skip|only)")
	fs.StringVar(&s.ChainID, "chainid", DefaultChainID, "chainid to sign for")
}

// EmitOn appends every set shared flag onto b in the canonical order
// (base flags first, then maketx flags). Skips flags at their defaults.
func (s *SharedFlags) EmitOn(b *Builder) {
	// Base.
	b.StringFlag("-home", s.Home)
	b.StringFlag("-remote", s.Remote)
	b.BoolFlag("-quiet", s.Quiet, false)
	b.BoolFlag("-insecure-password-stdin", s.InsecurePasswordStdin, false)
	b.StringFlag("-config", s.Config)
	// MakeTx.
	b.Int64Flag("-gas-wanted", s.GasWanted)
	b.StringFlag("-gas-fee", s.GasFee)
	b.StringFlag("-memo", s.Memo)
	b.BoolFlag("-broadcast", s.Broadcast, DefaultBroadcast)
	b.StringFlagIfChanged("-simulate", s.Simulate, DefaultSimulate)
	b.StringFlagIfChanged("-chainid", s.ChainID, DefaultChainID)
}
