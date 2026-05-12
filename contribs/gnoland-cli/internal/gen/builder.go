// Package gen builds a bash command string from typed flag inputs.
//
// Each token written to a Builder is run through shellquote.Join (POSIX
// single-quote escaping), so the resulting string is safe to print AND safe
// to pass to `bash -c`. The same string serves both purposes.
package gen

import (
	"fmt"
	"strconv"

	shellquote "github.com/kballard/go-shellquote"
)

// Builder assembles a bash one-liner from a sequence of literal tokens,
// flags, and positional arguments. Call String() to render.
type Builder struct {
	tokens []string
}

// New starts a builder seeded with one or more literal program tokens
// (typically the program name, e.g. "gnokey").
func New(program ...string) *Builder {
	b := &Builder{}
	b.tokens = append(b.tokens, program...)
	return b
}

// Add appends one or more literal tokens. Use for subcommand names like
// "maketx", "send". Tokens are still shell-quoted on render.
func (b *Builder) Add(tokens ...string) *Builder {
	b.tokens = append(b.tokens, tokens...)
	return b
}

// StringFlag appends `name value` only when value is non-empty.
func (b *Builder) StringFlag(name, value string) *Builder {
	if value == "" {
		return b
	}
	b.tokens = append(b.tokens, name, value)
	return b
}

// StringFlagIfChanged appends `name value` only when value != defaultValue.
// Use for flags with non-empty defaults (e.g. -simulate=test, -chainid=dev).
func (b *Builder) StringFlagIfChanged(name, value, defaultValue string) *Builder {
	if value == defaultValue {
		return b
	}
	b.tokens = append(b.tokens, name, value)
	return b
}

// Int64Flag appends `name value` only when value != 0.
func (b *Builder) Int64Flag(name string, value int64) *Builder {
	if value == 0 {
		return b
	}
	b.tokens = append(b.tokens, name, strconv.FormatInt(value, 10))
	return b
}

// BoolFlag appends `name=<value>` only when value differs from defaultValue.
// gnokey uses --flag=false form for explicit overrides, which is what we
// emit here for clarity (`-broadcast=false` rather than just `-broadcast`).
func (b *Builder) BoolFlag(name string, value, defaultValue bool) *Builder {
	if value == defaultValue {
		return b
	}
	b.tokens = append(b.tokens, fmt.Sprintf("%s=%t", name, value))
	return b
}

// RepeatedStringFlag appends one `name value` pair per element of values.
// Use for gnokey's `-args` flag which can be passed multiple times.
func (b *Builder) RepeatedStringFlag(name string, values []string) *Builder {
	for _, v := range values {
		b.tokens = append(b.tokens, name, v)
	}
	return b
}

// Positional appends one or more positional argument tokens at the end of
// the command (after all flags).
func (b *Builder) Positional(args ...string) *Builder {
	b.tokens = append(b.tokens, args...)
	return b
}

// String renders the command as a single bash-safe line. Each token is
// individually shell-quoted via shellquote.Join.
func (b *Builder) String() string {
	return shellquote.Join(b.tokens...)
}
