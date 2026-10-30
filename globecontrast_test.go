package tuimaps_test

import (
	"strconv"
	"strings"
	"testing"
	"unicode"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/colour"
)

// paletteColour is the colour a 256-colour entry shows: the cube and the grey
// ramp, the entries the library draws with (16 to 255).
func paletteColour(i int) (colour.RGB, bool) {
	levels := [6]uint8{0, 95, 135, 175, 215, 255}
	switch {
	case i >= 16 && i <= 231:
		i -= 16
		return colour.RGB{R: levels[i/36], G: levels[i/6%6], B: levels[i%6]}, true
	case i >= 232 && i <= 255:
		v := uint8(8 + 10*(i-232))
		return colour.RGB{R: v, G: v, B: v}, true
	}
	return colour.RGB{}, false
}

// cellColours reads a frame line's cells: each character with the foreground
// and background in force where it is drawn, the ground where no background
// is set, and whether its foreground is known.
type cell struct {
	r        rune
	fg, bg   colour.RGB
	fgKnown  bool
	position int
}

func cellColours(t *testing.T, line string, ground colour.RGB) []cell {
	t.Helper()
	var out []cell
	fg, bg, fgKnown := colour.RGB{}, ground, false
	rest := line
	for len(rest) > 0 {
		if loc := sgr.FindStringSubmatchIndex(rest); loc != nil && loc[0] == 0 {
			params := strings.Split(rest[loc[2]:loc[3]], ";")
			for i := 0; i < len(params); i++ {
				n, _ := strconv.Atoi(params[i])
				switch {
				case params[i] == "" || n == 0:
					fg, bg, fgKnown = colour.RGB{}, ground, false
				case n == 39:
					fg, fgKnown = colour.RGB{}, false
				case n == 49:
					bg = ground
				case (n == 38 || n == 48) && i+1 < len(params):
					var c colour.RGB
					known := false
					switch params[i+1] {
					case "2":
						if i+4 < len(params) {
							r, _ := strconv.Atoi(params[i+2])
							g, _ := strconv.Atoi(params[i+3])
							b, _ := strconv.Atoi(params[i+4])
							c, known = colour.RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, true
							i += 4
						}
					case "5":
						if i+2 < len(params) {
							k, _ := strconv.Atoi(params[i+2])
							c, known = paletteColour(k)
							i += 2
						}
					}
					if !known {
						t.Fatalf("a colour the check cannot read in %q", rest[loc[0]:loc[1]])
					}
					if n == 38 {
						fg, fgKnown = c, true
					} else {
						bg = c
					}
				}
			}
			rest = rest[loc[1]:]
			continue
		}
		r, size := []rune(rest)[0], len(string([]rune(rest)[0]))
		out = append(out, cell{r: r, fg: fg, bg: bg, fgKnown: fgKnown, position: len(out)})
		rest = rest[size:]
	}
	return out
}

// TestEveryGlobeCellKeepsItsContrast is plan task P7.0 (FR-16): every
// whole-globe specimen, with the night and its terminator over it, at every
// depth whose colours are fixed - truecolor and 256 - on both grounds, draws
// every cell's foreground at its rule's contrast on the cell's own
// background: 3:1 for line work and glyphs, 4.5:1 for text.
func TestEveryGlobeCellKeepsItsContrast(t *testing.T) {
	night := tuimaps.NightOverlay("night", noon)
	for _, s := range specimens {
		for _, d := range specimenDepths[:2] {
			for _, g := range specimenGrounds {
				m := world(t, specimenSize.Cols, specimenSize.Rows)
				must(t, m.FitWorld())
				m.ColourDepth(d.depth)
				must(t, m.Ground(g.behind))
				mustSet(t, m, s.overlay())
				mustSet(t, m, night)
				settle(t, m)
				f, err := m.Render(specimenSize, noon)
				if err != nil {
					t.Fatal(err)
				}
				ground := colour.RGB{R: g.behind.R, G: g.behind.G, B: g.behind.B}
				checked, failed := 0, 0
				for row, line := range f.Lines {
					for _, c := range cellColours(t, line, ground) {
						if c.r == ' ' || c.r == '⠀' {
							continue
						}
						if !c.fgKnown {
							t.Errorf("%s %s %s: row %d, cell %d %q has no colour of its own", s.name, d.name, g.name, row, c.position, c.r)
							continue
						}
						need := colour.LineContrast
						if unicode.IsLetter(c.r) || unicode.IsDigit(c.r) {
							need = colour.TextContrast
						}
						checked++
						if got := colour.Contrast(c.fg, c.bg); got < need-0.005 {
							failed++
							if failed <= 3 {
								t.Errorf("%s %s %s: row %d, cell %d %q at %.2f:1 on its background, under %.1f:1 (%v on %v)", s.name, d.name, g.name, row, c.position, c.r, got, need, c.fg, c.bg)
							}
						}
					}
				}
				if checked < 1000 {
					t.Errorf("%s %s %s: %d cells checked; a whole globe draws thousands", s.name, d.name, g.name, checked)
				}
				if failed > 3 {
					t.Errorf("%s %s %s: %d cells of %d under their contrast", s.name, d.name, g.name, failed, checked)
				}
			}
		}
	}
}
