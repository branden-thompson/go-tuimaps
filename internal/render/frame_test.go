package render

import (
	"go/build"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
	"github.com/mattn/go-runewidth"
)

var sgr = regexp.MustCompile("\x1b\\[[0-9;]*m")

func plain(line string) string { return sgr.ReplaceAllString(line, "") }

// embedded decodes the embedded tiles a view of the world wants.
func embedded(t *testing.T, v project.View) []Drawn {
	t.Helper()
	ids, err := v.Tiles()
	if err != nil {
		t.Fatal(err)
	}
	var out []Drawn
	for _, id := range ids {
		body, ok := assets.Tile(id.Z, id.X, id.Y)
		if !ok {
			t.Fatalf("%v is not embedded", id)
		}
		tile, err := mvt.Decode(body, mvt.Want{Layers: []string{"water", "waterway", "park", "boundary", "transportation", "place", "water_name"}, Language: "en"}, mvt.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, Drawn{Tile: tile, At: id, Exact: true})
	}
	return out
}

func input(t *testing.T, v project.View) Input {
	return Input{View: v, Tiles: embedded(t, v), Style: style.BuiltIn(), Labels: true, Credit: textsafe.Const("OpenFreeMap (c) OpenMapTiles Data from OpenStreetMap")}
}

func render(t *testing.T, in Input) Frame {
	t.Helper()
	r, err := NewRenderer(in.View.Cols, in.View.Rows)
	if err != nil {
		t.Fatal(err)
	}
	f, err := r.Draw(in)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func fitted(cols, rows int) project.View {
	return project.View{Centre: project.LonLat{Lon: 0, Lat: 20}, Zoom: 1.2, Cols: cols, Rows: rows}
}

// TestEveryLineExactWidth is plan task 09.17 (NFR-8), measured with the
// width library directly, and TestOnlyColourSequences is 09.18 (FR-34).
func TestEveryLineExactWidth(t *testing.T) {
	for _, size := range [][2]int{{149, 38}, {69, 12}, {20, 5}, {1, 1}} {
		v := fitted(size[0], size[1])
		f := render(t, input(t, v))
		if len(f.Lines) != size[1] {
			t.Fatalf("%v: %d lines", size, len(f.Lines))
		}
		for i, line := range f.Lines {
			if got := runewidth.StringWidth(plain(line)); got != size[0] {
				t.Errorf("%v: line %d is %d cells wide", size, i, got)
			}
		}
	}
}

func TestOnlyColourSequences(t *testing.T) {
	f := render(t, input(t, fitted(149, 38)))
	for i, line := range f.Lines {
		for _, r := range plain(line) {
			if r < 0x20 || r == 0x7f || (r >= 0x80 && r < 0xa0) {
				t.Fatalf("line %d holds the control character %U outside a colour sequence", i, r)
			}
		}
		if !strings.HasPrefix(line, "\x1b[") || !strings.HasSuffix(line, "\x1b[0m") {
			t.Errorf("line %d does not set its own colours and put them back (P-07)", i)
		}
	}
}

// TestParityP06_SgrForms and P07: colour is sent only when it changes, and
// every row stands alone.
func TestParityP06_SgrForms(t *testing.T) {
	v := fitted(60, 16)
	f := render(t, input(t, v))
	for i, line := range f.Lines {
		sequences := sgr.FindAllString(line, -1)
		if len(sequences) > 2*60 {
			t.Errorf("line %d has %d colour sequences for 60 cells", i, len(sequences))
		}
		for j := 1; j < len(sequences); j++ {
			if sequences[j] == sequences[j-1] && !strings.HasSuffix(sequences[j], "[0m") {
				t.Errorf("line %d repeats %q with nothing changed", i, sequences[j])
			}
		}
	}
	mono := input(t, v)
	mono.Depth = NoColour
	for i, line := range render(t, mono).Lines {
		if strings.Contains(line, "\x1b") {
			t.Errorf("line %d of a frame with no colour holds a sequence", i)
		}
	}
}

func TestParityP07_RowSelfContainment(t *testing.T) {
	f := render(t, input(t, fitted(60, 16)))
	for i, line := range f.Lines {
		first := sgr.FindString(line)
		if !strings.Contains(first, "38;2;") && !strings.Contains(first, "48;2;") {
			t.Errorf("line %d opens with %q; each row sets its colours afresh", i, first)
		}
		if strings.ContainsAny(line, "\r\n") {
			t.Errorf("line %d holds a line ending; the host joins the rows", i)
		}
	}
}

// TestNeverBlank: the world from embedded tiles shows land, water and names.
func TestNeverBlank(t *testing.T) {
	f := render(t, input(t, fitted(149, 38)))
	text := ""
	dots := 0
	for _, line := range f.Lines {
		p := plain(line)
		text += p + "\n"
		for _, r := range p {
			if r > 0x2800 && r <= 0x28ff {
				dots++
			}
		}
	}
	if dots < 300 {
		t.Errorf("%d cells of line work; the coasts and borders of the world are more than that\n%s", dots, text)
	}
	if !strings.Contains(text, "Africa") && !strings.Contains(text, "Atlantic") && !strings.Contains(text, "Brazil") {
		t.Errorf("no place name is drawn:\n%s", text)
	}
	if !strings.Contains(text, "OpenStreetMap") {
		t.Error("the credit line is not drawn (FR-14)")
	}
	if f.Status != Complete {
		t.Errorf("status %v", f.Status)
	}
	water := 0
	for _, line := range f.Lines {
		water += strings.Count(line, "48;2;24;44;72")
	}
	if water == 0 {
		t.Error("no cell has water's background")
	}
}

// TestDeterministicAcrossMapOrder is plan task 09.21 (NFR-6).
func TestDeterministicAcrossMapOrder(t *testing.T) {
	v := fitted(100, 30)
	in := input(t, v)
	want := strings.Join(render(t, in).Lines, "\n")
	rng := rand.New(rand.NewSource(7))
	for range 5 {
		shuffled := in
		shuffled.Tiles = append([]Drawn(nil), in.Tiles...)
		rng.Shuffle(len(shuffled.Tiles), func(i, j int) { shuffled.Tiles[i], shuffled.Tiles[j] = shuffled.Tiles[j], shuffled.Tiles[i] })
		if got := strings.Join(render(t, shuffled).Lines, "\n"); got != want {
			t.Fatal("the same tiles in another order gave other bytes")
		}
	}
	r, _ := NewRenderer(v.Cols, v.Rows)
	for range 3 {
		f, err := r.Draw(in)
		if err != nil || strings.Join(f.Lines, "\n") != want {
			t.Fatal("the same renderer gave other bytes the second time")
		}
	}
}

// TestFrameStatus is plan task 09.22, and TestNoTilesNotice part of 09.16.
func TestFrameStatus(t *testing.T) {
	v := fitted(80, 20)
	in := input(t, v)
	if got := render(t, in).Status; got != Complete {
		t.Errorf("every tile exact: %v", got)
	}
	standIn := in
	standIn.Tiles = append([]Drawn(nil), in.Tiles...)
	standIn.Tiles[0].Exact = false
	if got := render(t, standIn).Status; got != Sharpening {
		t.Errorf("a stand-in on screen: %v", got)
	}
	missing := in
	missing.Missing = 1
	if got := render(t, missing).Status; got != Sharpening {
		t.Errorf("a tile with nothing to draw for it: %v", got)
	}
	none := in
	none.Tiles = nil
	f := render(t, none)
	if f.Status != NoTiles {
		t.Errorf("nothing on hand: %v", f.Status)
	}
	text := ""
	for _, line := range f.Lines {
		text += plain(line) + "\n"
	}
	if !strings.Contains(text, "assets") {
		t.Errorf("with no tile from any source the frame names the assets package (FR-23):\n%s", text)
	}
	if strings.Count(text, "\n") != 20 {
		t.Error("the frame with no tiles is not the size asked for")
	}
	for _, s := range []Status{Complete, Sharpening, NoTiles, Status(0)} {
		if s.String() == "" {
			t.Errorf("status %d has no name", s)
		}
	}
}

func TestScaleMark(t *testing.T) {
	in := input(t, fitted(100, 30))
	in.Scale = true
	text := plain(render(t, in).Lines[29])
	if !regexp.MustCompile(`\x{251c}\x{2500}+\x{2524} [0-9,]+ km`).MatchString(text) {
		t.Errorf("the bottom line has no scale mark: %q", text)
	}
	off := input(t, fitted(100, 30))
	if strings.Contains(plain(render(t, off).Lines[29]), " km") {
		t.Error("the scale mark is drawn without being asked for")
	}
}

// TestLabelCollision is plan task 09.10, the mapping's TestParityP34: cell
// coordinates, a margin of 5, overlap inclusive.
func TestLabelCollision(t *testing.T)      { testCollision(t) }
func TestParityP34_Collision(t *testing.T) { testCollision(t) }

func testCollision(t *testing.T) {
	g, err := newGrid(60, 20)
	if err != nil {
		t.Fatal(err)
	}
	ink := uint8(colour.LabelPlace)
	if !g.label(Label{X: 40, Y: 40, Name: "Paris", Rank: 1, Ink: ink}) {
		t.Fatal("the first label was not placed")
	}
	// Paris occupies cells 18 to 22 of row 10; its box is 13..28 by 8..12.
	if g.label(Label{X: 2 * 28, Y: 40, Name: "Orly", Rank: 2, Ink: ink}) {
		t.Error("a label whose box touches the first was placed; the overlap is inclusive")
	}
	if g.label(Label{X: 40, Y: 4 * 12, Name: "Ivry", Rank: 2, Ink: ink}) {
		t.Error("a label two rows below was placed; the margin is two rows")
	}
	if !g.label(Label{X: 40, Y: 4 * 15, Name: "Evry", Rank: 2, Ink: ink}) {
		t.Error("a label five rows below was refused")
	}
	if g.label(Label{X: 2, Y: 40, Name: "Brest", Rank: 2, Ink: ink}) {
		t.Error("a label that would start left of the rectangle was placed; it is skipped whole (P-33)")
	}
	if g.label(Label{X: 118, Y: 4, Name: "Strasbourg", Rank: 2, Ink: ink}) {
		t.Error("a label that would run off the right was placed")
	}
}

// TestLabelByCluster is plan task 09.11, with P-10, P-11 and P-12.
func TestLabelByCluster(t *testing.T)          { testClusters(t) }
func TestParityP10_TextCells(t *testing.T)     { testClusters(t) }
func TestParityP11_WideChars(t *testing.T)     { testClusters(t) }
func TestParityP12_TextPlacement(t *testing.T) { testClusters(t) }

func testClusters(t *testing.T) {
	g, _ := newGrid(40, 5)
	name := "\xe6\x9d\xb1\xe4\xba\xac" // two wide characters: four cells
	if !g.label(Label{X: 40, Y: 8, Name: name, Rank: 1, Ink: uint8(colour.LabelPlace)}) {
		t.Fatal("not placed")
	}
	row := g.cells[2*40 : 3*40]
	if row[18].text != "\xe6\x9d\xb1" || !row[19].taken || row[19].text != "" || row[20].text != "\xe4\xba\xac" || !row[21].taken {
		t.Errorf("cells 18 to 21: %q %q %q %q; a wide character takes two cells, centred by width and not by bytes (L-7)", row[18].text, row[19].text, row[20].text, row[21].text)
	}
	if row[17].taken || row[22].taken {
		t.Error("the label took a cell it does not occupy")
	}
	// Text wins over dots in its cell.
	if g.cells[2*40+18].ink != uint8(colour.LabelPlace) {
		t.Error("a text cell's ink is its label's")
	}
	empty, _ := newGrid(40, 5)
	if !empty.label(Label{X: 40, Y: 8, Ink: 1}) || empty.cells[2*40+20].text != "\u25C9" {
		t.Errorf("a symbol with no name draws the place glyph (P-35): %q", empty.cells[2*40+20].text)
	}
}

// TestRenderImportsNoSlowPackage is plan task 09.20: the layout rule.
func TestRenderImportsNoSlowPackage(t *testing.T) {
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, imp := range pkg.Imports {
		for _, slow := range []string{"/internal/tiles", "/internal/fetch", "/internal/overlay", "/internal/work", "/internal/mvt", "/internal/archive"} {
			if strings.HasSuffix(imp, slow) {
				t.Errorf("render imports %s; it reads only what it is given", imp)
			}
		}
	}
}

func TestRenderRefusals(t *testing.T) {
	r, err := NewRenderer(10, 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Draw(Input{View: fitted(20, 4), Style: style.BuiltIn()}); err == nil {
		t.Error("a view of another size than the renderer's must be an error")
	}
	if _, err := r.Draw(Input{View: fitted(10, 4)}); err == nil {
		t.Error("no style must be an error")
	}
	var none *Renderer
	if _, err := none.Draw(Input{}); err == nil {
		t.Error("no renderer must be an error")
	}
	_ = scene.TileID{}
}
