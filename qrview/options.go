package qrview

import "github.com/charmbracelet/lipgloss"

// Option configures a Model at construction time.
type Option func(*Model)

// WithQuietZone sets the number of module-widths of white padding around
// the QR code. Default is 2.
func WithQuietZone(n int) Option {
	return func(m *Model) {
		m.enc = m.enc.WithQuietZone(n)
	}
}

// WithLabel sets a text label rendered below the QR code.
func WithLabel(s string) Option {
	return func(m *Model) { m.Label = s }
}

// WithColours sets the dark and light module colours directly.
// Default is black (0) and white (15).
func WithColours(dark, light lipgloss.Color) Option {
	return func(m *Model) {
		m.darkColour = dark
		m.lightColour = light
	}
}

// WithLabelStyle sets the style for the label text.
func WithLabelStyle(s lipgloss.Style) Option {
	return func(m *Model) { m.labelStyle = s }
}

// WithBorderStyle sets the style wrapping the rendered code.
func WithBorderStyle(s lipgloss.Style) Option {
	return func(m *Model) { m.borderStyle = s }
}

// WithInverted flips all modules — dark becomes light and vice versa.
// Useful for light terminal backgrounds. Costs one XOR per 4 cells at
// render time; no re-encoding or style rebuild.
func WithInverted() Option {
	return func(m *Model) { m.inverted = true }
}
