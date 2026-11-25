// Package render turns world state into a character frame and writes it to a
// terminal. It is a read-only consumer of world: nothing here may mutate the
// simulation.
package render

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/doogev/terminal-life/internal/world"
)

// Frame is a double-buffer target: characters plus a 256-colour index per cell.
type Frame struct {
	Chars  []rune
	Colors []uint8
	W, H   int
}

// NewFrame allocates a cleared frame.
func NewFrame(w, h int) *Frame {
	f := &Frame{Chars: make([]rune, w*h), Colors: make([]uint8, w*h), W: w, H: h}
	f.Clear()
	return f
}

// Clear resets every cell to a black space.
func (f *Frame) Clear() {
	for i := range f.Chars {
		f.Chars[i] = ' '
		f.Colors[i] = 0
	}
}

// Plant ramp: four visible levels plus empty, with a dark-to-bright green.
var (
	plantRunes  = [5]rune{' ', '.', ':', '*', '#'}
	plantColors = [5]uint8{0, 22, 28, 34, 40}
)

const (
	herbRune   = 'o'
	herbColor  = 220
	predRune   = 'O'
	predColor  = 196
	colorReset = "\x1b[0m"

	// plantVisible is the energy below which a cell reads as bare soil.
	// Sunlight is applied uniformly to every cell (DESIGN §3), so without a
	// threshold the whole grid would light up on tick 1 and the seeded
	// patches would be indistinguishable from the background.
	plantVisible = 0.05
)

// Draw fills f from the world. Read-only: it never writes back.
func Draw(f *Frame, w *world.World) {
	f.Clear()
	max := w.P.MaxPlantEnergy
	for y := 0; y < f.H && y < w.P.H; y++ {
		for x := 0; x < f.W && x < w.P.W; x++ {
			e := w.PlantAt(x, y)
			if e < plantVisible || max <= 0 {
				continue
			}
			lvl := 1 + int(float32(len(plantRunes)-2)*e/max)
			if lvl > len(plantRunes)-1 {
				lvl = len(plantRunes) - 1
			}
			i := y*f.W + x
			f.Chars[i] = plantRunes[lvl]
			f.Colors[i] = plantColors[lvl]
		}
	}
	for _, a := range w.Agents {
		if a.X < 0 || a.Y < 0 || a.X >= f.W || a.Y >= f.H {
			continue
		}
		i := a.Y*f.W + a.X
		if a.Kind == world.Predator {
			f.Chars[i], f.Colors[i] = predRune, predColor
		} else {
			f.Chars[i], f.Colors[i] = herbRune, herbColor
		}
	}
}

// Flush writes the frame with ANSI colour and a cursor-home, so successive
// frames overwrite in place. One escape sequence per colour run.
func (f *Frame) Flush(w io.Writer) error {
	var b strings.Builder
	b.Grow(f.W*f.H + f.H*16)
	b.WriteString("\x1b[H")
	for y := 0; y < f.H; y++ {
		last := uint8(255) // impossible colour, forces the first SGR
		for x := 0; x < f.W; x++ {
			i := y*f.W + x
			if c := f.Colors[i]; c != last {
				fmt.Fprintf(&b, "\x1b[38;5;%dm", c)
				last = c
			}
			b.WriteRune(f.Chars[i])
		}
		b.WriteString(colorReset)
		if y != f.H-1 {
			b.WriteByte('\n')
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// FlushPlain writes the frame as bare text with no escape sequences, for
// piped output and tests. The no-escape guarantee is asserted in ansi_test.
func (f *Frame) FlushPlain(w io.Writer) error {
	var b strings.Builder
	b.Grow(f.W*f.H + f.H)
	for y := 0; y < f.H; y++ {
		for x := 0; x < f.W; x++ {
			b.WriteRune(f.Chars[y*f.W+x])
		}
		b.WriteByte('\n')
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// IsTTY reports whether f is a character device, i.e. an interactive terminal.
func IsTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// EnterAltScreen switches to the alternate screen and hides the cursor.
func EnterAltScreen(w io.Writer) {
	io.WriteString(w, "\x1b[?1049h\x1b[?25l\x1b[2J\x1b[H")
}

// LeaveAltScreen restores colours, cursor and the original screen. Safe to
// call more than once.
func LeaveAltScreen(w io.Writer) {
	io.WriteString(w, colorReset+"\x1b[?25h\x1b[?1049l")
}
