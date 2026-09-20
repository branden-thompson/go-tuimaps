package render

import (
	"encoding/binary"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/style"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
	"github.com/mattn/go-runewidth"
)

// fuzzTile makes a tile out of arbitrary bytes: every coordinate a 16-bit
// integer can hold, parts of any length, every kind, names that are not
// clean. It is a tile the decoder could hand over, and some it could not.
func fuzzTile(data []byte) *scene.Tile {
	layers := []string{"water", "waterway", "boundary", "transportation", "park", "place", "water_name", "aeroway"}
	classes := []string{"ocean", "lake", "river", "motorway", "minor", "rail", "city", "country", "national_park", "runway", ""}
	tile := &scene.Tile{}
	for len(data) >= 8 && len(tile.Layers) < 6 {
		head := data[:8]
		data = data[8:]
		l := scene.Layer{Name: layers[int(head[0])%len(layers)], Extent: []uint16{4096, 512, 8192, 1, 0}[int(head[1])%5]}
		features := int(head[2])%5 + 1
		for range features {
			if len(data) < 4 {
				break
			}
			f := scene.Feature{
				Kind: scene.GeomKind(data[0] % 4), Class: classes[int(data[1])%len(classes)],
				Rank: int32(data[2]) - 3, AdminLevel: data[3] % 12, Maritime: data[3]&0x80 != 0,
				FirstPart: uint32(len(l.Parts)),
			}
			if data[2]&1 == 0 {
				// A name arrives cleaned, because a name is cleaned where it is
				// read, once per tile (D-120), and the type is what says so.
				f.Name = textsafe.Clean(string(data[:4]))
			}
			parts := int(data[0]>>4)%3 + 1
			data = data[4:]
			for range parts {
				points := 0
				if len(data) > 0 {
					points = int(data[0]) % 9
					data = data[1:]
				}
				for range points {
					if len(data) < 4 {
						break
					}
					l.Coords = append(l.Coords, int16(binary.LittleEndian.Uint16(data)), int16(binary.LittleEndian.Uint16(data[2:])))
					data = data[4:]
				}
				l.Parts = append(l.Parts, uint32(len(l.Coords)))
			}
			f.EndPart = uint32(len(l.Parts))
			l.Features = append(l.Features, f)
		}
		tile.Layers = append(tile.Layers, l)
	}
	return tile
}

func FuzzPaint(f *testing.F) {
	f.Add([]byte("a tile of nothing much"), uint8(0), uint8(0), uint8(0), uint8(40), uint8(12))
	f.Add([]byte{0, 0, 3, 0, 0, 0, 0, 0, 3, 0, 9, 2, 5, 0, 0, 0, 0, 255, 127, 255, 127, 0, 128, 0, 128, 255, 127, 0, 128, 0, 0, 0, 0}, uint8(1), uint8(1), uint8(1), uint8(80), uint8(24))
	f.Add([]byte{5, 0, 1, 0, 0, 0, 0, 0, 1, 6, 2, 0, 1, 0, 8, 0, 8}, uint8(3), uint8(7), uint8(2), uint8(1), uint8(1))
	user, err := style.Parse([]byte(userStyleForFuzz))
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, data []byte, z, x, y, cols, rows uint8) {
		at := scene.TileID{Z: z % 15}
		side := uint32(1) << at.Z
		at.X, at.Y = uint32(x)%side, uint32(y)%side
		centre, err := project.FromTile(float64(at.X)+0.5, float64(at.Y)+0.5, at.Z)
		if err != nil {
			t.Skip()
		}
		v := project.View{Centre: centre, Zoom: float64(at.Z) + float64(x%4)/2, Cols: int(cols)%160 + 1, Rows: int(rows)%50 + 1}
		if v.Validate() != nil {
			t.Skip()
		}
		r, err := NewRenderer(v.Cols, v.Rows)
		if err != nil {
			t.Fatal(err)
		}
		tile := fuzzTile(data)
		for _, s := range []*style.Style{style.BuiltIn(), user} {
			in := Input{View: v, Tiles: []Drawn{{Tile: tile, At: at, Exact: y&1 == 0}}, Style: s, Labels: true, Scale: x&1 == 0, Depth: colour.Depth(rows % 4),
				Credit: textsafe.Clean(string(data[:min(len(data), 40)]))}
			frame, err := r.Draw(in)
			if err != nil {
				continue // a tile whose parts run past its coordinates is refused, not drawn
			}
			if len(frame.Lines) != v.Rows {
				t.Fatalf("%d rows for %d", len(frame.Lines), v.Rows)
			}
			for i, line := range frame.Lines {
				text := plain(line)
				if got := runewidth.StringWidth(text); got != v.Cols {
					t.Fatalf("row %d is %d cells wide, want %d: %q", i, got, v.Cols, text)
				}
				for _, c := range text {
					if c < 0x20 || c == 0x7f || (c >= 0x80 && c < 0xa0) {
						t.Fatalf("row %d holds the control character %U", i, c)
					}
				}
			}
		}
	})
}

const userStyleForFuzz = `{"layers":[
 {"id":"w","type":"fill","source-layer":"water","paint":{"fill-color":"#224"}},
 {"id":"t","type":"line","source-layer":"transportation","paint":{"line-color":{"stops":[[0,"#888"],[9,"#fff"]]},"line-width":{"stops":[[0,1],[6,4],[12,9]]}}},
 {"id":"b","type":"line","source-layer":"boundary","filter":["<=","admin_level",4],"paint":{"line-color":"#f0f","line-width":3}},
 {"id":"p","type":"symbol","source-layer":"place","paint":{"text-color":"#ff0"}}]}`
