package tuimaps

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/assets"
)

// warmEast is a temperature field over the map, cold in the west and warm in
// the east, so that its bands run down the frame and every row crosses them.
func warmEast() Grid {
	const cols, rows = 24, 24
	g := Grid{West: -125, South: 24, East: -66, North: 50, Cols: cols, Rows: rows}
	g.Values = make([]float64, 0, cols*rows)
	// The field warms from west to east, so its bands run north-south and
	// every row crosses them. A contour is labelled where the band changes
	// along a row, so a field banded the other way is never labelled at all.
	for range rows {
		for c := range cols {
			g.Values = append(g.Values, -20.0+float64(c)*2.5)
		}
	}
	return g
}

// number finds a contour's value written on the map.
var anyValue = regexp.MustCompile(`-?\d+`)

// TestAFieldsContoursCarryTheirValues is the defect HUM LEAD found in the
// acceptance sitting (task 14.15, scenario 4) and is **the whole of NFR-15 for
// a field**: with no colour a temperature field is drawn as contour lines, and
// a contour that carries no value says where a band changes but not to what.
// HUM LEAD read both no-colour frames as "cannot tell".
//
// The renderer has always been able to write these values - D-35 designed them
// and `valueLabel` draws them - but nothing ever filled the labels in, so every
// field ever drawn without colour had bare lines.
func TestAFieldsContoursCarryTheirValues(t *testing.T) {
	m, err := New(WithSize(120, 40), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	m.ColourDepth(NoColour)
	// Point the map at the field, so its bands fill the frame rather than a
	// corner of a map of the world.
	if err := m.Recentre(LonLat{Lon: -96, Lat: 37}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(3); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(TemperatureGrid("temp", warmEast(), Celsius, drawnAt)); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	frame, err := m.Render(Size{Cols: 120, Rows: 40}, drawnAt)
	if err != nil {
		t.Fatal(err)
	}
	// The last row is the map's furniture - the scale bar and the credit - and
	// the scale bar carries a number of its own, so the body is what is asked.
	body := strings.Join(frame.Lines[:len(frame.Lines)-1], "\n")
	body = escapes.ReplaceAllString(body, "")
	if !anyValue.MatchString(body) {
		t.Fatalf("no contour on the no-colour field carries a value; the body holds no number at all:\n%s", body)
	}
}

// escapes are the colour sequences, which carry numbers of their own.
var escapes = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")
