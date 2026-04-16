// Package qrview provides a Bubble Tea component that renders a QR code
// in the terminal using Unicode half-block characters.
//
// Two vertical QR modules are packed into each character cell using ▀,
// ▄, █, and space. Each byte of the internal grid holds 4 consecutive
// terminal cells (2 bits per cell × 4 = 8 bits), and a 256-entry lookup
// table maps each byte directly to a pre-rendered styled string.
//
// Rendering is lazy and cached: the first call to View() packs the
// bitmap, builds the lookup tables, and emits the styled code block;
// subsequent calls return the cached output via [sync.OnceValue]. The
// cache is replaced (not reset) by [Model.SetData] and
// [Model.SetDataFromString]. The label and border wrapper are applied
// on every View() call, so Label can be changed freely.
//
// This design targets embedded controllers where the QR code changes
// rarely (device URL, serial number) but may be re-rendered on every
// frame alongside live status data.
//
// Usage:
//
//	m := qrview.New(bitmap)
//	m, err := qrview.NewFromString("https://192.168.1.100/config")
//	fmt.Println(m.View())
package qrview

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/apptly-oss/bubblespice/ansi"
)

// Model is the Bubble Tea model for a QR code display.
//
// Field order is dictated by fieldalignment — the public Label and the
// enc payload are grouped in the middle rather than at the top.
//
// cspell:words fieldalignment
type Model struct {
	labelStyle  lipgloss.Style
	borderStyle lipgloss.Style

	// renderCode caches the pipeline from enc → styled code block
	// (no border, no label). Bound by [New]/[NewFromString] and
	// replaced (via a fresh [sync.OnceValue]) by [Model.SetData] and
	// [Model.SetDataFromString]. Value-copies of Model share the same
	// closure, so all copies of a given state see the same cached
	// output.
	renderCode func() string

	// Label is optional text rendered below the QR code. It is applied
	// on every View() call and can be changed without invalidating the
	// cached code block.
	Label string

	// Colours stored explicitly — no extraction from styles needed.
	darkColour  lipgloss.Color
	lightColour lipgloss.Color

	// enc holds the normalised bitmap and derived dimensions.
	enc encoder

	// inverted swaps dark/light at cell-table build time. Read by
	// buildCode, so it survives SetData the same way colours do.
	inverted bool
}

// New creates a QR view from a raw bitmap. bitmap[row][col] should be
// true for dark modules. [newEncoder] normalises and deep-copies the
// bitmap, so later caller mutations don't reach back into the model.
func New(bitmap [][]bool, opts ...Option) Model {
	m := newModel(newEncoder(bitmap, 2))
	for _, opt := range opts {
		opt(&m)
	}
	m.renderCode = sync.OnceValue(m.buildCode)
	return m
}

// NewFromString creates a QR view by encoding the given string.
func NewFromString(data string, opts ...Option) (Model, error) {
	e, err := newEncoderFromString(data, 2)
	if err != nil {
		return Model{}, err
	}
	m := newModel(e)
	for _, opt := range opts {
		opt(&m)
	}
	m.renderCode = sync.OnceValue(m.buildCode)
	return m, nil
}

// newModel returns a Model with default colours and label style set
// but with renderCode unbound. Callers must apply options and then
// bind renderCode via [sync.OnceValue].
func newModel(enc encoder) Model {
	return Model{
		enc:         enc,
		darkColour:  ansi.Black,
		lightColour: ansi.BrightWhite,
		labelStyle: lipgloss.NewStyle().
			Foreground(ansi.Grey252).
			Align(lipgloss.Center),
	}
}

// SetData replaces the QR bitmap. The cached render is replaced so
// the next View() call reflects the new bitmap.
func (m Model) SetData(bitmap [][]bool) Model {
	m.enc = m.enc.WithData(bitmap)
	m.renderCode = sync.OnceValue(m.buildCode)
	return m
}

// SetDataFromString re-encodes a string and replaces the bitmap.
func (m Model) SetDataFromString(data string) (Model, error) {
	e, err := newEncoderFromString(data, m.enc.QuietZone())
	if err != nil {
		return m, err
	}
	m.enc = e
	m.renderCode = sync.OnceValue(m.buildCode)
	return m, nil
}

// Empty returns true if no bitmap data has been set.
func (m Model) Empty() bool {
	return m.enc.Empty()
}

// Init is a no-op.
func (Model) Init() tea.Cmd { return nil }

// Update is a pass-through — display-only component.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders the QR code, using the cached code block and applying
// the border + optional label on every call.
func (m Model) View() string {
	if m.Empty() {
		return ""
	}
	code := m.borderStyle.Render(m.renderCode())
	if m.Label == "" {
		return code
	}
	cols, _ := m.enc.Size()
	label := m.labelStyle.Width(cols).Render(m.Label)
	return lipgloss.JoinVertical(lipgloss.Center, code, label)
}
