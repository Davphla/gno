// Package validate holds syntactic validators for user-supplied inputs.
// Validators never touch the network, the keybase, or the filesystem
// (except for Dir, which stats the given path).
package validate

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Required wraps a validator to reject the empty string with a labelled
// error message. Returns a survey-compatible validator (any -> error).
func Required(label string) func(any) error {
	return func(v any) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("%s: expected string", label)
		}
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s is required", label)
		}
		return nil
	}
}

// Bech32 validates that the input parses as a bech32 address.
func Bech32(v any) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("expected string")
	}
	if s == "" {
		return errors.New("address is required")
	}
	if _, err := crypto.AddressFromBech32(s); err != nil {
		return fmt.Errorf("invalid bech32 address: %w", err)
	}
	return nil
}

// Coins validates that the input parses as a std.Coins (e.g. "1ugnot").
// Empty input is allowed (represents "no send amount").
func Coins(v any) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("expected string")
	}
	if s == "" {
		return nil
	}
	if _, err := std.ParseCoins(s); err != nil {
		return fmt.Errorf("invalid coins: %w", err)
	}
	return nil
}

// RequiredCoins is Coins but rejects empty input.
func RequiredCoins(v any) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("expected string")
	}
	if s == "" {
		return errors.New("coins required")
	}
	return Coins(s)
}

// Coin validates that the input parses as a single std.Coin.
func Coin(v any) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("expected string")
	}
	if s == "" {
		return errors.New("coin required")
	}
	if _, err := std.ParseCoin(s); err != nil {
		return fmt.Errorf("invalid coin: %w", err)
	}
	return nil
}

// PositiveInt validates that the input parses as a positive int64.
func PositiveInt(v any) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("expected string")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid integer: %s", s)
	}
	if n <= 0 {
		return errors.New("must be positive")
	}
	return nil
}

// PkgPath validates the surface shape of a gno.land package path. It does
// NOT check for existence on chain — that's gnokey's job at run time.
func PkgPath(v any) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("expected string")
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("pkgpath is required")
	}
	if !strings.HasPrefix(s, "gno.land/") {
		return errors.New("pkgpath must start with gno.land/")
	}
	if strings.Contains(s, "//") || strings.Contains(s, " ") {
		return errors.New("pkgpath must not contain '//' or spaces")
	}
	return nil
}

// Dir validates that the input is an existing directory on the filesystem.
func Dir(v any) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("expected string")
	}
	if s == "" {
		return errors.New("directory is required")
	}
	info, err := os.Stat(s)
	if err != nil {
		return fmt.Errorf("cannot access directory: %w", err)
	}
	if !info.IsDir() {
		return errors.New("not a directory")
	}
	return nil
}
