// Package buffer provides a [bytes.Buffer] wrapper whose chainable
// write helpers discard errors, letting callers build up output
// without error-handling boilerplate. Interface-compliant methods
// ([io.Writer], [io.WriterTo], [io.StringWriter], [io.ByteWriter])
// keep their standard signatures.
//
// Not safe for concurrent use, same as [bytes.Buffer].
package buffer

import (
	"bytes"
	"fmt"
	"io"
)

// LazyBuffer is a [bytes.Buffer] whose chainable helpers return
// *LazyBuffer instead of (int, error), so writes compose without
// intermediate error checks. The zero value is ready to use.
type LazyBuffer bytes.Buffer

var (
	_ io.Writer       = (*LazyBuffer)(nil)
	_ io.WriterTo     = (*LazyBuffer)(nil)
	_ io.StringWriter = (*LazyBuffer)(nil)
	_ io.ByteWriter   = (*LazyBuffer)(nil)
)

// Sys returns the underlying [bytes.Buffer].
// Safe to call on a nil receiver; returns nil.
func (buf *LazyBuffer) Sys() *bytes.Buffer {
	if buf == nil {
		return nil
	}
	return (*bytes.Buffer)(buf)
}

// Len reports the number of bytes currently stored.
func (buf *LazyBuffer) Len() int { return buf.Sys().Len() }

// String returns the stored data as string.
func (buf *LazyBuffer) String() string { return buf.Sys().String() }

// Bytes returns the stored data as a byte slice. The slice aliases
// internal storage and is invalidated by the next write; see
// [bytes.Buffer.Bytes].
func (buf *LazyBuffer) Bytes() []byte { return buf.Sys().Bytes() }

// Write implements the [io.Writer] interface.
func (buf *LazyBuffer) Write(b []byte) (int, error) { return buf.Sys().Write(b) }

// WriteTo implements the [io.WriterTo] interface.
func (buf *LazyBuffer) WriteTo(out io.Writer) (int64, error) { return buf.Sys().WriteTo(out) }

// WriteString implements the [io.StringWriter] interface.
// For the chainable form, use [LazyBuffer.WriteStrings].
func (buf *LazyBuffer) WriteString(s string) (int, error) { return buf.Sys().WriteString(s) }

// WriteByte implements the [io.ByteWriter] interface.
func (buf *LazyBuffer) WriteByte(c byte) error { return buf.Sys().WriteByte(c) }

// WriteRune appends the UTF-8 encoding of r.
// For the chainable form, use [LazyBuffer.WriteRunes].
func (buf *LazyBuffer) WriteRune(r rune) (int, error) { return buf.Sys().WriteRune(r) }

// Print formats operands with [fmt.Print] semantics and appends
// the result. Returns the buffer for method chaining.
func (buf *LazyBuffer) Print(a ...any) *LazyBuffer {
	_, _ = fmt.Fprint(buf.Sys(), a...)
	return buf
}

// Println formats operands with [fmt.Println] semantics and appends
// the result. Returns the buffer for method chaining.
func (buf *LazyBuffer) Println(a ...any) *LazyBuffer {
	_, _ = fmt.Fprintln(buf.Sys(), a...)
	return buf
}

// Printf formats operands with [fmt.Printf] semantics and appends
// the result. Returns the buffer for method chaining.
func (buf *LazyBuffer) Printf(format string, a ...any) *LazyBuffer {
	_, _ = fmt.Fprintf(buf.Sys(), format, a...)
	return buf
}

// WriteRunes appends the given runes as UTF-8 characters to the buffer.
// Returns the buffer for method chaining.
func (buf *LazyBuffer) WriteRunes(runes ...rune) *LazyBuffer {
	for _, r := range runes {
		_, _ = buf.Sys().WriteRune(r)
	}
	return buf
}

// WriteBytes writes the given byte slices to the buffer.
// Returns the buffer for method chaining.
func (buf *LazyBuffer) WriteBytes(s ...[]byte) *LazyBuffer {
	for _, b := range s {
		_, _ = buf.Sys().Write(b)
	}
	return buf
}

// WriteStrings writes the given strings to the buffer.
// Returns the buffer for method chaining.
func (buf *LazyBuffer) WriteStrings(ss ...string) *LazyBuffer {
	for _, s := range ss {
		_, _ = buf.Sys().WriteString(s)
	}
	return buf
}

// Grow guarantees space for n more bytes in the buffer.
// Negative or zero n is a no-op. Returns the buffer for method chaining.
func (buf *LazyBuffer) Grow(n int) *LazyBuffer {
	if n > 0 {
		buf.Sys().Grow(n)
	}
	return buf
}

// Reset clears the buffer for reuse.
// Returns the buffer for method chaining.
func (buf *LazyBuffer) Reset() *LazyBuffer {
	buf.Sys().Reset()
	return buf
}

// NewLazyBuffer creates a new LazyBuffer with pre-allocated capacity.
// Negative or zero capacity yields an empty buffer.
func NewLazyBuffer(capacity int) *LazyBuffer {
	return (&LazyBuffer{}).Grow(capacity)
}
