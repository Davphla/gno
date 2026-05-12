package gen

import (
	"testing"

	shellquote "github.com/kballard/go-shellquote"
)

// shellquoteSplit is a thin testing helper that delegates to the upstream
// Split (which itself errors on unbalanced quotes).
func shellquoteSplit(t *testing.T, s string) ([]string, error) {
	t.Helper()
	return shellquote.Split(s)
}

func TestBuilder_BasicEmission(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*Builder) *Builder
		want string
	}{
		{
			name: "program only",
			fn:   func(b *Builder) *Builder { return b },
			want: "gnokey",
		},
		{
			name: "Add subcommand chain",
			fn:   func(b *Builder) *Builder { return b.Add("maketx", "send") },
			want: "gnokey maketx send",
		},
		{
			name: "StringFlag empty value skipped",
			fn:   func(b *Builder) *Builder { return b.StringFlag("-to", "") },
			want: "gnokey",
		},
		{
			name: "StringFlag set",
			fn:   func(b *Builder) *Builder { return b.StringFlag("-to", "g1abc") },
			want: "gnokey -to g1abc",
		},
		{
			name: "Int64Flag zero skipped",
			fn:   func(b *Builder) *Builder { return b.Int64Flag("-gas-wanted", 0) },
			want: "gnokey",
		},
		{
			name: "Int64Flag set",
			fn:   func(b *Builder) *Builder { return b.Int64Flag("-gas-wanted", 1_000_000) },
			want: "gnokey -gas-wanted 1000000",
		},
		{
			name: "BoolFlag at default skipped",
			fn:   func(b *Builder) *Builder { return b.BoolFlag("-broadcast", true, true) },
			want: "gnokey",
		},
		{
			name: "BoolFlag overrides default",
			fn:   func(b *Builder) *Builder { return b.BoolFlag("-broadcast", false, true) },
			want: "gnokey -broadcast=false",
		},
		{
			name: "StringFlagIfChanged at default skipped",
			fn:   func(b *Builder) *Builder { return b.StringFlagIfChanged("-simulate", "test", "test") },
			want: "gnokey",
		},
		{
			name: "StringFlagIfChanged different from default",
			fn:   func(b *Builder) *Builder { return b.StringFlagIfChanged("-simulate", "skip", "test") },
			want: "gnokey -simulate skip",
		},
		{
			name: "RepeatedStringFlag empty slice skipped",
			fn:   func(b *Builder) *Builder { return b.RepeatedStringFlag("-args", nil) },
			want: "gnokey",
		},
		{
			name: "RepeatedStringFlag emits one pair per element in order",
			fn:   func(b *Builder) *Builder { return b.RepeatedStringFlag("-args", []string{"a", "b", "c"}) },
			want: "gnokey -args a -args b -args c",
		},
		{
			name: "Positional appended at end",
			fn:   func(b *Builder) *Builder { return b.StringFlag("-to", "g1abc").Positional("mykey") },
			want: "gnokey -to g1abc mykey",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(New("gnokey")).String()
			if got != tt.want {
				t.Fatalf("got %q\nwant %q", got, tt.want)
			}
		})
	}
}

// TestBuilder_ShellQuoting pins the exact strings emitted for problematic
// inputs. The shell-quoting library (kballard/go-shellquote) is trusted,
// but pinning these means any upstream behavior change shows up loudly.
//
// The contract: the printed string is what gets passed to bash -c, so each
// case below must be safe when interpreted by bash. The companion
// TestBuilder_ShellQuoting_RoundTrip test verifies that round-tripping
// emit -> shell-split returns the original tokens (the real safety check).
//
// Note that go-shellquote mixes single-quote wrapping for tokens containing
// whitespace / single quotes / dollar signs with backslash escaping for
// individual metacharacters (`;`, `|`, `*`, `\`). Both styles produce
// inert tokens under bash.
func TestBuilder_ShellQuoting(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{"empty token", "", "''"},
		{"safe alphanumeric", "abc123", "abc123"},
		{"safe punctuation", "gno.land/r/demo/foo", "gno.land/r/demo/foo"},
		{"space", "hello world", "'hello world'"},
		{"single quote", "it's me", `'it'\''s me'`},
		{"double quote", `"hi"`, `\"hi\"`},
		{"dollar command substitution", "$(rm -rf /)", "'$(rm -rf /)'"},
		{"backtick command substitution", "`rm -rf /`", "'`rm -rf /`'"},
		{"semicolon chained", "a;b", `a\;b`},
		{"pipe", "a|b", `a\|b`},
		{"glob", "foo*", `foo\*`},
		{"backslash", `a\b`, `a\\b`},
		{"newline", "line1\nline2", "'line1\nline2'"},
		{"unicode", "café", "café"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New("gnokey").Positional(tt.token)
			want := "gnokey " + tt.want
			if got := b.String(); got != want {
				t.Fatalf("got %q\nwant %q", got, want)
			}
		})
	}
}

// TestBuilder_ShellQuoting_RoundTrip verifies the real safety invariant:
// every token emitted by the builder must round-trip through shellquote.Split
// back to the original input. If this property holds for every token, then
// `bash -c <emitted>` will see the original argv when running gnokey.
func TestBuilder_ShellQuoting_RoundTrip(t *testing.T) {
	dangerous := []string{
		"",
		"abc",
		"hello world",
		"it's me",
		`"hi"`,
		"$(rm -rf /)",
		"`rm -rf /`",
		"a;b",
		"a|b",
		"foo*",
		`a\b`,
		"line1\nline2",
		"café",
		// every ASCII metachar, just in case
		"<>&;|*?[]{}()$`!\"'#~ \t\n\\",
	}
	for _, tok := range dangerous {
		t.Run(tok, func(t *testing.T) {
			b := New("gnokey").Positional(tok, "trailing")
			emitted := b.String()
			parts, err := shellquoteSplit(t, emitted)
			if err != nil {
				t.Fatalf("Split(%q) error: %v", emitted, err)
			}
			if len(parts) != 3 {
				t.Fatalf("Split(%q) = %q, want 3 tokens", emitted, parts)
			}
			if parts[0] != "gnokey" || parts[1] != tok || parts[2] != "trailing" {
				t.Fatalf("round-trip mismatch:\n  in:  %q\n  out: %q", tok, parts[1])
			}
		})
	}
}

func TestBuilder_FullSendCommand(t *testing.T) {
	b := New("gnokey").Add("maketx", "send")
	b.StringFlag("-remote", "rpc.gno.land:443")
	b.Int64Flag("-gas-wanted", 1_000_000)
	b.StringFlag("-gas-fee", "1ugnot")
	b.StringFlagIfChanged("-chainid", "gnoland1", "dev")
	b.StringFlag("-to", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	b.StringFlag("-send", "1ugnot")
	b.Positional("mykey")

	want := "gnokey maketx send -remote rpc.gno.land:443 -gas-wanted 1000000 -gas-fee 1ugnot -chainid gnoland1 -to g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -send 1ugnot mykey"
	if got := b.String(); got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}
