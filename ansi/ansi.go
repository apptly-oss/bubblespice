// Package ansi provides helpers for handling ANSI escape sequences in
// strings produced by terminal styling libraries.
package ansi

import "regexp"

// ansiRE matches a minimal ANSI escape sequence: ESC followed by any
// run of non-letter bytes and a terminating ASCII letter. It covers
// the CSI sequences emitted by terminal styling libraries such as
// lipgloss; it does not implement the full ECMA-48 grammar.
var ansiRE = regexp.MustCompile(`\x1b[^A-Za-z]*[A-Za-z]`)

// Strip removes ANSI escape sequences from s, returning the visible
// text. It is intended for tests and diagnostics that need to match
// against rendered output without the styling.
func Strip(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}
