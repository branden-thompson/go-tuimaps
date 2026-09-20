package tuimaps_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// framesDir is where the reference frames are kept: readiness evidence, not
// code. **They are candidates until HUM LEAD has approved them** (plan task
// 14.16, RS-15); until then this test's work is to keep them from changing
// without anyone noticing.
const framesDir = "06_docs/02_features/go-tuimaps/07-readiness/reference-frames"

// updateFrames is the environment variable that writes the frames again. A
// run that writes them proves nothing, so it is never the default.
const updateFrames = "TUIMAPS_WRITE_REFERENCE_FRAMES"

// m1Slice is the M1 scenarios this release draws: scenario 5 is wind, which
// arrives with the wind overlay (D-44).
func m1Slice() []string {
	return []string{"scenario-1", "scenario-2", "scenario-3", "scenario-4", "scenario-6", "scenario-7"}
}

// m1Sizes are the two sizes M1 is judged at: the first host's window, and
// the smallest rectangle the library supports (NFR-7).
func m1Sizes() []tuimaps.Size {
	return []tuimaps.Size{{Cols: 149, Rows: 38}, {Cols: 69, Rows: 12}}
}

// m1Depths are the two ways each frame is drawn: in colour, and with none
// at all - which is how NFR-15's reviewer session reads them (D-67).
func m1Depths() []struct {
	name  string
	depth tuimaps.Depth
} {
	return []struct {
		name  string
		depth tuimaps.Depth
	}{{"colour", tuimaps.Truecolor}, {"plain", tuimaps.NoColour}}
}

// TestReferenceFrames is plan task 14.3: every M1 scenario of the slice, at
// both sizes, in colour and with none. The frames are the evidence HUM LEAD
// judges M1a from (14.15) and the reviewer answers the M1 questions from
// (NFR-15), so what matters here is that they are exactly reproducible: the
// same tiles, the same clock, the same frame, every run.
func TestReferenceFrames(t *testing.T) {
	writing := os.Getenv(updateFrames) != ""
	for _, name := range m1Slice() {
		for _, size := range m1Sizes() {
			for _, depth := range m1Depths() {
				kept := filepath.Join(framesDir, fmt.Sprintf("%s-%dx%d-%s.txt", name, size.Cols, size.Rows, depth.name))
				drawn := scenarioFrame(t, name, size, depth.depth)
				if writing {
					if err := os.WriteFile(kept, []byte(drawn), 0o644); err != nil {
						t.Fatal(err)
					}
					continue
				}
				body, err := os.ReadFile(kept)
				if err != nil {
					t.Fatalf("%s: %v; run the tests once with %s set to write them", kept, err, updateFrames)
				}
				if string(body) != drawn {
					t.Errorf("%s is not the frame the library draws now. If the change is meant, "+
						"write the frames again with %s set - and they are HUM LEAD's to approve (14.16)", kept, updateFrames)
				}
			}
		}
	}
}

// TestReferenceFramesAreTheSameTwice: a frame that differed between two
// runs would be no evidence at all (NFR-6).
func TestReferenceFramesAreTheSameTwice(t *testing.T) {
	for _, name := range m1Slice() {
		size := tuimaps.Size{Cols: 69, Rows: 12}
		if first, again := scenarioFrame(t, name, size, tuimaps.NoColour), scenarioFrame(t, name, size, tuimaps.NoColour); first != again {
			t.Errorf("%s drew two different frames from the same data", name)
		}
	}
}

// scenarioFrame is one scenario drawn: the places and the overlays of its
// file, framed so that both are on the map, settled, and rendered on a
// fixed clock.
func scenarioFrame(t *testing.T, name string, size tuimaps.Size, depth tuimaps.Depth) string {
	t.Helper()
	m := world(t, size.Cols, size.Rows)
	m.ColourDepth(depth)
	s := scenarioNamed(t, name)
	places := make([]tuimaps.Place, 0, len(s.Places))
	for _, p := range s.Places {
		places = append(places, tuimaps.Place{Name: p.Name, At: tuimaps.LonLat{Lon: p.Lon, Lat: p.Lat}})
	}
	if _, err := m.SetPlaces(places); err != nil {
		t.Fatal(err)
	}
	putScenario(t, m, s)
	at := make([]tuimaps.LonLat, 0, len(places))
	for _, p := range places {
		at = append(at, p.At)
	}
	if err := m.FitTo(at, m.Overlays(), 2); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	frame, err := m.Render(size, noon)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(frame.Lines, "\n") + "\n"
}

// scenarioNamed reads one scenario file.
func scenarioNamed(t *testing.T, name string) scenario {
	t.Helper()
	var s scenario
	read(t, filepath.Join("06_docs", "02_features", "go-tuimaps", "02-analysis", "scenarios", name+".json"), &s)
	return s
}

// putScenario hands a scenario's overlays to a map: the same call the app
// makes, from the same files.
func putScenario(t *testing.T, m *tuimaps.Map, s scenario) {
	t.Helper()
	for _, shape := range s.Shapes {
		if _, err := m.Set(overlayOf(t, shape.ID, shape.Kind, shape.Rings)); err != nil {
			t.Fatalf("%s: %v", shape.ID, err)
		}
	}
	for _, f := range s.Fields {
		grid := tuimaps.Grid{West: f.West, South: f.South, East: f.East, North: f.North,
			Cols: f.Cols, Rows: f.Rows, Values: f.Values}
		if _, err := m.Set(tuimaps.TemperatureGrid(f.ID, grid, tuimaps.Celsius, noon)); err != nil {
			t.Fatalf("%s: %v", f.ID, err)
		}
	}
	for _, picture := range s.Pictures {
		png, err := base64.StdEncoding.DecodeString(picture.PNG)
		if err != nil {
			t.Fatalf("%s: %v", picture.ID, err)
		}
		image := tuimaps.Image{West: picture.West, South: picture.South, East: picture.East, North: picture.North,
			Projection: tuimaps.PlateCarree, PNG: png, Tolerance: picture.Tolerance}
		for _, e := range picture.Table {
			image.Table = append(image.Table, tuimaps.TableEntry{
				Colour: tuimaps.RGB{R: e.R, G: e.G, B: e.B}, Value: e.Value, Missing: e.Missing})
		}
		if _, err := m.Set(tuimaps.RadarImage(picture.ID, image, noon)); err != nil {
			t.Fatalf("%s: %v", picture.ID, err)
		}
		if _, err := m.Settle(context.Background()); err != nil {
			t.Fatal(err) // a picture is classified by a job before it can be drawn
		}
	}
}

// TestM1GuardSharedFrame is plan task 14.5: M1's own guard. Every scenario
// must be judgeable at once - the place and the hazard in one frame, at the
// smallest size a person is likely to use - because a question about where
// a place is relative to a hazard cannot be answered from a frame that
// holds only one of them (D-76).
func TestM1GuardSharedFrame(t *testing.T) {
	const cols, rows = 80, 24
	for _, name := range m1Slice() {
		s := scenarioNamed(t, name)
		m := world(t, cols, rows)
		m.ColourDepth(tuimaps.NoColour)
		places := make([]tuimaps.Place, 0, len(s.Places))
		at := make([]tuimaps.LonLat, 0, len(s.Places))
		for _, p := range s.Places {
			places = append(places, tuimaps.Place{Name: p.Name, At: tuimaps.LonLat{Lon: p.Lon, Lat: p.Lat}})
			at = append(at, tuimaps.LonLat{Lon: p.Lon, Lat: p.Lat})
		}
		if _, err := m.SetPlaces(places); err != nil {
			t.Fatal(err)
		}
		putScenario(t, m, s)
		if err := m.FitTo(at, m.Overlays(), 2); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if _, err := m.Settle(context.Background()); err != nil {
			t.Fatal(err)
		}
		frame, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon)
		if err != nil {
			t.Fatal(err)
		}
		// Every place is on the map, and so is the hazard: the frame with
		// the overlays differs from the frame without them, which is what
		// "the hazard is in this frame" means when read from the picture.
		drawn := strings.Join(frame.Lines, "\n")
		for _, id := range m.Overlays() {
			if _, err := m.Remove(id); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := m.Settle(context.Background()); err != nil {
			t.Fatal(err)
		}
		bare, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, noon)
		if err != nil {
			t.Fatal(err)
		}
		if drawn == strings.Join(bare.Lines, "\n") {
			t.Errorf("%s: fitting put the places in the frame and left the hazard out of it", name)
		}
		if len(frame.Lines) != rows {
			t.Errorf("%s: %d rows, want %d", name, len(frame.Lines), rows)
		}
	}
}
