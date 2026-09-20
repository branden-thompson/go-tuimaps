package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// maxTotal is the bound on the embedded set (FR-28, D-82).
const maxTotal = 2_500_000

// every calls f for each tile of zoom 0 to 3.
func every(f func(id scene.TileID)) {
	for z := uint8(0); z <= MaxZoom; z++ {
		for x := uint32(0); x < 1<<z; x++ {
			for y := uint32(0); y < 1<<z; y++ {
				f(scene.TileID{Z: z, X: x, Y: y})
			}
		}
	}
}

// TestAssetsDecodeThroughGate is plan task 04.11: every embedded tile passes
// the hardened decoder with its default limits, and the whole set is under
// its bound.
func TestAssetsDecodeThroughGate(t *testing.T) {
	want := mvt.Want{Layers: []string{"water", "waterway", "landcover", "park", "boundary", "transportation", "aeroway", "place", "water_name", "aerodrome_label"}, Language: "en"}
	count, total, kept, largest := 0, 0, 0, 0
	every(func(id scene.TileID) {
		body, ok := Tile(id.Z, id.X, id.Y)
		if !ok {
			t.Errorf("%v is not embedded", id)
			return
		}
		tile, err := mvt.Decode(body, want, mvt.DefaultLimits())
		if err != nil {
			t.Errorf("%v: %v", id, err)
			return
		}
		if len(tile.Layers) == 0 {
			t.Errorf("%v decodes to nothing", id)
		}
		count++
		total += len(body)
		kept += tile.Bytes()
		largest = max(largest, tile.Bytes())
	})
	if count != 85 {
		t.Errorf("%d tiles, want 85", count)
	}
	if total > maxTotal {
		t.Errorf("the embedded set is %d bytes, over its bound of %d", total, maxTotal)
	}
	t.Logf("85 tiles: %d bytes embedded; %d bytes kept once decoded, the largest %d", total, kept, largest)
}

// TestEmbeddedTilesCarryWhatTheRolesNeed: the world tile keeps each border's
// administrative level and each place's rank. The first set generated had
// neither: the decoder read upstream's keys, which these tiles do not carry.
func TestEmbeddedTilesCarryWhatTheRolesNeed(t *testing.T) {
	body, _ := Tile(0, 0, 0)
	tile, err := mvt.Decode(body, mvt.Want{Layers: []string{"boundary", "place"}, Language: "en"}, mvt.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	countries, ranked, places := 0, 0, 0
	for _, l := range tile.Layers {
		for _, f := range l.Features {
			if l.Name == "boundary" && f.AdminLevel == 2 {
				countries++
			}
			if l.Name == "place" {
				places++
				if f.Rank > 0 {
					ranked++
				}
			}
		}
	}
	if countries == 0 || places == 0 || ranked != places {
		t.Errorf("%d country borders; %d of %d places ranked", countries, ranked, places)
	}
}

// TestOnlyTheEmbeddedZooms: what is not embedded is simply absent.
func TestOnlyTheEmbeddedZooms(t *testing.T) {
	for _, id := range []scene.TileID{{Z: 4, X: 0, Y: 0}, {Z: 1, X: 2, Y: 0}, {Z: 3, X: 0, Y: 8}, {Z: 200, X: 0, Y: 0}} {
		if body, ok := Tile(id.Z, id.X, id.Y); ok || body != nil {
			t.Errorf("%v: %d bytes, %v", id, len(body), ok)
		}
	}
}

// TestAssetsMatchTheHashList is FR-28a's re-hash: the pin, the hash list and
// the tiles in this build are the ones that were generated together.
func TestAssetsMatchTheHashList(t *testing.T) {
	lines := strings.Split(strings.TrimSpace(Hashes()), "\n")
	pin := sha256.Sum256([]byte(Pin()))
	if lines[0] != "# pin "+hex.EncodeToString(pin[:]) {
		t.Fatalf("the hash list was made for another pin: %q", lines[0])
	}
	for _, field := range []string{"archive planet/", "\nlength ", "\nentity-tag "} {
		if !strings.Contains(Pin(), field) {
			t.Errorf("the pin lacks %q", strings.TrimSpace(field))
		}
	}
	if strings.Contains(Pin(), "://") {
		t.Error("the pin holds an address; it names the archive only")
	}
	listed := map[string]string{}
	sources := 0
	for _, l := range lines[1:] {
		f := strings.Fields(l)
		if len(f) != 3 {
			t.Fatalf("a malformed line: %q", l)
		}
		if strings.HasPrefix(f[1], "source/") {
			sources++
			continue
		}
		listed[f[1]] = f[0] + " " + f[2]
	}
	if sources != 85 || len(listed) != 85 {
		t.Errorf("%d source tiles and %d output tiles listed, want 85 of each", sources, len(listed))
	}
	every(func(id scene.TileID) {
		body, _ := Tile(id.Z, id.X, id.Y)
		sum := sha256.Sum256(body)
		name := "tiles/" + strconv.Itoa(int(id.Z)) + "-" + strconv.Itoa(int(id.X)) + "-" + strconv.Itoa(int(id.Y)) + ".pbf.gz"
		if listed[name] != hex.EncodeToString(sum[:])+" "+strconv.Itoa(len(body)) {
			t.Errorf("%s does not match the hash list", name)
		}
	})
}

// TestNoticeSaysWhatTheHashesProve: the data notice credits the data and is
// honest about the hash list (FR-28, FR-28a).
func TestNoticeSaysWhatTheHashesProve(t *testing.T) {
	for _, want := range []string{"OpenStreetMap", "Open Database License", "https://opendatacommons.org/licenses/odbl/", "OpenMapTiles", "OpenFreeMap", "not that the source was authentic"} {
		if !strings.Contains(Notice(), want) {
			t.Errorf("the notice lacks %q", want)
		}
	}
}

// TestAssetsRegisterExplicitly is plan task 04.12: importing the package
// changes nothing. It has no init function, its one package-level variable
// is the embedded files, and it imports nothing of the library's, so there
// is nothing it could register with.
func TestAssetsRegisterExplicitly(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	vars, parsed := 0, 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		parsed++
		for _, imp := range file.Imports {
			if strings.Contains(imp.Path.Value, "go-tuimaps") {
				t.Errorf("imports %s; the package stands alone", imp.Path.Value)
			}
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.Name == "init" && d.Recv == nil {
					t.Error("an init function; importing the package must change nothing")
				}
			case *ast.GenDecl:
				if d.Tok == token.VAR {
					vars += len(d.Specs)
				}
			}
		}
	}
	if parsed == 0 {
		t.Fatal("no source file was read")
	}
	if vars != 1 {
		t.Errorf("%d package-level variables, want 1: the embedded files", vars)
	}
}

// TestParityP47_EmbeddedTiles is the parity row for upstream's embedded
// tiles (P-47, Extended). Upstream carries zoom 0 and zoom 1 - one tile and
// four - and consults them before disk and before the network, whatever
// source is named. This release carries zoom 0 to 3 instead (D-27, D-33),
// which is the extension; that they never override a chosen source is the
// pipeline's rule and is proved by TestAssetsNeverOverrideChosenSource,
// because this package cannot name a source at all: it hands out bytes and
// has no way to reach anything.
func TestParityP47_EmbeddedTiles(t *testing.T) {
	// Upstream's own two zooms are here, whole.
	for _, z := range []uint8{0, 1} {
		across := uint32(1) << z
		for x := range across {
			for y := range across {
				if _, ok := Tile(z, x, y); !ok {
					t.Errorf("zoom %d, tile %d,%d is missing; upstream carries this zoom whole", z, x, y)
				}
			}
		}
	}
	// And the extension: every tile of every zoom up to the deepest.
	if MaxZoom != 3 {
		t.Errorf("the deepest embedded zoom is %d; D-27 and D-33 say 3", MaxZoom)
	}
	held := 0
	for z := uint8(0); z <= MaxZoom; z++ {
		across := uint32(1) << z
		for x := range across {
			for y := range across {
				body, ok := Tile(z, x, y)
				if !ok {
					t.Errorf("zoom %d, tile %d,%d is missing", z, x, y)
					continue
				}
				if len(body) == 0 {
					t.Errorf("zoom %d, tile %d,%d is empty", z, x, y)
				}
				held++
			}
		}
	}
	if want := 1 + 4 + 16 + 64; held != want {
		t.Errorf("%d tiles are embedded, want %d - one, four, sixteen and sixty-four", held, want)
	}
	// Past the deepest zoom there is nothing, said as nothing rather than
	// as an empty tile: that is what makes the pipeline ask elsewhere.
	if _, ok := Tile(MaxZoom+1, 0, 0); ok {
		t.Error("a zoom deeper than the embedded set answered with a tile")
	}
}
