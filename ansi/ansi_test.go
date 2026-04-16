package ansi_test

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/apptly-oss/bubblespice/ansi"
)

type paletteCase struct {
	name string
	got  lipgloss.Color
	want lipgloss.Color
}

func (c paletteCase) Test(t *testing.T) {
	t.Helper()
	if c.got != c.want {
		t.Errorf("ansi.%s = %q, want %q", c.name, c.got, c.want)
	}
}

// TestPalette pins the named palette constants to their numeric
// lipgloss.Color values. Guards against an accidental rename or value
// drift breaking downstream renders that depend on specific ANSI
// codes.
func TestPalette(t *testing.T) {
	cases := []paletteCase{
		{name: "Black", got: ansi.Black, want: "0"},
		{name: "BrightWhite", got: ansi.BrightWhite, want: "15"},
		{name: "Grey252", got: ansi.Grey252, want: "252"},
	}
	for _, c := range cases {
		t.Run(c.name, c.Test)
	}
}

type stripCase struct {
	name string
	in   string
	want string
}

func (c stripCase) Test(t *testing.T) {
	t.Helper()
	got := ansi.Strip(c.in)
	if got != c.want {
		t.Errorf("Strip(%q) = %q, want %q", c.in, got, c.want)
	}
}

func TestStrip(t *testing.T) {
	// cspell:ignore mred mbold mfoo mbar
	cases := []stripCase{
		{name: "empty", in: "", want: ""},
		{name: "plain", in: "hello world", want: "hello world"},
		{name: "csi-colour", in: "\x1b[31mred\x1b[0m", want: "red"},
		{name: "nested", in: "\x1b[1m\x1b[31mbold red\x1b[0m\x1b[0m", want: "bold red"},
		{name: "utf8", in: "\x1b[32m✓\x1b[0m ok", want: "✓ ok"},
		{name: "newline", in: "\x1b[1mfoo\n\x1b[0mbar", want: "foo\nbar"},
		{name: "esc-between-words", in: "a\x1b[0mb", want: "ab"},
	}
	for _, c := range cases {
		t.Run(c.name, c.Test)
	}
}
