package render

import (
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

func vertex(t *testing.T, lon, lat float64) scene.Vertex {
	t.Helper()
	x, y, err := project.ToTile(project.LonLat{Lon: lon, Lat: lat}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return scene.Vertex{X: uint32(x * (1 << 32)), Y: uint32(y * (1 << 32))}
}

func lonLatBox(t *testing.T, west, south, east, north float64) []scene.Vertex {
	return []scene.Vertex{vertex(t, west, south), vertex(t, east, south), vertex(t, east, north), vertex(t, west, north), vertex(t, west, south)}
}

// TestAlertAreaDrawn: an alert area is its tint, as the background of the
// cells it owns, and its outline, which always shows (FR-16: a tint is never
// the only edge), with its label placed before any name of the basemap's.
func TestAlertAreaDrawn(t *testing.T) {
	v := project.View{Centre: project.LonLat{Lon: -95, Lat: 38}, Zoom: 3.5, Cols: 120, Rows: 40}
	area := scene.Shape{Kind: scene.ShapeArea, Role: uint8(colour.AlertSevereOutline), Label: "Tornado Warning", Rings: [][]scene.Vertex{lonLatBox(t, -99, 36, -91, 40)}}
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Labels: true, Shapes: []scene.Shape{area}, OverlaysVersion: 1}
	r, err := NewRenderer(v.Cols, v.Rows)
	if err != nil {
		t.Fatal(err)
	}
	f, err := r.Draw(in)
	if err != nil {
		t.Fatal(err)
	}
	centre := r.grid.cells[20*120+60]
	if centre.area != uint8(colour.AlertSevereTint) {
		t.Errorf("the cell at the area's middle has area ink %d, want the severe tint's", centre.area)
	}
	if corner := r.grid.cells[2*120+2]; corner.area == uint8(colour.AlertSevereTint) {
		t.Error("a cell far from the area has its tint")
	}
	outlined := 0
	for _, c := range r.grid.cells {
		if c.ink == uint8(colour.AlertSevereOutline) && c.text == "" {
			outlined++
		}
	}
	if outlined < 40 {
		t.Errorf("%d cells carry the outline's colour; the box is some 160 cells round", outlined)
	}
	text := ""
	for _, line := range f.Lines {
		text += plain(line) + "\n"
	}
	if !strings.Contains(text, "Tornado Warning") {
		t.Errorf("the area's label is not drawn:\n%s", text)
	}
	if !strings.Contains(strings.Join(f.Lines, ""), "48;2;112;36;28") {
		t.Error("no cell has the severe tint as its background")
	}
	// Moving the area is a change the frame must show.
	moved := in
	moved.Shapes = []scene.Shape{{Kind: scene.ShapeArea, Role: uint8(colour.AlertSevereOutline), Rings: [][]scene.Vertex{lonLatBox(t, -89, 36, -85, 40)}}}
	moved.OverlaysVersion = 2
	before := r.Redraws()
	if _, err := r.Draw(moved); err != nil || r.Redraws() != before+1 {
		t.Errorf("a changed overlay did not redraw the frame: %v", err)
	}
}

// TestOverlayLabelBeforeBasemap: a place name never hides a warning's word.
func TestOverlayLabelBeforeBasemap(t *testing.T) {
	v := project.View{Centre: project.LonLat{Lon: -95, Lat: 38}, Zoom: 4, Cols: 120, Rows: 40}
	places := scene.Layer{Name: "place", Extent: 4096, Features: []scene.Feature{{Kind: scene.GeomPoint, Class: "city", Name: textsafe.Clean("Centreville"), Rank: 1, EndPart: 1}}, Coords: []int16{2048, 2048}, Parts: []uint32{2}}
	centre, _ := project.FromTile(0.5, 0.5, 0)
	_ = centre
	at := scene.TileID{Z: 4, X: 3, Y: 6}
	tileCentre, _ := project.FromTile(float64(at.X)+0.5, float64(at.Y)+0.5, at.Z)
	area := scene.Shape{Kind: scene.ShapeArea, Role: uint8(colour.AlertExtremeOutline), Label: "Flood Warning",
		Rings: [][]scene.Vertex{lonLatBox(t, tileCentre.Lon-2, tileCentre.Lat-1, tileCentre.Lon+2, tileCentre.Lat+1)}}
	view := project.View{Centre: tileCentre, Zoom: 4, Cols: 120, Rows: 40}
	in := Input{View: view, Tiles: []Drawn{{Tile: &scene.Tile{Layers: []scene.Layer{places}}, At: at, Exact: true}}, Style: style.BuiltIn(), Labels: true, Shapes: []scene.Shape{area}, OverlaysVersion: 1}
	text := ""
	for _, line := range render(t, in).Lines {
		text += plain(line) + "\n"
	}
	if !strings.Contains(text, "Flood Warning") || strings.Contains(text, "Centreville") {
		t.Errorf("the warning's word and a city's name want the same spot, and the warning's must win:\n%s", text)
	}
	_ = v
}

// TestLineAndPointShapes: a track is a line in its colour; a point is a mark.
func TestLineAndPointShapes(t *testing.T) {
	v := project.View{Centre: project.LonLat{Lon: -95, Lat: 38}, Zoom: 3.5, Cols: 120, Rows: 40}
	track := scene.Shape{Kind: scene.ShapeLine, Role: uint8(colour.Track), Rings: [][]scene.Vertex{{vertex(t, -105, 38), vertex(t, -85, 38)}}}
	point := scene.Shape{Kind: scene.ShapePoint, Role: uint8(colour.Marker), Label: "Home", Rings: [][]scene.Vertex{{vertex(t, -95, 42)}}}
	in := Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Labels: true, Shapes: []scene.Shape{track, point}, OverlaysVersion: 1}
	r, _ := NewRenderer(v.Cols, v.Rows)
	f, err := r.Draw(in)
	if err != nil {
		t.Fatal(err)
	}
	tracked := 0
	for col := range 120 {
		if r.grid.cells[20*120+col].ink == uint8(colour.Track) {
			tracked++
		}
	}
	if tracked < 60 {
		t.Errorf("%d cells of the middle row carry the track's colour", tracked)
	}
	text := ""
	for _, line := range f.Lines {
		text += plain(line) + "\n"
	}
	if !strings.Contains(text, "Home") {
		t.Errorf("the point's label is not drawn:\n%s", text)
	}
	// A shape wholly outside the view costs nothing and draws nothing.
	far := Input{View: v, Tiles: in.Tiles, Style: in.Style, Shapes: []scene.Shape{{Kind: scene.ShapeArea, Role: uint8(colour.AlertMinorOutline), Rings: [][]scene.Vertex{lonLatBox(t, 100, -40, 110, -30)}}}, OverlaysVersion: 3}
	r2, _ := NewRenderer(v.Cols, v.Rows)
	if _, err := r2.Draw(far); err != nil {
		t.Fatal(err)
	}
	for _, c := range r2.grid.cells {
		if c.area == uint8(colour.AlertMinorTint) || c.ink == uint8(colour.AlertMinorOutline) {
			t.Fatal("a shape on the other side of the world was drawn")
		}
	}
}
