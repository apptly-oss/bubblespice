package ansi

import "github.com/charmbracelet/lipgloss"

// Named ANSI palette entries as [lipgloss.Color] values, suitable for
// direct use with lipgloss styles. The numeric forms ("0", "15", …)
// remain valid; these constants are provided so call sites can avoid
// bare numeric literals.
//
// For the full 16-colour and 256-colour palettes see
// [github.com/charmbracelet/x/ansi] — its BasicColor / IndexedColor
// values are not lipgloss.Color-typed, so they must be converted
// before use with lipgloss styles, but the naming and numeric ranges
// are canonical.
const (
	// Black is colour 0 from the 16-colour palette.
	Black lipgloss.Color = "0"
	// BrightWhite is colour 15 from the 16-colour palette (the
	// high-intensity white).
	BrightWhite lipgloss.Color = "15"
	// Grey252 is colour 252 from the 256-colour greyscale ramp — a
	// near-white grey commonly used for secondary text on dark
	// backgrounds.
	Grey252 lipgloss.Color = "252"
)
