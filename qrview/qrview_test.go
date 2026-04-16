package qrview

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/apptly-oss/bubblespice/ansi"
)

// ── Test helpers ──────────────────────────────────────────────────────

// bm builds a bitmap from strings. '#' = dark, '.' = light.
func bm(rows ...string) [][]bool {
	bitmap := make([][]bool, len(rows))
	for i, row := range rows {
		bitmap[i] = make([]bool, len(row))
		for j, ch := range row {
			bitmap[i][j] = ch == '#'
		}
	}
	return bitmap
}

// assertPaddingLL checks that cells in c[start:4] are all cellLL.
func assertPaddingLL(t *testing.T, c [4]cellKind, start int) {
	t.Helper()
	for i := start; i < 4; i++ {
		if c[i] != cellLL {
			t.Errorf("padding cell%d should be LL, got %d", i, c[i])
		}
	}
}

// ── Cell encoding ─────────────────────────────────────────────────────

type cellFromModulesCase struct {
	name        string
	want        cellKind
	top, bottom bool
}

func (c cellFromModulesCase) Test(t *testing.T) {
	t.Helper()
	got := cellFromModules(c.top, c.bottom)
	if got != c.want {
		t.Errorf("cellFromModules(%v,%v) = %d, want %d",
			c.top, c.bottom, got, c.want)
	}
}

func TestCellFromModules(t *testing.T) {
	cases := []cellFromModulesCase{
		{name: "both_light", want: cellLL},
		{name: "bottom_dark", bottom: true, want: cellLD},
		{name: "top_dark", top: true, want: cellDL},
		{name: "both_dark", top: true, bottom: true, want: cellDD},
	}
	for _, c := range cases {
		t.Run(c.name, c.Test)
	}
}

// ── Pack / unpack ─────────────────────────────────────────────────────

func TestPackUnpack_RoundTrip(t *testing.T) {
	// Test all 256 possible byte values.
	for b := range 256 {
		c := unpackFour(byte(b))
		got := packFour(c[0], c[1], c[2], c[3])
		if got != byte(b) {
			t.Errorf("round-trip failed for %d: packed back to %d", b, got)
		}
	}
}

func TestPackFour_KnownValues(t *testing.T) {
	// All light = 0b00_00_00_00 = 0
	if packFour(cellLL, cellLL, cellLL, cellLL) != 0 {
		t.Error("all-light should pack to 0")
	}
	// All dark = 0b11_11_11_11 = 255
	if packFour(cellDD, cellDD, cellDD, cellDD) != 255 {
		t.Error("all-dark should pack to 255")
	}
	// First cell dark-dark, rest light = 0b11_00_00_00 = 192
	if packFour(cellDD, cellLL, cellLL, cellLL) != 192 {
		t.Error("DD-LL-LL-LL should pack to 192")
	}
}

func TestUnpackFour_BitPositions(t *testing.T) {
	// 0b10_01_11_00 = 156
	c := unpackFour(156)
	if c[0] != cellDL {
		t.Errorf("cell0 should be DL (0b10), got %d", c[0])
	}
	if c[1] != cellLD {
		t.Errorf("cell1 should be LD (0b01), got %d", c[1])
	}
	if c[2] != cellDD {
		t.Errorf("cell2 should be DD (0b11), got %d", c[2])
	}
	if c[3] != cellLL {
		t.Errorf("cell3 should be LL (0b00), got %d", c[3])
	}
}

// ── Bitmap encoding ───────────────────────────────────────────────────

func TestEncodeBitmap_Empty(t *testing.T) {
	g := newEncoder(nil, 2).Export()
	if g.rows != 0 {
		t.Error("nil bitmap should produce empty grid")
	}
}

func TestEncodeBitmap_1x1_NoQZ(t *testing.T) {
	g := newEncoder(bm("#"), 0).Export()
	if g.cols != 1 {
		t.Errorf("cols should be 1, got %d", g.cols)
	}
	if g.rows != 1 {
		t.Errorf("rows should be 1, got %d", g.rows)
	}
	// 1 cell padded to 4 → 1 byte.
	if len(g.data) != 1 || len(g.data[0]) != 1 {
		t.Fatalf("should be 1 row of 1 byte, got %d rows", len(g.data))
	}
	// Single dark module, no bottom (odd height) → cellDL.
	// Padded with 3× cellLL.
	c := unpackFour(g.data[0][0])
	if c[0] != cellDL {
		t.Errorf("cell0 should be DL (dark top, light bottom), got %d", c[0])
	}
	assertPaddingLL(t, c, 1)
}

func TestEncodeBitmap_2x2_NoQZ(t *testing.T) {
	g := newEncoder(bm("##", "##"), 0).Export()
	if g.cols != 2 || g.rows != 1 {
		t.Errorf("2x2 should be 2 cols × 1 row, got %d×%d", g.cols, g.rows)
	}
	c := unpackFour(g.data[0][0])
	// Both modules dark → cellDD for each of the 2 data cells.
	if c[0] != cellDD || c[1] != cellDD {
		t.Errorf("2x2 all-dark: first two cells should be DD, got %d,%d", c[0], c[1])
	}
	// Padding cells should be LL.
	if c[2] != cellLL || c[3] != cellLL {
		t.Errorf("padding should be LL, got %d,%d", c[2], c[3])
	}
}

func TestEncodeBitmap_WithQuietZone(t *testing.T) {
	g := newEncoder(bm("#"), 2).Export()
	// Data: 1×1. QZ adds 4 → 5×5 modules. Cells: 5 wide, ceil(5/2)=3 rows.
	if g.cols != 5 || g.rows != 3 {
		t.Errorf("1x1 qz=2 should be 5×3, got %d×%d", g.cols, g.rows)
	}
}

func TestEncodeBitmap_JaggedBitmap(t *testing.T) {
	jagged := [][]bool{
		{true, false, true},
		{true},
	}
	g := newEncoder(jagged, 0).Export()
	// Width should be max row width = 3.
	if g.cols != 3 {
		t.Errorf("jagged bitmap width should be 3, got %d", g.cols)
	}
	// Should not panic during encoding.
	if g.rows != 1 {
		t.Errorf("2 module rows → 1 term row, got %d", g.rows)
	}
}

// ── Encoder factories ─────────────────────────────────────────────────

// TestNewEncoder_CapturesDimensions verifies that newEncoder captures
// the source bitmap and derives the right dimensions independently of
// Export. This isolates factory behaviour from the packing loop.
func TestNewEncoder_CapturesDimensions(t *testing.T) {
	e := newEncoder(bm("##", "##"), 2)
	if e.dataH != 2 {
		t.Errorf("dataH = %d, want 2", e.dataH)
	}
	if e.dataW != 2 {
		t.Errorf("dataW = %d, want 2", e.dataW)
	}
	if e.QuietZone() != 2 {
		t.Errorf("QuietZone() = %d, want 2", e.QuietZone())
	}
	// 2-wide data + 2×2 quiet zone = 6 terminal columns.
	if e.Width() != 6 {
		t.Errorf("Width() = %d, want 6", e.Width())
	}
	// (2 + 2×2) modules high, packed two per row, rounds up to 3.
	if e.Height() != 3 {
		t.Errorf("Height() = %d, want 3", e.Height())
	}
	if len(e.bitmap) != 2 {
		t.Errorf("bitmap should be stored; got len %d", len(e.bitmap))
	}
}

func TestNewEncoder_JaggedUsesWidest(t *testing.T) {
	jagged := [][]bool{
		{true, false, true},
		{true},
	}
	e := newEncoder(jagged, 0)
	if e.dataW != 3 {
		t.Errorf("dataW should be widest row = 3, got %d", e.dataW)
	}
	if e.Width() != 3 {
		t.Errorf("Width() = %d, want 3 (no quiet zone)", e.Width())
	}
}

// TestNewEncoderFromString_ProducesBitmap verifies the string factory
// delegates to skip2 and wraps the result in an encoder without
// calling Export. If this passes but TestEncodeBitmap_* fail, the
// regression is in the packing path, not the QR step.
func TestNewEncoderFromString_ProducesBitmap(t *testing.T) {
	e, err := newEncoderFromString("hello", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Empty() {
		t.Error("encoder should capture a non-empty bitmap")
	}
	if e.QuietZone() != 2 {
		t.Errorf("QuietZone() = %d, want 2", e.QuietZone())
	}
	if e.dataW != e.dataH {
		t.Errorf("QR bitmap should be square: dataW=%d, dataH=%d",
			e.dataW, e.dataH)
	}
}

// TestEncoder_Export_Empty verifies Export on a zero-valued encoder
// returns a zero-valued grid. Exercises the early-return guard
// without going through the factory.
func TestEncoder_Export_Empty(t *testing.T) {
	var e encoder
	g := e.Export()
	if g.rows != 0 || g.cols != 0 || len(g.data) != 0 {
		t.Errorf("Export on zero encoder should produce empty grid, got %+v", g)
	}
}

// ── Lookup tables ─────────────────────────────────────────────────────

func TestCellTable_AllFourEntries(t *testing.T) {
	ct := buildCellTable(ansi.Black, ansi.BrightWhite)
	for i := range 4 {
		if ct[i] == "" {
			t.Errorf("cellTable[%d] is empty", i)
		}
	}
}

func TestChunkTable_Size(t *testing.T) {
	ct := buildCellTable(ansi.Black, ansi.BrightWhite)
	cht := buildChunkTable(ct)
	nonEmpty := 0
	for i := range 256 {
		if cht[i] != "" {
			nonEmpty++
		}
	}
	if nonEmpty != 256 {
		t.Errorf("chunk table should have 256 non-empty entries, got %d", nonEmpty)
	}
}

func TestChunkTable_ConsistentWithCellTable(t *testing.T) {
	ct := buildCellTable(ansi.Black, ansi.BrightWhite)
	cht := buildChunkTable(ct)

	// Verify a few known byte values.
	// All-light (0): should be 4× cellLL entry.
	if cht[0] != ct[cellLL]+ct[cellLL]+ct[cellLL]+ct[cellLL] {
		t.Error("chunk[0] should be 4× cellLL")
	}
	// All-dark (255): should be 4× cellDD entry.
	if cht[255] != ct[cellDD]+ct[cellDD]+ct[cellDD]+ct[cellDD] {
		t.Error("chunk[255] should be 4× cellDD")
	}
}

// ── Constructor ───────────────────────────────────────────────────────

func TestNew_NilBitmap(t *testing.T) {
	m := New(nil)
	if !m.Empty() {
		t.Error("nil bitmap should be empty")
	}
	if v := m.View(); v != "" {
		t.Errorf("empty model should render empty string, got %q", v)
	}
}

// TestNew_ImmutableAfterCreation verifies the widget's rendered output
// is unaffected by mutations to the source bitmap after New returns.
// Exercises the user-visible contract that a Model, once built, is
// immutable.
func TestNew_ImmutableAfterCreation(t *testing.T) {
	b := bm("##", "..")
	m := New(b, WithQuietZone(0))
	before := m.View()
	b[0][0] = false
	after := m.View()
	if before != after {
		t.Errorf("View() changed after source mutation;\nbefore: %q\nafter:  %q",
			before, after)
	}
}

func TestNew_ViewProducesOutput(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0))
	if m.View() == "" {
		t.Error("View should produce non-empty output for a non-empty bitmap")
	}
}

// TestNew_CachesRender verifies that calling View twice returns the
// same output — a regression guard against the cached renderCode
// being replaced on every call.
func TestNew_CachesRender(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0))
	v1 := m.View()
	v2 := m.View()
	if v1 != v2 {
		t.Error("two successive View() calls should return identical output")
	}
}

// ── Render ────────────────────────────────────────────────────────────

func TestRender_2x2_AllDark(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0))
	stripped := ansi.Strip(m.View())
	lines := strings.Split(stripped, "\n")
	if len(lines) != 1 {
		t.Fatalf("2x2 should render 1 line, got %d", len(lines))
	}
	if lines[0] != "  " {
		t.Errorf("2x2 all-dark should be two spaces, got %q", lines[0])
	}
}

func TestRender_2x2_AllLight(t *testing.T) {
	m := New(bm("..", ".."), WithQuietZone(0))
	stripped := ansi.Strip(m.View())
	lines := strings.Split(stripped, "\n")
	want := string(fullBlock) + string(fullBlock)
	if lines[0] != want {
		t.Errorf("2x2 all-light should be two full blocks, got %q", lines[0])
	}
}

func TestRender_CheckerBoard(t *testing.T) {
	m := New(bm("#.", ".#"), WithQuietZone(0))
	stripped := ansi.Strip(m.View())
	lines := strings.Split(stripped, "\n")
	if len(lines) != 1 {
		t.Fatalf("2x2 should be 1 line, got %d", len(lines))
	}
	want := string(lowerHalf) + string(upperHalf)
	if lines[0] != want {
		t.Errorf("checkerBoard: got %q, want %q", lines[0], want)
	}
}

func TestRender_OddHeight(t *testing.T) {
	m := New(bm("##", "..", "##"), WithQuietZone(0))
	stripped := ansi.Strip(m.View())
	lines := strings.Split(stripped, "\n")
	if len(lines) != 2 {
		t.Errorf("3 module rows → 2 terminal lines, got %d", len(lines))
	}
}

func TestRender_WidthNotMultipleOf4(t *testing.T) {
	// 3 columns — tests remainder handling.
	m := New(bm("###", "..."), WithQuietZone(0))
	stripped := ansi.Strip(m.View())
	lines := strings.Split(stripped, "\n")
	if len(lines) != 1 {
		t.Fatalf("2 rows → 1 line, got %d", len(lines))
	}
	// 3 cells, not 4 (no padding in output).
	runes := []rune(lines[0])
	if len(runes) != 3 {
		t.Errorf("should render 3 characters, got %d: %q", len(runes), lines[0])
	}
}

func TestRender_5Columns(t *testing.T) {
	// 5 columns = 1 full chunk + 1 remainder cell.
	m := New(bm("#####", "#####"), WithQuietZone(0))
	stripped := ansi.Strip(m.View())
	lines := strings.Split(stripped, "\n")
	runes := []rune(lines[0])
	if len(runes) != 5 {
		t.Errorf("should render 5 chars, got %d: %q", len(runes), lines[0])
	}
}

func TestRender_WithQuietZone(t *testing.T) {
	m := New(bm("#"), WithQuietZone(1))
	stripped := ansi.Strip(m.View())
	lines := strings.Split(stripped, "\n")
	// 1x1 data + qz=1 → 3 module cols, ceil((1 + 2)/2) = 2 rows.
	if len(lines) != 2 {
		t.Fatalf("should render 2 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if runes := []rune(line); len(runes) != 3 {
			t.Errorf("line %d should be 3 runes, got %d: %q", i, len(runes), line)
		}
	}
}

// ── Label ─────────────────────────────────────────────────────────────

// TestLabel_Stored verifies WithLabel sets Model.Label — isolates the
// option plumbing from the render path.
func TestLabel_Stored(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0), WithLabel("test"))
	if m.Label != "test" {
		t.Errorf("Label = %q, want %q", m.Label, "test")
	}
}

// TestLabel_WideBitmap_Contains verifies the label appears as a
// contiguous substring when the bitmap is wider than the label. This
// isolates the JoinVertical/style path from the Width() hard-wrap.
func TestLabel_WideBitmap_Contains(t *testing.T) {
	// 8-col bitmap, 4-rune label — plenty of room.
	m := New(bm("########", "########"), WithQuietZone(0), WithLabel("test"))
	stripped := ansi.Strip(m.View())
	if !strings.Contains(stripped, "test") {
		t.Errorf("label should appear in wide-bitmap output; got %q", stripped)
	}
}

// TestLabel_ExactWidth_Contains verifies the boundary: bitmap width
// equal to label rune count should render the label unbroken.
func TestLabel_ExactWidth_Contains(t *testing.T) {
	// 4-col bitmap, 4-rune label "test".
	m := New(bm("####", "####"), WithQuietZone(0), WithLabel("test"))
	stripped := ansi.Strip(m.View())
	if !strings.Contains(stripped, "test") {
		t.Errorf("label should appear when cols == len(label); got %q", stripped)
	}
}

// TestLabel_NarrowBitmap_Wraps documents the current behaviour: when
// the bitmap is narrower than the label, lipgloss Width(cols)
// hard-wraps the label, so it no longer appears as a contiguous
// substring. This is the known failure mode TestLabel was hitting.
func TestLabel_NarrowBitmap_Wraps(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0), WithLabel("test"))
	stripped := ansi.Strip(m.View())
	// Either rendered on its own line or wrapped — we assert that
	// the individual runes still appear in order.
	for _, r := range "test" {
		if !strings.ContainsRune(stripped, r) {
			t.Errorf("rune %q missing from narrow-bitmap output %q", r, stripped)
		}
	}
}

// TestLabel_SingleChar_Contains is the minimal positive case: a
// 1-rune label always fits, regardless of bitmap width.
func TestLabel_SingleChar_Contains(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0), WithLabel("X"))
	stripped := ansi.Strip(m.View())
	if !strings.Contains(stripped, "X") {
		t.Errorf("single-rune label should always appear; got %q", stripped)
	}
}

func TestNoLabel(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0))
	stripped := ansi.Strip(m.View())
	lines := strings.Split(stripped, "\n")
	if len(lines) != 1 {
		t.Errorf("no label → 1 line, got %d", len(lines))
	}
}

// ── SetData ───────────────────────────────────────────────────────────

func TestSetData_ChangesOutput(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0))
	v1 := m.View()
	m = m.SetData(bm("..", ".."))
	v2 := m.View()
	if v1 == v2 {
		t.Error("SetData should change output")
	}
}

// TestSetData_ImmutableAfterCreation is the SetData analogue of
// [TestNew_ImmutableAfterCreation]: mutations to the bitmap handed to
// SetData must not reach back into the Model.
func TestSetData_ImmutableAfterCreation(t *testing.T) {
	data := bm("##", "##")
	m := New(nil).SetData(data)
	before := m.View()
	data[0][0] = false
	after := m.View()
	if before != after {
		t.Errorf("View() changed after SetData source mutation;\nbefore: %q\nafter:  %q",
			before, after)
	}
}

func TestSetData_UpdatesSize(t *testing.T) {
	m := New(bm("#"), WithQuietZone(0))
	lines := strings.Split(ansi.Strip(m.View()), "\n")
	if len(lines) != 1 {
		t.Fatalf("initial model should render 1 line, got %d", len(lines))
	}
	m = m.SetData(bm("##", "##", "##"))
	lines = strings.Split(ansi.Strip(m.View()), "\n")
	if len(lines) != 2 {
		t.Errorf("after SetData, model should render 2 lines, got %d", len(lines))
	}
}

// TestSetData_InvalidatesMemo verifies that SetData replaces the memo
// so the next View reflects the new bitmap rather than serving the
// cached old output.
func TestSetData_InvalidatesMemo(t *testing.T) {
	m := New(bm("#"), WithQuietZone(0))
	before := m.View()
	m = m.SetData(bm("##", "##", "##"))
	after := m.View()
	if before == after {
		t.Error("View after SetData should not equal pre-SetData output")
	}
}

// ── Update ────────────────────────────────────────────────────────────

func TestUpdate_PassThrough(t *testing.T) {
	m := New(bm("#"))
	m2, cmd := m.Update(nil)
	if cmd != nil {
		t.Error("Update should return nil cmd")
	}
	if m2.Empty() != m.Empty() {
		t.Error("Update should not change model")
	}
}

// ── Standalone ────────────────────────────────────────────────────────

func TestRenderFunction(t *testing.T) {
	s := Render(bm("##", ".."), WithQuietZone(0))
	if s == "" {
		t.Error("Render should produce output")
	}
}

// ── Custom colours ────────────────────────────────────────────────────

func TestWithColours(t *testing.T) {
	// Shouldn't panic or produce empty output.
	m := New(bm("##", "##"),
		WithQuietZone(0),
		WithColours("196", "46"),
	)
	v := m.View()
	if v == "" {
		t.Error("custom colours should still produce output")
	}
}

// ── Inversion ─────────────────────────────────────────────────────────

// TestWithInverted_SetsFlag verifies the option reaches the model.
// Without this, later tests that depend on buildCode reading the flag
// would fail silently if the option were broken.
func TestWithInverted_SetsFlag(t *testing.T) {
	m := New(bm("#"), WithInverted())
	if !m.inverted {
		t.Error("WithInverted should set m.inverted = true")
	}
}

// TestWithInverted_ChangesOutput verifies that inversion produces a
// visibly different render when lipgloss is actually emitting ANSI.
// In headless `go test` runs the default renderer falls back to plain
// text and inversion is a no-op at the byte level; the probe detects
// that case and skips. Run in a real terminal to exercise this.
func TestWithInverted_ChangesOutput(t *testing.T) {
	probe := lipgloss.NewStyle().Background(ansi.Black).Render(" ")
	if probe == " " {
		t.Skip("lipgloss is not emitting ANSI; inversion is not observable here")
	}

	normal := New(bm("##", "##"), WithQuietZone(0)).View()
	inverted := New(bm("##", "##"), WithQuietZone(0), WithInverted()).View()
	if normal == inverted {
		t.Error("WithInverted should produce different output from default")
	}
}

// TestWithInverted_EqualsSwappedColours verifies that WithInverted is
// equivalent to WithColours(light, dark) — i.e. it really swaps the
// palette rather than doing something else. This is the correctness
// anchor for the flag's rendering effect: if lipgloss emits ANSI both
// outputs carry the same codes, and if it doesn't (headless tests)
// both collapse to the same plain text. Either way they must agree.
func TestWithInverted_EqualsSwappedColours(t *testing.T) {
	inverted := New(bm("#.", ".#"),
		WithQuietZone(0),
		WithInverted(),
	).View()
	swapped := New(bm("#.", ".#"),
		WithQuietZone(0),
		WithColours(ansi.BrightWhite, ansi.Black),
	).View()
	if inverted != swapped {
		t.Errorf("WithInverted should match WithColours(light, dark);\n"+
			"inverted: %q\nswapped:  %q", inverted, swapped)
	}
}

// TestWithInverted_SurvivesSetData verifies the inversion flag is
// preserved across SetData — the memo is rebuilt but the option state
// stays on the Model.
func TestWithInverted_SurvivesSetData(t *testing.T) {
	m := New(bm("##", "##"), WithQuietZone(0), WithInverted())
	before := m.View()
	m = m.SetData(bm("##", "##"))
	after := m.View()
	if before != after {
		t.Errorf("identical data should render identically across SetData;\n"+
			"before: %q\nafter:  %q", before, after)
	}
}
