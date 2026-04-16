package qrview

import (
	// cspell:words qrcode
	qrcode "github.com/skip2/go-qrcode"
)

// packedGrid is the bitmap (with quiet zone) encoded into
// terminal-cell rows. Each row is a []byte where each byte holds 4
// cells. Rows are right-padded with cellLL (light) to a multiple of 4
// so every byte is fully valid.
//
// cols is the actual terminal-column width before the right padding;
// len(data[y]) == roundUp(cols, 4) / 4. rows == len(data).
type packedGrid struct {
	data [][]byte
	cols int
	rows int
}

// encoder holds the invariants needed while packing a QR bitmap into
// terminal cells. Dimensions are cached in module space so the data
// flow is linear:
//
//		data ──(+2*quietZone)──▶ total (module) ──(pack half-blocks)──▶ cells
//
//	  - dataH/dataW — raw bitmap dimensions, quiet zone excluded.
//	  - totalH/totalW — module-space dimensions with quiet zone included
//	    (totalX = dataX + 2*quietZone). totalW also equals the terminal
//	    column count (one module per column); the terminal row count is
//	    ceil(totalH/2), computed by [encoder.Height] on demand.
type encoder struct {
	bitmap    [][]bool
	dataH     int
	dataW     int
	totalH    int
	totalW    int
	quietZone int
}

// newEncoder builds an encoder from a raw bitmap. The bitmap is
// normalised — deep-copied and padded to a rectangular shape (the
// widest row wins) — so the encoder is insulated from later caller
// mutations and [encoder.module] only needs a single bounds check.
func newEncoder(bitmap [][]bool, quietZone int) encoder {
	if quietZone < 0 {
		quietZone = 0
	}
	normalised := normaliseBitmap(bitmap)
	dataH := len(normalised)
	var dataW int
	if dataH > 0 {
		dataW = len(normalised[0])
	}
	return encoder{
		bitmap:    normalised,
		dataH:     dataH,
		dataW:     dataW,
		totalH:    dataH + 2*quietZone,
		totalW:    dataW + 2*quietZone,
		quietZone: quietZone,
	}
}

// newEncoderFromString builds an [encoder] from a string by QR-encoding
// it via skip2/go-qrcode at Medium error correction. This is the only
// place that imports the encoding library; the rest of qrview is
// encoder-agnostic. [newEncoder] normalises and deep-copies the
// library's bitmap, so the encoder is insulated from it after return.
func newEncoderFromString(data string, quietZone int) (encoder, error) {
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return encoder{}, err
	}
	// Disable the library's own border — we render quiet zone ourselves.
	qr.DisableBorder = true
	return newEncoder(qr.Bitmap(), quietZone), nil
}

// WithQuietZone returns a copy of e with the quiet zone updated.
// Negative values clamp to 0. Raw dimensions are preserved; only the
// totals are recomputed, so this is O(1).
func (e encoder) WithQuietZone(n int) encoder {
	if n < 0 {
		n = 0
	}
	e.quietZone = n
	e.totalH = e.dataH + 2*n
	e.totalW = e.dataW + 2*n
	return e
}

// WithData returns a copy of e with the bitmap replaced. The existing
// quiet zone is preserved; [newEncoder] normalises and deep-copies
// the bitmap internally.
func (e encoder) WithData(bitmap [][]bool) encoder {
	return newEncoder(bitmap, e.quietZone)
}

// QuietZone returns the current quiet-zone width in modules.
func (e encoder) QuietZone() int { return e.quietZone }

// Width returns the rendered width in terminal columns (= total
// module columns, quiet zone included).
func (e encoder) Width() int { return e.totalW }

// Height returns the rendered height in terminal rows — two vertical
// modules pack into each row, quiet zone included, rounded up.
func (e encoder) Height() int { return (e.totalH + 1) / 2 }

// Empty reports whether the encoder has no data.
func (e encoder) Empty() bool { return e.dataH == 0 }

// Size returns the rendered dimensions in terminal columns and rows,
// including the quiet zone. An empty encoder returns (0, 0).
func (e encoder) Size() (cols, rows int) {
	if e.Empty() {
		return 0, 0
	}
	return e.Width(), e.Height()
}

// Export packs the bitmap into terminal cells and returns the final
// packedGrid. An encoder over an empty bitmap yields a zero-value
// packedGrid.
func (e encoder) Export() packedGrid {
	if e.Empty() {
		return packedGrid{}
	}

	packedBytes := roundUp(e.Width(), 4) / 4
	data := make([][]byte, e.Height())
	for j := range data {
		modRow0 := j*2 - e.quietZone
		modRow1 := modRow0 + 1

		row := make([]byte, packedBytes)
		for i := range packedBytes {
			c := e.four(i*4, modRow0, modRow1)
			row[i] = packFour(c[0], c[1], c[2], c[3])
		}
		data[j] = row
	}

	return packedGrid{
		data: data,
		cols: e.Width(),
		rows: len(data),
	}
}

// module returns the QR module at (row, col) in bitmap-space, treating
// anything outside the data rectangle as light (quiet-zone padding).
// Single bounds check is safe because [newEncoder] pads every row to
// dataW.
func (e encoder) module(row, col int) bool {
	if row < 0 || col < 0 || row >= e.dataH || col >= e.dataW {
		return false
	}
	return e.bitmap[row][col]
}

// four packs the 4 cells starting at terminal-column cx, for the two
// module rows modRow0/modRow1 stacked into each cell.
func (e encoder) four(cx, modRow0, modRow1 int) [4]cellKind {
	var c [4]cellKind
	for i := range 4 {
		col := cx + i
		if col >= e.totalW {
			c[i] = cellLL // right padding beyond totalW
			continue
		}
		modCol := col - e.quietZone
		top := e.module(modRow0, modCol)
		bot := e.module(modRow1, modCol)
		c[i] = cellFromModules(top, bot)
	}
	return c
}

// normaliseBitmap returns a rectangular deep copy of src: every row
// has the same width — the longest row in src — with missing cells
// filled as light (false). An empty src returns nil.
func normaliseBitmap(src [][]bool) [][]bool {
	if len(src) == 0 {
		return nil
	}
	w := 0
	for _, row := range src {
		if l := len(row); l > w {
			w = l
		}
	}
	dst := make([][]bool, len(src))
	for i, row := range src {
		dst[i] = make([]bool, w)
		copy(dst[i], row)
	}
	return dst
}

func roundUp(n, m int) int {
	return m * ((n + m - 1) / m)
}
