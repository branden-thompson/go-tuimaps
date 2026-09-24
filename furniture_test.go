package tuimaps_test

// furniture_test.go — v0.2.0 L3.3-L3.5 (L-8.3): an image never erases the
// map's furniture, at no colour and at sixteen colours.

import (
	"image/color"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// rainOver is a two-frame loop over a box, every pixel the given colour.
func rainOver(t *testing.T, west, south, east, north float64, c color.Color) tuimaps.Overlay {
	t.Helper()
	var frames []tuimaps.LoopFrame
	for _, min := range []int{-5, 0} {
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(time.Duration(min) * time.Minute), PNG: solidPNG(t, 16, 16, c)})
	}
	return tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: west, South: south, East: east, North: north,
		Projection: tuimaps.PlateCarree, Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 60}}, Exact: true}, noon)
}

// furnitureScene draws specimen 29's furniture with no tiles at all, so that
// everything on the frame is the overlays' or the frame's own: an alert with
// its outline, hatch and label, a named place, the stale word, the frame
// time, the no-tiles notice, the scale and the credit. The loop over the
// west half is heavy rain, or clear.
func furnitureScene(t *testing.T, depth tuimaps.Depth, rain bool) []string {
	t.Helper()
	m, err := tuimaps.New(tuimaps.WithSize(69, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.ColourDepth(depth)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -90, Lat: 35}))
	must(t, m.Zoom(4))
	mustSet(t, m, warningAt("alerts", -97, 30, -83, 40))
	if _, err := m.AddPlace(tuimaps.Place{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: -92, Lat: 36}}); err != nil {
		t.Fatal(err)
	}
	c := color.Color(color.NRGBA{A: 0})
	if rain {
		c = color.NRGBA{R: 200, A: 255}
	}
	mustSet(t, m, rainOver(t, -110, 20, -90, 50, c))
	settle(t, m)
	f, err := m.Render(tuimaps.Size{Cols: 69, Rows: 12}, noon.Add(2*time.Hour)) // the alert, kept an hour, is stale
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, len(f.Lines))
	for i, line := range f.Lines {
		out[i] = plainText(line)
	}
	return out
}

// TestAnImageNeverErasesFurniture is L3.3 and L3.4 at no colour, and L3.5 at
// sixteen colours: every cell drawn over a clear loop is drawn the same over
// heavy rain, except that rain may take a hatch cell (inside rain the shade
// fills the cell; the outline and its digit carry the severity, L-8.3). The
// hatch outside the rain is still there, and the rain is drawn.
func TestAnImageNeverErasesFurniture(t *testing.T) {
	for _, depth := range []tuimaps.Depth{tuimaps.NoColour, tuimaps.Colours16} {
		clear, rain := furnitureScene(t, depth, false), furnitureScene(t, depth, true)
		kept, shaded, hatch, took := 0, 0, 0, 0
		for row := range clear {
			a, b := []rune(clear[row]), []rune(rain[row])
			if len(a) != len(b) {
				t.Fatalf("depth %v row %d: %d cells and %d", depth, row, len(a), len(b))
			}
			for col := range a {
				if strings.ContainsRune("░▒▓", b[col]) {
					shaded++
				}
				if strings.ContainsRune("╱╲╳", b[col]) {
					hatch++
				}
				if a[col] == '⠀' || a[col] == ' ' {
					continue
				}
				if strings.ContainsRune("╱╲╳", a[col]) && strings.ContainsRune("░▒▓", b[col]) {
					took++
					continue // rain takes a hatch cell
				}
				if a[col] != b[col] {
					t.Errorf("depth %v, row %d col %d: %q over a clear loop, %q over rain", depth, row, col, a[col], b[col])
					continue
				}
				kept++
			}
		}
		if shaded == 0 {
			t.Errorf("depth %v: no rain was drawn, so this proves nothing", depth)
		}
		if kept < 40 {
			t.Errorf("depth %v: only %d furniture cells to keep, so this proves little", depth, kept)
		}
		if depth == tuimaps.NoColour && hatch == 0 {
			t.Errorf("depth %v: the hatch outside the rain is gone", depth)
		}
		if depth == tuimaps.NoColour && took == 0 {
			t.Errorf("depth %v: inside the rain the hatch kept its cells; the shade fills them (D-77)", depth)
		}
	}
}
