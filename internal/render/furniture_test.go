package render

import (
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// TestStaleMark is plan task 09.16 (FR-32): when an overlay's data is no
// longer current the frame says so, in a word and not a colour alone, and
// says nothing when the data is current.
func TestStaleMark(t *testing.T) {
	_, in := gulf(t)
	fresh := render(t, in)
	if strings.Contains(plain(strings.Join(fresh.Lines, "\n")), staleMark) {
		t.Error("a frame whose data is current carries the stale mark")
	}
	in.Stale = true
	marked := render(t, in)
	first := plain(marked.Lines[0])
	if !strings.Contains(first, staleMark) {
		t.Errorf("the stale mark is not on the frame's first row:\n%s", first)
	}
	if !strings.HasSuffix(strings.TrimRight(first, "⠀"), staleMark) {
		t.Errorf("the stale mark is not at the right end of the first row:\n%s", first)
	}
	// It is drawn in its own token, so that it is not read as a place name.
	if colours(marked.Lines[0], staleMark) == colours(marked.Lines[0], "⠀") {
		t.Error("the stale mark is drawn in the colour of whatever is beside it")
	}
	// With no colour at all it is still a word.
	in.Depth = NoColour
	if !strings.Contains(plain(render(t, in).Lines[0]), staleMark) {
		t.Error("the stale mark is gone when there is no colour to draw it in")
	}
}

// TestFooterLineOffByDefault is the rest of 09.16 (P-57): the footer is the
// host's own line of text, drawn inside the map only when the host asks for
// it, and it is off by default.
func TestFooterLineOffByDefault(t *testing.T) {
	_, in := gulf(t)
	f := render(t, in)
	for i, line := range f.Lines {
		if strings.Contains(plain(line), "center") {
			t.Fatalf("row %d carries a footer nobody asked for:\n%s", i, plain(line))
		}
	}
	in.Footer = textsafe.Const("center: 26.000, -84.000 | zoom: 3")
	on := render(t, in)
	row := plain(on.Lines[len(on.Lines)-2])
	if !strings.HasPrefix(row, "center: 26.000, -84.000 | zoom: 3") {
		t.Errorf("the footer is not at the left of the row above the last:\n%s", row)
	}
	// The last row is still the scale mark's and the credit's.
	if last := plain(on.Lines[len(on.Lines)-1]); !strings.Contains(last, "OpenFreeMap") {
		t.Errorf("the footer took the credit's row:\n%s", last)
	}
	// A footer too wide for the map is cut, never wrapped, and the row is
	// still exactly as wide as the map.
	in.Footer = textsafe.Clean(strings.Repeat("wide ", 80))
	cut := render(t, in)
	if got := textsafe.Width(textsafe.Clean(plain(cut.Lines[len(cut.Lines)-2]))); got != in.View.Cols {
		t.Errorf("the row with a long footer is %d cells wide, want %d", got, in.View.Cols)
	}
}

// TestFurnitureKeepsItsCells: the furniture is the frame's own, and nothing
// the map draws takes its cells (FR-12, L2 Render step 9).
func TestFurnitureKeepsItsCells(t *testing.T) {
	v, in := gulf(t)
	in.Stale = true
	in.Scale = true
	at, _, _ := quarter(t, v)
	top, err := v.FromDot(float64(v.Cols*2-2), 1) // the stale mark's own corner
	if err != nil {
		t.Fatal(err)
	}
	in.Markers = []Marker{{At: top, Shape: MarkerGlyph, Text: "◉", Ink: uint8(colour.Marker)}}
	_ = at
	f := render(t, in)
	if !strings.Contains(plain(f.Lines[0]), staleMark) {
		t.Errorf("a marker took the stale mark's cells:\n%s", plain(f.Lines[0]))
	}
}
