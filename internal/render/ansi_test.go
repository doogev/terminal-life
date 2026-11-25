package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/doogev/terminal-life/internal/world"
)

func TestFlushPlainHasNoEscapeSequences(t *testing.T) {
	f := NewFrame(6, 3)
	for i := range f.Chars {
		f.Chars[i] = '#'
		f.Colors[i] = 40
	}
	var b bytes.Buffer
	if err := f.FlushPlain(&b); err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(b.String(), 0x1b) {
		t.Fatalf("plain output contains ESC: %q", b.String())
	}
	if got := strings.Count(b.String(), "\n"); got != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", got, b.String())
	}
}

func TestFlushContainsCursorHomeAndColour(t *testing.T) {
	f := NewFrame(3, 1)
	f.Chars[0] = '#'
	f.Colors[0] = 40
	var b bytes.Buffer
	if err := f.Flush(&b); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	if !strings.HasPrefix(out, "\x1b[H") {
		t.Fatalf("expected cursor home first, got %q", out[:8])
	}
	if !strings.Contains(out, "\x1b[38;5;40m") {
		t.Fatalf("expected colour 40 SGR, got %q", out)
	}
	if !strings.Contains(out, colorReset) {
		t.Fatalf("expected colour reset per row, got %q", out)
	}
}

func TestDrawRendersPlantsAndAgents(t *testing.T) {
	p := world.DefaultParams()
	p.W, p.H = 4, 1
	w := world.New(p, 1)
	w.Plants[w.Index(0, 0)] = p.MaxPlantEnergy
	w.Agents = append(w.Agents,
		world.Agent{X: 1, Y: 0, Kind: world.Herbivore},
		world.Agent{X: 2, Y: 0, Kind: world.Predator},
	)

	f := NewFrame(4, 1)
	Draw(f, w)

	if f.Chars[0] != '#' {
		t.Fatalf("full plant should render as '#', got %q", f.Chars[0])
	}
	if f.Chars[1] != herbRune || f.Colors[1] != herbColor {
		t.Fatalf("herbivore cell wrong: %q/%d", f.Chars[1], f.Colors[1])
	}
	if f.Chars[2] != predRune || f.Colors[2] != predColor {
		t.Fatalf("predator cell wrong: %q/%d", f.Chars[2], f.Colors[2])
	}
	if f.Chars[3] != ' ' {
		t.Fatalf("empty cell should be blank, got %q", f.Chars[3])
	}
}

func TestDrawIsReadOnly(t *testing.T) {
	p := world.DefaultParams()
	p.W, p.H = 3, 3
	w := world.New(p, 1)
	w.Plants[w.Index(1, 1)] = 0.5
	before := w.TotalPlantEnergy()

	f := NewFrame(3, 3)
	Draw(f, w)

	if after := w.TotalPlantEnergy(); after != before {
		t.Fatalf("Draw mutated the world: %v -> %v", before, after)
	}
}
