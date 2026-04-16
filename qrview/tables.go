package qrview

import (
	"io"

	"github.com/charmbracelet/lipgloss"
)

// cellTable maps each of the 4 cell kinds to a pre-rendered styled string.
type cellTable [4]string

// chunkTable maps each possible byte (4 packed cells) to a pre-rendered
// 4-character styled string. Built once from a cellTable.
type chunkTable [256]string

func buildCellTable(darkColour, lightColour lipgloss.Color) cellTable {
	// Colour assignments:
	//
	//   cellDD (both dark):  " "  bg=darkColour  (background shows)
	//   cellLL (both light): "█"  fg=lightColour  (full block covers cell)
	//   cellDL (top dark, bottom light): "▄"  fg=lightColour bg=darkColour
	//   cellLD (top light, bottom dark): "▀"  fg=lightColour bg=darkColour
	//
	// The light style renders full blocks with matching fg and bg so the
	// cell is uniformly light even if the font has block-drawing gaps.

	dark := lipgloss.NewStyle().Background(darkColour)
	light := lipgloss.NewStyle().
		Foreground(lightColour).
		Background(lightColour)
	mixed := lipgloss.NewStyle().
		Foreground(lightColour).
		Background(darkColour)

	return cellTable{
		cellLL: light.Render(string(fullBlock)),
		cellLD: mixed.Render(string(upperHalf)),
		cellDL: mixed.Render(string(lowerHalf)),
		cellDD: dark.Render(string(emptyBlock)),
	}
}

func buildChunkTable(ct cellTable) chunkTable {
	var t chunkTable
	for b := range 256 {
		c := unpackFour(byte(b))
		t[b] = ct[c[0]] + ct[c[1]] + ct[c[2]] + ct[c[3]]
	}
	return t
}

// codeTables bundles the cell and chunk lookup tables so they can be
// passed to [codeTables.emitRow] as a single value. The chunk table is
// derived from the cell table; see [newCodeTables].
type codeTables struct {
	cells  cellTable
	chunks chunkTable
}

// newCodeTables builds the per-frame lookup tables from the existing
// cell table.
func newCodeTables(cells cellTable) codeTables {
	return codeTables{cells: cells, chunks: buildChunkTable(cells)}
}

// emitRow appends one row of packed cells to w. Full 4-cell chunks are
// emitted from the chunk lookup table; the final partial chunk (1-3
// cells) falls back to the cell table. row is always long enough to
// cover fullChunks + (remainder > 0) bytes by [encoder.Export]'s
// padding invariant.
//
// w is an [io.StringWriter] so the table code doesn't need to know
// about the concrete buffer type. The in-tree caller passes a
// [*bytes.Buffer], whose WriteString cannot fail — hence the ignored
// return values.
func (t codeTables) emitRow(w io.StringWriter, row []byte, fullChunks, remainder int) {
	for i := range fullChunks {
		_, _ = w.WriteString(t.chunks[row[i]])
	}
	if remainder == 0 {
		return
	}
	c := unpackFour(row[fullChunks])
	for i := range remainder {
		_, _ = w.WriteString(t.cells[c[i]])
	}
}
