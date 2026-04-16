package qrview

import "bytes"

// Render renders a QR bitmap to a string without a tea.Program.
func Render(bitmap [][]bool, opts ...Option) string {
	return New(bitmap, opts...).View()
}

// RenderString encodes a string as QR and returns rendered output.
func RenderString(data string, opts ...Option) (string, error) {
	m, err := NewFromString(data, opts...)
	if err != nil {
		return "", err
	}
	return m.View(), nil
}

// buildCode is the cached body of rendering: resolve the colour
// direction, build the cell/chunk lookup tables, pack the bitmap, and
// walk the packed grid into a styled string. Runs at most once per
// Model state via [sync.OnceValue] in [Model.renderCode].
func (m Model) buildCode() string {
	var b bytes.Buffer

	dark, light := m.darkColour, m.lightColour
	if m.inverted {
		dark, light = light, dark
	}
	grid := m.enc.Export()
	if grid.rows > 0 {
		b.Grow(grid.cols * grid.rows * 24)

		tables := newCodeTables(buildCellTable(dark, light))
		fullChunks := grid.cols / 4
		remainder := grid.cols % 4
		for y, row := range grid.data {
			if y > 0 {
				_, _ = b.WriteRune('\n')
			}
			tables.emitRow(&b, row, fullChunks, remainder)
		}
	}

	return b.String()
}
