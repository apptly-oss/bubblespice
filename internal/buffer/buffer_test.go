package buffer

import (
	"fmt"
	"io"
	"testing"

	"darvaza.org/core"
)

// failingWriter is an [io.Writer] that always fails with
// [io.ErrShortWrite]. Used to exercise error-propagation paths.
type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) { return 0, io.ErrShortWrite }

// ── NewLazyBuffer capacity (table-driven) ─────────────────────────────

var _ core.TestCase = capacityTestCase{}

type capacityTestCase struct {
	name     string
	input    string
	capacity int
}

func newCapacityTestCase(name string, capacity int, input string) capacityTestCase {
	return capacityTestCase{
		name:     name,
		input:    input,
		capacity: capacity,
	}
}

func (tc capacityTestCase) Name() string { return tc.name }

func (tc capacityTestCase) Test(t *testing.T) {
	t.Helper()
	buf := NewLazyBuffer(tc.capacity)
	core.AssertMustNotNil(t, buf, "buffer")
	buf.WriteStrings(tc.input)
	core.AssertEqual(t, tc.input, buf.String(), "content")
}

func TestNewLazyBuffer(t *testing.T) {
	cases := []capacityTestCase{
		newCapacityTestCase("positive capacity", 100, "test"),
		newCapacityTestCase("zero capacity", 0, "zero"),
		newCapacityTestCase("negative capacity", -10, "negative"),
	}
	core.RunTestCases(t, cases)
}

// ── LazyBuffer scenarios ──────────────────────────────────────────────

func TestLazyBuffer(t *testing.T) {
	t.Run("method chaining", runTestMethodChaining)
	t.Run("reset", runTestReset)
	t.Run("write strings", runTestWriteStrings)
	t.Run("write bytes", runTestWriteBytes)
	t.Run("write runes", runTestWriteRunes)
	t.Run("print methods", runTestPrintMethods)
	t.Run("print spacing", runTestPrintSpacing)
	t.Run("grow", runTestGrow)
	t.Run("grow non-positive", runTestGrowNonPositive)
	t.Run("sys typed-nil", runTestSysTypedNil)
	t.Run("sys identity", runTestSysIdentity)
	t.Run("io writer", runTestIOWriter)
	t.Run("write to", runTestWriteTo)
	t.Run("write to error", runTestWriteToError)
	t.Run("bytes aliasing", runTestBytesAliasing)
	t.Run("print chain identity", runTestPrintChainIdentity)
	t.Run("new lazy buffer capacity", runTestNewLazyBufferCapacity)
	t.Run("write string singular", runTestWriteStringSingular)
	t.Run("write byte singular", runTestWriteByteSingular)
	t.Run("write rune singular", runTestWriteRuneSingular)
	t.Run("complex chaining", runTestComplexChaining)
}

func runTestMethodChaining(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	result := buf.WriteStrings("hello", " ", "world").
		WriteRunes('!').
		Printf(" number: %d", 42).
		WriteBytes([]byte(" end"))
	core.AssertSame(t, &buf, result, "chain identity")
	core.AssertEqual(t, "hello world! number: 42 end", buf.String(), "content")
}

func runTestReset(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	buf.WriteStrings("some text")
	core.AssertMustEqual(t, 9, buf.Len(), "pre-reset len")
	result := buf.Reset()
	core.AssertSame(t, &buf, result, "reset identity")
	core.AssertEqual(t, 0, buf.Len(), "post-reset len")
	core.AssertEqual(t, "", buf.String(), "post-reset string")
}

func runTestWriteStrings(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	buf.WriteStrings("a", "b", "c")
	core.AssertEqual(t, "abc", buf.String(), "content")
}

func runTestWriteBytes(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	buf.WriteBytes([]byte("hello"), []byte(" "), []byte("world"))
	core.AssertEqual(t, "hello world", buf.String(), "content")
}

func runTestWriteRunes(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	buf.WriteRunes('H', 'e', 'l', 'p', '!', '🚀')
	core.AssertEqual(t, "Help!🚀", buf.String(), "content")
}

func runTestPrintMethods(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	buf.Print("hello").
		Print(" ").
		Println("world").
		Printf("number: %d", 123)
	core.AssertEqual(t, "hello world\nnumber: 123", buf.String(), "content")
}

func runTestPrintSpacing(t *testing.T) {
	t.Helper()
	var buf LazyBuffer

	buf.Print("a", "b")
	core.AssertEqual(t, "ab", buf.String(), "Print two strings (no space)")

	buf.Reset().Print("k=", 42)
	core.AssertEqual(t, "k=42", buf.String(), "Print string+non-string (no space)")

	buf.Reset().Print(1, 2)
	core.AssertEqual(t, "1 2", buf.String(), "Print two non-strings (space)")

	buf.Reset().Println("a", "b")
	core.AssertEqual(t, "a b\n", buf.String(), "Println two strings (space + newline)")

	buf.Reset().Println(1, 2)
	core.AssertEqual(t, "1 2\n", buf.String(), "Println two non-strings (space + newline)")
}

func runTestGrow(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	result := buf.Grow(100)
	core.AssertSame(t, &buf, result, "grow identity")
	buf.WriteStrings("test")
	core.AssertEqual(t, "test", buf.String(), "content")
}

func runTestGrowNonPositive(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	core.AssertNoPanic(t, func() { buf.Grow(0) }, "grow zero")
	core.AssertNoPanic(t, func() { buf.Grow(-10) }, "grow negative")
}

func runTestSysTypedNil(t *testing.T) {
	t.Helper()
	var nilBuf *LazyBuffer
	core.AssertNil(t, nilBuf.Sys(), "nil receiver Sys()")
}

func runTestSysIdentity(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	sys := buf.Sys()
	core.AssertMustNotNil(t, sys, "sys")
	_, _ = sys.WriteString("via sys")
	core.AssertEqual(t, "via sys", buf.String(), "shared storage")
}

func runTestIOWriter(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	n, err := buf.Write([]byte("test"))
	core.AssertNoError(t, err, "write")
	core.AssertEqual(t, 4, n, "bytes written")
	core.AssertEqual(t, 4, buf.Len(), "len")
	core.AssertSliceEqual(t, []byte("test"), buf.Bytes(), "bytes")
}

func runTestWriteTo(t *testing.T) {
	t.Helper()
	var src, dst LazyBuffer
	src.WriteStrings("payload")
	n, err := src.WriteTo(&dst)
	core.AssertNoError(t, err, "writeTo")
	core.AssertEqual(t, int64(7), n, "bytes copied")
	core.AssertEqual(t, "payload", dst.String(), "dst content")
	core.AssertEqual(t, 0, src.Len(), "src drained")
}

func runTestWriteToError(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	buf.WriteStrings("payload")
	n, err := buf.WriteTo(failingWriter{})
	core.AssertError(t, err, "writeTo error propagated")
	core.AssertEqual(t, int64(0), n, "bytes reported")
}

func runTestBytesAliasing(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	buf.Grow(32)
	buf.WriteStrings("hello")
	b := buf.Bytes()
	core.AssertMustEqual(t, "hello", string(b), "initial content")

	// Bytes() aliases internal storage; an in-capacity rewrite
	// mutates the backing array, which the previously-returned
	// slice observes.
	buf.Reset()
	buf.WriteStrings("WORLD")
	core.AssertEqual(t, "WORLD", string(b), "aliased slice sees new writes")
}

func runTestPrintChainIdentity(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	core.AssertSame(t, &buf, buf.Print("x"), "Print identity")
	core.AssertSame(t, &buf, buf.Println("y"), "Println identity")
	core.AssertSame(t, &buf, buf.Printf("%s", "z"), "Printf identity")
}

func runTestNewLazyBufferCapacity(t *testing.T) {
	t.Helper()
	buf := NewLazyBuffer(100)
	core.AssertMustNotNil(t, buf, "buffer")
	core.AssertTrue(t, buf.Sys().Cap() >= 100, "capacity pre-allocated")
}

func runTestWriteStringSingular(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	n, err := buf.WriteString("hello")
	core.AssertNoError(t, err, "writeString")
	core.AssertEqual(t, 5, n, "bytes written")
	core.AssertEqual(t, "hello", buf.String(), "content")
}

func runTestWriteByteSingular(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	err := buf.WriteByte('Z')
	core.AssertNoError(t, err, "writeByte")
	core.AssertEqual(t, "Z", buf.String(), "content")
}

func runTestWriteRuneSingular(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	n, err := buf.WriteRune('🚀')
	core.AssertNoError(t, err, "writeRune")
	core.AssertEqual(t, 4, n, "utf-8 bytes")
	core.AssertEqual(t, "🚀", buf.String(), "content")
}

func runTestComplexChaining(t *testing.T) {
	t.Helper()
	var buf LazyBuffer
	buf.Grow(100).
		WriteStrings(`{"name":"`).
		WriteStrings("John Doe").
		WriteStrings(`","age":`).
		Printf("%d", 30).
		WriteStrings(`,"active":`).
		Print(true).
		WriteRunes('}')
	core.AssertEqual(t, `{"name":"John Doe","age":30,"active":true}`, buf.String(), "json")

	buf.Reset().WriteStrings("new content")
	core.AssertEqual(t, "new content", buf.String(), "after reset")
}

// ── Benchmarks ────────────────────────────────────────────────────────

func BenchmarkLazyBuffer_Chaining(b *testing.B) {
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		var buf LazyBuffer
		buf.WriteStrings("hello", " ", "world").
			Printf(" %d", i).
			WriteRunes('!')
		i++
	}
}

func BenchmarkLazyBuffer_JSON(b *testing.B) {
	b.ReportAllocs()
	i := 0
	for b.Loop() {
		var buf LazyBuffer
		buf.WriteStrings(`{"id":"`).
			Printf("%d", i).
			WriteStrings(`","name":"item_`).
			Printf("%d", i).
			WriteStrings(`","value":`).
			Printf("%d", i*10).
			WriteRunes('}')
		i++
	}
}

// ── Examples ──────────────────────────────────────────────────────────

func ExampleLazyBuffer_chaining() {
	var buf LazyBuffer

	buf.WriteStrings("Hello", " ").
		Print("World").
		Printf(" - %d", 2024)

	_, _ = fmt.Println(buf.String())
	// Output: Hello World - 2024
}

func ExampleLazyBuffer_json() {
	var buf LazyBuffer

	buf.WriteStrings(`{"name":"`).
		WriteStrings("Alice").
		WriteStrings(`","score":`).
		Printf("%d", 95).
		WriteRunes('}')

	_, _ = fmt.Println(buf.String())
	// Output: {"name":"Alice","score":95}
}
