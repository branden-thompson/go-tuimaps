package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// TestParityP19_VisibleTiles: tiles outside the grid are dropped, so the
// world does not repeat across the antimeridian.
func TestParityP19_VisibleTiles(t *testing.T) {
	v := project.View{Centre: project.LonLat{Lon: 179, Lat: 0}, Zoom: 2, Cols: 200, Rows: 40}
	tiles, err := v.Tiles()
	if err != nil || len(tiles) == 0 {
		t.Fatal(tiles, err)
	}
	for _, id := range tiles {
		if id.Z != 2 || id.X > 3 || id.Y > 3 {
			t.Errorf("%v is outside the grid of zoom 2", id)
		}
	}
	seen := map[scene.TileID]bool{}
	for _, id := range tiles {
		if seen[id] {
			t.Errorf("%v is asked for twice; the world does not wrap", id)
		}
		seen[id] = true
	}
}

// TestParityP22_OceanOutsideWorld: beyond the world's edge is ocean, in the
// style's water colour - upstream lit dots in a colour written into the code.
func TestParityP22_OceanOutsideWorld(t *testing.T) {
	v := project.View{Centre: project.LonLat{}, Zoom: 0, Cols: 200, Rows: 80} // 400 by 320 dots round a world of 256
	p := painter(t, v)
	if err := p.World(v); err != nil {
		t.Fatal(err)
	}
	for _, cell := range [][2]int{{2, 40}, {197, 40}, {100, 1}, {100, 78}, {3, 3}} {
		if !p.Water(cell[0], cell[1]) {
			t.Errorf("cell %v is outside the world and is not ocean", cell)
		}
	}
	for _, cell := range [][2]int{{100, 40}, {40, 12}, {160, 70}} {
		if p.Water(cell[0], cell[1]) {
			t.Errorf("cell %v is inside the world and was made ocean before any tile was drawn", cell)
		}
	}
	if err := p.World(project.View{}); err == nil {
		t.Error("a view with no size must be an error")
	}
}

// TestParityP23_DrawOrderZ2 and P24: layers are drawn in one order, whatever
// order a tile carries them in.
func TestParityP23_DrawOrderZ2(t *testing.T)      { testDrawOrder(t) }
func TestParityP24_DrawOrderBelowZ2(t *testing.T) { testDrawOrder(t) }

func testDrawOrder(t *testing.T) {
	v := worldView()
	road := scene.Layer{Name: "transportation", Extent: 4096, Features: []scene.Feature{{Kind: scene.GeomLine, Class: "motorway", EndPart: 1}}, Coords: []int16{0, 2048, 4095, 2048}, Parts: []uint32{4}}
	border := scene.Layer{Name: "boundary", Extent: 4096, Features: []scene.Feature{{Kind: scene.GeomLine, AdminLevel: 2, EndPart: 1}}, Coords: []int16{0, 2048, 4095, 2048}, Parts: []uint32{4}}
	a, b := painter(t, v), painter(t, v)
	if err := a.Tile(v, &scene.Tile{Layers: []scene.Layer{road, border}}, here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	if err := b.Tile(v, &scene.Tile{Layers: []scene.Layer{border, road}}, here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	for col := range 128 {
		ga, ia := a.Lines().Cell(col, 32)
		gb, ib := b.Lines().Cell(col, 32)
		if ga != gb || ia != ib {
			t.Fatalf("cell %d: %U ink %d against %U ink %d; a tile's own layer order must not show", col, ga, ia, gb, ib)
		}
	}
	// The border is drawn after the road, as upstream draws admin after road:
	// where the two lie on the same dots the border's colour is what shows.
	if _, ink := a.Lines().Cell(60, 32); ink != uint8(colour.BorderCountry) {
		t.Errorf("ink %d where a border lies on a road; want the border's", ink)
	}
}

// TestParityP25_LabelDeferral: what a symbol rule takes is kept for the label
// pass, sorted by rank and stable among equals; with labels off, none is drawn.
func TestParityP25_LabelDeferral(t *testing.T) {
	v := worldView()
	places := scene.Layer{Name: "place", Extent: 4096,
		Features: []scene.Feature{
			{Kind: scene.GeomPoint, Class: "city", Name: "Second", Rank: 5, FirstPart: 0, EndPart: 1},
			{Kind: scene.GeomPoint, Class: "city", Name: "First", Rank: 2, FirstPart: 1, EndPart: 2},
		},
		Coords: []int16{2048, 2048, 2060, 2048}, Parts: []uint32{2, 4}}
	in := Input{View: v, Tiles: []Drawn{{Tile: &scene.Tile{Layers: []scene.Layer{places}}, At: here, Exact: true}}, Style: style.BuiltIn(), Labels: true}
	text := ""
	for _, line := range render(t, in).Lines {
		text += plain(line)
	}
	if !contains(text, "First") || contains(text, "Second") {
		t.Error("two names on one spot: the more important is placed first and the other gives way")
	}
	in.Labels = false
	text = ""
	for _, line := range render(t, in).Lines {
		text += plain(line)
	}
	if contains(text, "First") {
		t.Error("a name is drawn with labels turned off")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TestParityP28_FeatureCull: a feature whose box misses the view is not
// rastered at all. TestParityP29_Scaling: consecutive points that fall on
// one dot are one point. TestParityP30_LineClipPad: 64 dots of padding, and
// fills are clipped too (D-16).
func TestParityP28_FeatureCull(t *testing.T) {
	v := viewOf(here, 3) // an eighth of the tile's side is in view, at its middle
	p := painter(t, v)
	corner := tileWith("transportation", scene.Feature{Kind: scene.GeomLine, Class: "motorway"}, 10, 10, 200, 200)
	if err := p.Tile(v, corner, here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	if p.Culled() != 1 {
		t.Errorf("%d features culled; the road in the tile's far corner is nowhere near the view", p.Culled())
	}
	through := tileWith("transportation", scene.Feature{Kind: scene.GeomLine, Class: "motorway"}, 0, 2048, 4095, 2048)
	if err := p.Tile(v, through, here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	if p.Culled() != 1 || !p.Lines().Lit(128, 128) {
		t.Error("a road whose ends are both far outside, but which crosses the view, was culled")
	}
}

func TestParityP29_Scaling(t *testing.T) {
	v := worldView() // 256 dots for 4096 units: sixteen units a dot
	p := painter(t, v)
	wiggle := tileWith("transportation", scene.Feature{Kind: scene.GeomLine, Class: "motorway"}, 160, 160, 161, 163, 165, 170, 320, 160, 321, 161)
	if err := p.Tile(v, wiggle, here, style.BuiltIn()); err != nil {
		t.Fatal(err)
	}
	if got := p.lastPoints; got != 2 {
		t.Errorf("%d points kept of five that fall on two dots", got)
	}
}

func TestParityP30_LineClipPad(t *testing.T) {
	if clipMargin != 64 {
		t.Errorf("the pad is %d dots, want upstream's 64", clipMargin)
	}
	c := canvas(t, 4, 2)
	c.Fill([][]Point{{{-500, -500}, {500, -500}, {500, 500}, {-500, 500}}}, 1)
	for y := range 8 {
		for x := range 8 {
			if !c.Lit(x, y) {
				t.Fatalf("dot %d,%d: a fill far larger than the rectangle covers it, and only it", x, y)
			}
		}
	}
}

// TestParityP32_LabelAnchor: each vertex is tried in turn until one fits.
// TestParityP33_LabelBounds: an anchor above or left of the rectangle is
// passed over, never drawn at a negative row.
func TestParityP32_LabelAnchor(t *testing.T) { testAnchors(t) }
func TestParityP33_LabelBounds(t *testing.T) { testAnchors(t) }

func testAnchors(t *testing.T) {
	g, _ := newGrid(40, 10)
	ink := uint8(colour.LabelWater)
	placed := g.labelAt(Label{Name: "Pacific Ocean", Ink: ink}, []Point{{-40, 12}, {30, -9}, {2, 12}, {40, 20}, {60, 28}})
	if !placed {
		t.Fatal("no vertex was accepted, though the fourth fits")
	}
	if g.cells[5*40+14].text != "P" {
		t.Errorf("the name is not centred on the fourth vertex: row 5 reads %q at column 14", g.cells[5*40+14].text)
	}
	if g.labelAt(Label{Name: "Nowhere", Ink: ink}, []Point{{-40, -40}, {4000, 4000}}) {
		t.Error("a name none of whose vertices is in the rectangle was placed")
	}
	if g.labelAt(Label{Name: "Nothing", Ink: ink}, nil) {
		t.Error("a name with no vertex was placed")
	}
}
