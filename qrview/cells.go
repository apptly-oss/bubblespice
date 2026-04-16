package qrview

// cellKind encodes the state of one terminal cell — two vertically
// stacked QR modules — in 2 bits:
//
//	00  both light     (█ — foreground fills cell)
//	01  top light, bottom dark  (▀ — upper half drawn)
//	10  top dark, bottom light  (▄ — lower half drawn)
//	11  both dark      (' ' — background shows through)
//
// Four cells pack into one byte, left-to-right in high-to-low bits:
//
//	bit  7 6   5 4   3 2   1 0
//	     cell0 cell1 cell2 cell3
type cellKind byte

const (
	cellLL cellKind = 0b00 // both light
	cellLD cellKind = 0b01 // top light, bottom dark
	cellDL cellKind = 0b10 // top dark, bottom light
	cellDD cellKind = 0b11 // both dark
)

const (
	emptyBlock = ' '
	fullBlock  = '█'
	upperHalf  = '▀'
	lowerHalf  = '▄'
)

// revive:disable-next-line:flag-parameter
func cellFromModules(top, bottom bool) cellKind {
	var c cellKind
	if top {
		c |= 0b10
	}
	if bottom {
		c |= 0b01
	}
	return c
}

// packFour packs 4 cells into a byte, left-to-right = high-to-low.
func packFour(c0, c1, c2, c3 cellKind) byte {
	return byte(c0)<<6 | byte(c1)<<4 | byte(c2)<<2 | byte(c3)
}

// unpackFour extracts 4 cells from a byte.
func unpackFour(b byte) [4]cellKind {
	return [4]cellKind{
		cellKind((b >> 6) & 0x03),
		cellKind((b >> 4) & 0x03),
		cellKind((b >> 2) & 0x03),
		cellKind((b >> 0) & 0x03),
	}
}
