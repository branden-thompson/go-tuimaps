package render

import (
	"slices"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// namesShown are the basemap's names a frame shows, sorted, each once.
func namesShown(r *Renderer, f Frame) []string {
	text := ""
	for _, line := range f.Lines {
		text += plain(line) + "\n"
	}
	var out []string
	for _, l := range r.painter.Labels() {
		if name := l.Name.String(); name != "" && strings.Contains(text, name) && !slices.Contains(out, name) {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return out
}

// kansas is a view of the plains with the basemap's names on.
func kansas(t *testing.T) Input {
	t.Helper()
	v := project.View{Centre: project.LonLat{Lon: -95, Lat: 38}, Zoom: 3.5, Cols: 120, Rows: 40}
	return Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Labels: true, OverlaysVersion: 1}
}

// framed draws an input and says the names it shows and its text.
func framed(t *testing.T, in Input) ([]string, string) {
	t.Helper()
	r, f := drawn(t, in)
	text := ""
	for _, line := range f.Lines {
		text += plain(line) + "\n"
	}
	return namesShown(r, f), text
}

// AN ALERT OUTSIDE THE MOMENT STILL HOLDS ITS WORD'S ROOM (L-28, watchpost
// D-200): its shape is reserved, not drawn - no tint, no outline, no word -
// and the basemap's names are placed as they are with the alert drawn, so a
// loop playing past its hours never reshuffles them (watchpost UAT-2 U2-46).
func TestAReservedAlertHoldsItsWordsRoom(t *testing.T) {
	area := scene.Shape{Kind: scene.ShapeArea, Role: uint8(colour.AlertSevereOutline), Label: "Tornado Warning", Rings: [][]scene.Vertex{lonLatBox(t, -99, 36, -91, 40)}}
	bare, _ := framed(t, kansas(t))
	shown := kansas(t)
	shown.Shapes = []scene.Shape{area}
	withAlert, text := framed(t, shown)
	if !strings.Contains(text, "Tornado Warning") {
		t.Fatalf("the alert's word is not drawn:\n%s", text)
	}
	if slices.Equal(bare, withAlert) {
		t.Fatalf("the alert's word displaces no name here (%v): the test proves nothing", bare)
	}
	reserved := kansas(t)
	reserved.Reserved = []scene.Shape{area}
	r, f := drawn(t, reserved)
	held := namesShown(r, f)
	if !slices.Equal(held, withAlert) {
		t.Errorf("with the alert reserved the names are %v; with it drawn %v", held, withAlert)
	}
	for i, c := range r.grid.cells {
		if c.area == uint8(colour.AlertSevereTint) || c.ink == uint8(colour.AlertSevereOutline) {
			t.Fatalf("cell %d carries the reserved alert's colour: a reserved shape is never drawn", i)
		}
	}
	if text := strings.Join(f.Lines, "\n"); strings.Contains(plain(text), "Tornado") {
		t.Error("a reserved alert's word is drawn")
	}
	if len(f.Dropped) != 0 {
		t.Errorf("a reserved alert's word is told dropped: %+v", f.Dropped)
	}
}

// A RESERVED OUTLINE'S DIGITS HOLD THEIR CELLS (L-28, D-65): the severity
// digits written along a drawn alert's outline keep names off those cells,
// so its reserved form keeps them off too - and writes no digit.
func TestAReservedOutlineHoldsItsDigits(t *testing.T) {
	area := scene.Shape{Kind: scene.ShapeArea, Role: uint8(colour.AlertSevereOutline), Mark: "3", Rings: [][]scene.Vertex{lonLatBox(t, -104, 33, -86, 43)}}
	bare, _ := framed(t, kansas(t))
	shown := kansas(t)
	shown.Shapes = []scene.Shape{area}
	withAlert, _ := framed(t, shown)
	if slices.Equal(bare, withAlert) {
		t.Fatalf("the outline's digits displace no name here (%v): the test proves nothing", bare)
	}
	reserved := kansas(t)
	reserved.Reserved = []scene.Shape{area}
	r, f := drawn(t, reserved)
	if held := namesShown(r, f); !slices.Equal(held, withAlert) {
		t.Errorf("with the outline reserved the names are %v; with it drawn %v", held, withAlert)
	}
	for i, c := range r.grid.cells {
		if c.text == "3" && c.ink == uint8(colour.AlertSevereOutline) {
			t.Fatalf("cell %d carries a reserved outline's digit", i)
		}
	}
}

// A FIELD OR IMAGE NOT DRAWN THIS FRAME STILL COUNTS AGAINST THE NAMES (L-28):
// a loop's gap, or a layer outside the moment, leaves the name budget as it
// is under the overlay, so the names do not grow and shrink as the loop plays.
func TestTheNameBudgetHoldsAcrossALoop(t *testing.T) {
	bare, _ := framed(t, kansas(t))
	under := kansas(t)
	under.Fields = []scene.Field{bands(40, 30)}
	covered, _ := framed(t, under)
	if slices.Equal(bare, covered) {
		t.Fatalf("an overlay changes no name here (%v): the test proves nothing", bare)
	}
	for name, in := range map[string]func(*Input){
		"a loop on a gap":            func(in *Input) { in.ImageHeld = true },
		"a layer outside the moment": func(in *Input) { in.Covered = true },
	} {
		frame := kansas(t)
		in(&frame)
		if got, _ := framed(t, frame); !slices.Equal(got, covered) {
			t.Errorf("%s: the names are %v; under a drawn overlay %v", name, got, covered)
		}
	}
}

// A frame reserving a shape is not the frame that did not: the reuse check
// sees the reserved shapes and the cover change (contract, section 5).
func TestReservingIsAChangeOfFrame(t *testing.T) {
	area := scene.Shape{Kind: scene.ShapeArea, Role: uint8(colour.AlertSevereOutline), Label: "Tornado Warning", Rings: [][]scene.Vertex{lonLatBox(t, -99, 36, -91, 40)}}
	in := kansas(t)
	r, _ := drawn(t, in)
	for name, change := range map[string]func(*Input){
		"reserved": func(in *Input) { in.Reserved = []scene.Shape{area} },
		"covered":  func(in *Input) { in.Covered = true },
	} {
		next := in
		change(&next)
		if r.sameOverlays(next) {
			t.Errorf("%s: the frame would be reused", name)
		}
	}
}

// ONLY AN ALERT'S WORDS ARE HELD (L-28): every other overlay's words come
// after the names (L-23), so one outside the moment holds nothing from them.
func TestAReservedPointHoldsNothing(t *testing.T) {
	bare, _ := framed(t, kansas(t))
	quake := scene.Shape{Kind: scene.ShapeArea, Role: uint8(colour.Track), Label: "Magnitude Five Quake Here", Rings: [][]scene.Vertex{lonLatBox(t, -99, 36, -91, 40)}}
	shown := kansas(t)
	shown.Reserved = []scene.Shape{quake}
	if got, _ := framed(t, shown); !slices.Equal(got, bare) {
		t.Errorf("a reserved shape that is no alert moved the names: %v; bare %v", got, bare)
	}
}
