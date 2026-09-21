package textsafe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestClosedGlyphList is plan task 02.10: every character the renderer may
// emit of its own accord is in one closed list, and each measures one cell
// under the pinned width table (NFR-8).
func TestClosedGlyphList(t *testing.T) {
	count := 0
	for r := rune(0); r <= utf8.MaxRune; r++ {
		if !InGlyphList(r) {
			continue
		}
		count++
		if w := clusterWidth(string(r)); w != 1 {
			t.Errorf("U+%04X is in the closed list and measures %d cells; the renderer's own glyphs are one cell each", r, w)
		}
		if got := Clean(string(r)).String(); got != string(r) {
			t.Errorf("U+%04X is in the closed list, yet cleaning changes it to %q", r, got)
		}
	}
	// 95 printable ASCII, 256 braille, 6 markers, 8 box-drawing, 3 hatch, 3 shades, U+FFFD.
	if want := 95 + 256 + 6 + 8 + 3 + 3 + 1; count != want {
		t.Errorf("the closed list holds %d characters, want %d", count, want)
	}
}

func TestGlyphListMembersAndStrangers(t *testing.T) {
	in := []rune{' ', 'A', '~', 0x2800, 0x28FF, 0x00B7, 0x25C9, 0x2500, 0x251C, 0x2571, 0x2572, 0x2591, 0x2593, 0xFFFD}
	out := []rune{0, '\n', 0x1b, 0x7F, 0x00E9, 0x2190, 0x2197, 0x2588, 0x2598, 0x27FF, 0x2900, 0x6771, 0x1F600}
	for _, r := range in {
		if !InGlyphList(r) {
			t.Errorf("U+%04X should be in the closed list", r)
		}
	}
	for _, r := range out {
		if InGlyphList(r) {
			t.Errorf("U+%04X is in the closed list; arrows and block quadrants come with later releases, and nothing else is the renderer's own", r)
		}
	}
}

// TestGlyphListWasSeenInATerminal ties the list to its evidence: every
// non-ASCII glyph in it, and a sample of braille, is on the terminal test
// card HUM LEAD ran (the PLAN entry checks).
func TestGlyphListWasSeenInATerminal(t *testing.T) {
	card, err := os.ReadFile(filepath.Join("..", "..", "06_docs", "02_features", "go-tuimaps", "03-architecture-design", "terminal-matrix", "test-card.txt"))
	if err != nil {
		t.Fatal(err)
	}
	braille := 0
	for r := rune(0x80); r <= 0xFFFF; r++ {
		if !InGlyphList(r) {
			continue
		}
		on := strings.ContainsRune(string(card), r)
		if r >= 0x2800 && r <= 0x28FF {
			if on {
				braille++
			}
			continue
		}
		if !on {
			t.Errorf("U+%04X is in the closed list and is not on the terminal test card", r)
		}
	}
	if braille < 8 {
		t.Errorf("only %d braille patterns are on the test card", braille)
	}
}
