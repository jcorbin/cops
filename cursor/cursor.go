// Package cursor models cursor state and has methods to modify color,
// position, and visibility, appending the corresponding ANSI escape sequences
// to a write buffer.
package cursor

import (
	"image"
	"image/color"
	"strconv"

	"github.com/kriskowal/cops/vtcolor"
)

// Cursor models the known or unknown states of a cursor.
type Cursor struct {
	// Position is the position of the cursor.
	// Negative values indicate that the X or Y position is not known,
	// so the next position change must be relative to the beginning of the
	// same line or possibly the origin.
	Position image.Point
	// Foreground is the foreground color for subsequent text.
	// Transparent indicates that the color is unknown, so the next text must
	// be preceded by an SGR (set graphics) ANSI sequence to set it.
	Foreground color.RGBA
	// Foreground is the foreground color for subsequent text.
	// Transparent indicates that the color is unknown, so the next text must
	// be preceded by an SGR (set graphics) ANSI sequence to set it.
	Background color.RGBA
}

var (
	// Lost indicates that the cursor position is unknown.
	Lost = image.Point{-1, -1}

	// Start is a cursor state that makes no assumptions about the cursor's
	// position or colors, necessitating a seek from origin and explicit color
	// settings for the next text.
	Start = Cursor{
		Position:   Lost,
		Foreground: vtcolor.Transparent,
		Background: vtcolor.Transparent,
	}

	// Reset is a cursor state indicating that the cursor is at the origin
	// and that the foreground color is white (7), background black (0).
	// This is the state cur.Reset() returns to, and the state for which
	// cur.Reset() will append nothing to the buffer.
	Reset = Cursor{
		Position:   image.ZP,
		Foreground: vtcolor.Colors[7],
		Background: vtcolor.Colors[0],
	}
)

// Hide hides the cursor.
func (c Cursor) Hide(buf []byte) ([]byte, Cursor) {
	return append(buf, "\033[?25l"...), c
}

// Show reveals the cursor.
func (c Cursor) Show(buf []byte) ([]byte, Cursor) {
	return append(buf, "\033[?25h"...), c
}

// Clear erases the whole display.
func (c Cursor) Clear(buf []byte) ([]byte, Cursor) {
	// Clear implicitly invalidates the cursor position since its behavior is
	// inconsistent across terminal implementations.
	return append(buf, "\033[2J"...), Cursor{
		Position:   Lost,
		Foreground: c.Foreground,
		Background: c.Background,
	}
}

// Reset returns the terminal to default white on black colors.
func (c Cursor) Reset(buf []byte) ([]byte, Cursor) {
	if c.Foreground == vtcolor.Colors[7] && c.Background == vtcolor.Colors[0] {
		return buf, c
	}
	return append(buf, "\033[m"...), Cursor{
		Position:   c.Position,
		Foreground: vtcolor.Colors[7],
		Background: vtcolor.Colors[0],
	}
}

// Home seeks the cursor to the origin, using display absolute coordinates.
func (c Cursor) Home(buf []byte) ([]byte, Cursor) {
	c.Position = image.ZP
	return append(buf, "\033[H"...), c
}

// Go moves the cursor to another position, prefering to use relative motion,
// using line relative if the column is unknown, using display origin relative
// only if the line is also unknown. If the column is unknown, use "\r" to seek
// to column 0 of the same line.
func (c Cursor) Go(buf []byte, to image.Point) ([]byte, Cursor) {
	if c.Position == Lost {
		// If the cursor position is completely unknown, move relative to
		// screen origin. This mode must be avoided to render relative to
		// cursor position inline with a scrolling log, by setting the cursor
		// position relative to an arbitrary origin before rendering.
		buf = append(buf, "\033["...)
		buf = append(buf, strconv.Itoa(to.Y)...)
		buf = append(buf, ";"...)
		buf = append(buf, strconv.Itoa(to.X)...)
		buf = append(buf, "H"...)
		c.Position = to
		return buf, c
	}

	if c.Position.X == -1 {
		// If only horizontal position is unknown, return to first column and
		// march forward.
		// Rendering a non-ASCII cell of unknown or indeterminite width may
		// invalidate the column number.
		// For example, a skin tone emoji may or may not render as a single
		// column glyph.
		buf = append(buf, "\r"...)
		c.Position.X = 0
		// Continue...
	}

	if to.X == 0 && to.Y == c.Position.Y+1 {
		buf, c = c.Reset(buf)
		buf = append(buf, "\r\n"...)
		c.Position.X = 0
		c.Position.Y++
	} else if to.X == 0 && c.Position.X != 0 {
		buf, c = c.Reset(buf)
		buf = append(buf, "\r"...)
		c.Position.X = 0

		// In addition to scrolling back to the first column generally, this
		// has the effect of resetting the column if writing a multi-byte
		// string invalidates the cursor's horizontal position.
		// For example, a skin tone emoji may or may not render as a single
		// column glyph.
	}

	if to.Y < c.Position.Y {
		buf = append(buf, "\033["...)
		buf = append(buf, strconv.Itoa(c.Position.Y-to.Y)...)
		buf = append(buf, "A"...)
	} else if to.Y > c.Position.Y {
		buf = append(buf, "\033["...)
		buf = append(buf, strconv.Itoa(to.Y-c.Position.Y)...)
		buf = append(buf, "B"...)
	}
	if to.X < c.Position.X {
		buf = append(buf, "\033["...)
		buf = append(buf, strconv.Itoa(c.Position.X-to.X)...)
		buf = append(buf, "D"...)
	} else if to.X > c.Position.X {
		buf = append(buf, "\033["...)
		buf = append(buf, strconv.Itoa(to.X-c.Position.X)...)
		buf = append(buf, "C"...)
	}

	c.Position = to
	return buf, c
}
