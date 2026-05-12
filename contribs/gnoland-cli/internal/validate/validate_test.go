package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRequired(t *testing.T) {
	v := Required("name")
	if err := v(""); err == nil {
		t.Errorf("expected error for empty")
	}
	if err := v("   "); err == nil {
		t.Errorf("expected error for whitespace-only")
	}
	if err := v("ok"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := v(42); err == nil {
		t.Errorf("expected error for non-string")
	}
}

func TestBech32(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"", true},
		{"not-a-bech32", true},
		{"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			err := Bech32(c.in)
			if c.wantErr && err == nil {
				t.Error("expected error")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCoinsAndRequiredCoins(t *testing.T) {
	if err := Coins(""); err != nil {
		t.Errorf("Coins on empty should be nil, got %v", err)
	}
	if err := RequiredCoins(""); err == nil {
		t.Error("RequiredCoins on empty should error")
	}
	if err := Coins("1ugnot"); err != nil {
		t.Errorf("Coins on 1ugnot: %v", err)
	}
	if err := Coins("not-coins"); err == nil {
		t.Error("Coins on garbage should error")
	}
}

func TestCoin(t *testing.T) {
	if err := Coin(""); err == nil {
		t.Error("Coin on empty should error")
	}
	if err := Coin("1ugnot"); err != nil {
		t.Errorf("Coin on 1ugnot: %v", err)
	}
	if err := Coin("1ugnot,2gnot"); err == nil {
		t.Error("Coin must reject multi-coin")
	}
}

func TestPositiveInt(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"1000000", false},
		{"1", false},
		{"0", true},
		{"-1", true},
		{"abc", true},
		{"", true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			err := PositiveInt(c.in)
			if c.wantErr && err == nil {
				t.Error("expected error")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPkgPath(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"gno.land/r/demo/foo", false},
		{"gno.land/p/demo/avl", false},
		{"", true},
		{"github.com/foo/bar", true},
		{"gno.land//r/foo", true},
		{"gno.land/r/foo bar", true},
		{"  gno.land/r/demo/foo  ", false}, // trimmed
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			err := PkgPath(c.in)
			if c.wantErr && err == nil {
				t.Error("expected error")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDir(t *testing.T) {
	tmp := t.TempDir()
	if err := Dir(tmp); err != nil {
		t.Errorf("real dir errored: %v", err)
	}

	file := filepath.Join(tmp, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Dir(file); err == nil {
		t.Error("Dir on a file should error")
	}
	if err := Dir(filepath.Join(tmp, "nope")); err == nil {
		t.Error("Dir on nonexistent path should error")
	}
	if err := Dir(""); err == nil {
		t.Error("Dir on empty should error")
	}
}
