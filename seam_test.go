package tuimaps_test

import (
	"context"
	"image/color"
	"regexp"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// seamMap is a map centred on the antimeridian's west side, at a zoom where
// the view reaches well past it.
func seamMap(t *testing.T, lon float64) (*tuimaps.Map, tuimaps.Size) {
	t.Helper()
	size := tuimaps.Size{Cols: 100, Rows: 30}
	m, err := tuimaps.New(tuimaps.WithSize(size.Cols, size.Rows), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	if err := m.Zoom(3); err != nil {
		t.Fatal(err)
	}
	if err := m.Recentre(tuimaps.LonLat{Lon: lon, Lat: 62}); err != nil {
		t.Fatal(err)
	}
	m.ColourDepth(tuimaps.NoColour)
	return m, size
}

// drawnHalves counts the braille cells with dots in the frame's west and
// east halves.
func drawnHalves(t *testing.T, m *tuimaps.Map, size tuimaps.Size) (west, east int, text string) {
	t.Helper()
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	f, err := m.Render(size, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	text = regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(strings.Join(f.Lines, "\n"), "")
	for _, line := range strings.Split(text, "\n")[1 : len(f.Lines)-1] {
		cells := []rune(line)
		for i, r := range cells {
			if r > 0x2800 && r <= 0x28ff {
				if i < len(cells)/2 {
					west++
				} else {
					east++
				}
			}
		}
	}
	return west, east, text
}

// TestTheWorldRepeatsAcrossTheAntimeridian is L-7.4 (D-86; watchpost UAT-2
// U2-7, D-90): a view whose east half lies past 180° draws the land there -
// Alaska's west and the Aleutians from a view centred on Chukotka - and the
// frame is filled; upstream drew nothing past the edge.
func TestTheWorldRepeatsAcrossTheAntimeridian(t *testing.T) {
	m, size := seamMap(t, 179.5) // the east half almost wholly past 180°
	west, east, text := drawnHalves(t, m, size)
	if west == 0 || east < west/4 {
		t.Errorf("past 180° the frame is empty: %d cells drawn west, %d east:\n%s", west, east, text)
	}
	named := false // a place's or a sea's name past the seam: the names repeat with the world
	rows := strings.Split(text, "\n")
	for _, l := range rows[1 : len(rows)-1] { // not the stamp's row, not the credit's
		runes := []rune(l)
		if len(runes) > size.Cols/2 && strings.ContainsAny(string(runes[size.Cols/2+2:]), "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			named = true
		}
	}
	if !named {
		t.Errorf("no name is drawn past 180°:\n%s", text)
	}
}

// TestAnOverlayAcrossTheSeamIsDrawnWhole: an alert area just east of 180°
// is drawn in the view west of it, on the side it lies - never torn between
// two copies of the world, which would streak a line across the frame.
func TestAnOverlayAcrossTheSeamIsDrawnWhole(t *testing.T) {
	m, size := seamMap(t, 175)
	ring := []tuimaps.LonLat{{Lon: -178, Lat: 58}, {Lon: -174, Lat: 58}, {Lon: -174, Lat: 61}, {Lon: -178, Lat: 61}, {Lon: -178, Lat: 58}}
	if _, err := m.Set(tuimaps.Overlay{ID: "alert/seam", Valid: time.Unix(0, 0), Keeps: 6 * time.Hour,
		Features: []tuimaps.Feature{{Kind: tuimaps.Polygon, Rings: [][]tuimaps.LonLat{ring}, Role: tuimaps.AlertSevere, Label: "Gale Warning", ID: "g"}}}); err != nil {
		t.Fatal(err)
	}
	_, _, text := drawnHalves(t, m, size)
	lines := strings.Split(text, "\n")
	found := false
	for _, l := range lines {
		if at := strings.Index(l, "Gale"); at >= 0 {
			found = true
			if col := len([]rune(l[:at])); col < size.Cols/2 {
				t.Errorf("the alert east of 180° is labelled at column %d, west of the view's centre", col)
			}
		}
	}
	if !found {
		t.Errorf("the alert past the seam is not drawn:\n%s", text)
	}
}

// TestAHostCanTakeTheStampOver is L-1.14 (D-87; watchpost D-92): with the
// stamp off, the top row carries no frame time and no stale word - the host
// shows them - and on again it does; the switch moves Changed.
func TestAHostCanTakeTheStampOver(t *testing.T) {
	size := tuimaps.Size{Cols: 80, Rows: 20}
	m, err := tuimaps.New(tuimaps.WithSize(size.Cols, size.Rows), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	noon := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	var frames []tuimaps.LoopFrame
	for i := 2; i >= 0; i-- {
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(-time.Duration(i) * 5 * time.Minute), PNG: solidPNG(t, 40, 40, color.NRGBA{R: 200, A: 255})})
	}
	if _, err := m.Set(tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: -100, South: 30, East: -90, North: 40, Projection: tuimaps.PlateCarree,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 45}}, Exact: true}, noon)); err != nil {
		t.Fatal(err)
	}
	top := func() string {
		f, err := m.Render(size, noon)
		if err != nil {
			t.Fatal(err)
		}
		return regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(f.Lines[0], "")
	}
	if !strings.Contains(top(), "12:00") {
		t.Fatalf("with the stamp on, the top row is %q", top())
	}
	was := m.Changed()
	m.ShowStamp(false)
	if m.Changed() == was {
		t.Error("the switch did not move Changed")
	}
	if strings.Contains(top(), "12:00") {
		t.Errorf("with the stamp off, the top row still says the time: %q", top())
	}
	m.ShowStamp(true)
	if !strings.Contains(top(), "12:00") {
		t.Error("on again, the stamp is not back")
	}
}
