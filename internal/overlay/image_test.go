package overlay

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func fixtureFile(t testing.TB, rel string) []byte {
	t.Helper()
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	body, err := testkit.LoadFixture(root, rel)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// providerTable is the fixture's published table: 256 entries, one "missing".
func providerTable(t testing.TB) []TableEntry {
	t.Helper()
	var raw []struct {
		DBZ *float64 `json:"dbz"`
		RGB [3]uint8 `json:"rgb"`
	}
	if err := json.Unmarshal(fixtureFile(t, "radar/provider-colour-table.json"), &raw); err != nil {
		t.Fatal(err)
	}
	var table []TableEntry
	for _, e := range raw {
		entry := TableEntry{Colour: colour.RGB{R: e.RGB[0], G: e.RGB[1], B: e.RGB[2]}, Missing: e.DBZ == nil}
		if e.DBZ != nil {
			entry.Value = *e.DBZ
		}
		table = append(table, entry)
	}
	return table
}

func radar(t testing.TB, pngBytes []byte) Overlay {
	return Overlay{ID: "radar", Valid: noon, Keeps: 10 * time.Minute, Credit: "RainViewer",
		Image: &Image{PNG: pngBytes, West: -85.4, South: 25.6, East: -79.4, North: 29.6, Projection: WebMercator,
			Table: providerTable(t), Exact: true, Type: Type{Preset: "radar", Unit: "dBZ"}}}
}

// picturePNG encodes a small picture of the given colours, one a pixel.
func picturePNG(t testing.TB, w, h int, at func(x, y int) color.Color) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, at(x, y))
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// headerOnly is a PNG's signature and header and nothing else: enough to be
// refused by, and nothing that could be decoded.
func headerOnly(w, h uint32, depth, colourType byte) []byte {
	b := []byte("\x89PNG\r\n\x1a\n")
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr, w)
	binary.BigEndian.PutUint32(ihdr[4:], h)
	ihdr[8], ihdr[9] = depth, colourType
	chunk := append([]byte("IHDR"), ihdr...)
	b = binary.BigEndian.AppendUint32(b, 13)
	b = append(b, chunk...)
	return binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(chunk))
}

// TestImageHeaderFirst is plan task 10.15 (FR-9), with 10.19, 10.20, 10.22
// and 10.23's hand-in half: everything that can be refused is refused from
// the header, before a byte is decoded.
func TestImageHeaderFirst(t *testing.T) {
	gulf := fixtureFile(t, "radar/gulf-2026-09-19.png")
	if res, err := store(t).Set(radar(t, gulf)); err != nil || !res.Created {
		t.Fatalf("the fixture's radar image: %+v, %v", res, err)
	}
	cases := []struct {
		name  string
		spoil func(o *Overlay)
		kind  fault.Kind
		says  string
	}{
		{"not a PNG", func(o *Overlay) { o.Image.PNG = []byte("GIF89a and the rest") }, fault.ImageRefused, "PNG"},
		{"no bytes", func(o *Overlay) { o.Image.PNG = nil }, fault.ImageRefused, "PNG"},
		{"16 bits a channel", func(o *Overlay) { o.Image.PNG = headerOnly(600, 400, 16, 6) }, fault.ImageRefused, "16-bit"},
		{"more than 1,048,576 pixels", func(o *Overlay) { o.Image.PNG = headerOnly(2000, 2000, 8, 6) }, fault.OverImageCap, "1,048,576"},
		{"a size that overflows", func(o *Overlay) { o.Image.PNG = headerOnly(4294967295, 4294967295, 8, 6) }, fault.OverImageCap, "1,048,576"},
		{"no size", func(o *Overlay) { o.Image.PNG = headerOnly(0, 400, 8, 6) }, fault.ImageRefused, "size"},
		{"over the map's image cap", func(o *Overlay) { o.Image.PNG = fixtureFile(t, "radar/indiana-2026-09-19.png") }, fault.OverImageCap, "250,000"},
		{"no table", func(o *Overlay) { o.Image.Table = nil }, fault.MissingTable, "table"},
		{"a projection not stated", func(o *Overlay) { o.Image.Projection = 0 }, fault.ImageRefused, "projection"},
		{"a projection the library does not have", func(o *Overlay) { o.Image.Projection = Projection(9) }, fault.ImageRefused, "projection"},
		{"a tolerance of 40", func(o *Overlay) { o.Image.Exact, o.Image.Tolerance = false, 40 }, fault.MalformedTable, "25"},
		{"a tolerance that is no number", func(o *Overlay) { o.Image.Exact, o.Image.Tolerance = false, math.NaN() }, fault.MalformedTable, "25"},
		{"bounds the wrong way round", func(o *Overlay) { o.Image.West, o.Image.East = o.Image.East, o.Image.West }, fault.InvalidCoordinates, "west"},
		{"an image and a grid together", func(o *Overlay) { o.Grid = temperatures(2, 2, func(int, int) float64 { return 1 }).Grid }, fault.SizeMismatch, "one"},
	}
	for _, c := range cases {
		o := radar(t, gulf)
		c.spoil(&o)
		if _, err := store(t).Set(o); !isKind(err, c.kind) || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: %v; want the %v kind, saying %q", c.name, err, c.kind, c.says)
		}
	}
	// The cap is a byte a pixel, and the refusal says what would fit.
	_, err := store(t).Set(radar(t, fixtureFile(t, "radar/indiana-2026-09-19.png")))
	if err == nil || !strings.Contains(err.Error(), "770,000") {
		t.Errorf("%v; the refusal says how large the image is as well as the cap", err)
	}
	if _, err := NewStore(Caps{ImageBytes: 300_000}); err == nil {
		t.Error("the image cap may be lowered, never raised")
	}
}

// TestImageBytesCappedBeforeDecode is part of 10.22.
func TestImageBytesCappedBeforeDecode(t *testing.T) {
	huge := append(headerOnly(10, 10, 8, 6), make([]byte, 9<<20)...)
	if _, err := store(t).Set(radar(t, huge)); !isKind(err, fault.OverImageCap) {
		t.Errorf("nine megabytes of PNG: %v", err)
	}
}

// TestTableRequired and TestTableRules are plan task 10.16 (D-45).
func TestTableRules(t *testing.T) {
	entry := func(r uint8, v float64) TableEntry { return TableEntry{Colour: colour.RGB{R: r}, Value: v} }
	cases := map[string][]TableEntry{
		"a colour twice":            {entry(1, 10), entry(1, 20)},
		"values out of order":       {entry(1, 20), entry(2, 10)},
		"a value that is no number": {entry(1, 10), entry(2, math.NaN())},
		"a value that is infinite":  {entry(1, 10), entry(2, math.Inf(1))},
		"nothing but missing":       {{Colour: colour.RGB{R: 1}, Missing: true}},
	}
	for name, table := range cases {
		if err := CheckTable(table); !isKind(err, fault.MalformedTable) {
			t.Errorf("%s: %v", name, err)
		}
	}
	over := make([]TableEntry, 257)
	for i := range over {
		over[i] = TableEntry{Colour: colour.RGB{R: uint8(i), G: uint8(i >> 8)}, Value: float64(i)}
	}
	if err := CheckTable(over); !isKind(err, fault.MalformedTable) {
		t.Errorf("257 entries: %v", err)
	}
	if err := CheckTable(nil); !isKind(err, fault.MissingTable) {
		t.Errorf("no table: %v", err)
	}
	if err := CheckTable(providerTable(t)); err != nil {
		t.Errorf("the provider's own table, 256 entries and one of them missing: %v", err)
	}
	same := []TableEntry{entry(1, 10), entry(2, 10), entry(3, 20)}
	if err := CheckTable(same); err != nil {
		t.Errorf("two colours for one value is a provider's business: %v", err)
	}
}

// TestExactTableZeroUnmatched is plan task 10.17 (specimen 22): with the
// provider's own table and no tolerance at all, every pixel matches.
func TestExactTableZeroUnmatched(t *testing.T) {
	s := store(t)
	s.Set(radar(t, fixtureFile(t, "radar/gulf-2026-09-19.png")))
	if err := s.PrepareJob("radar", 6).Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	raster, report, ok := s.Raster("radar")
	if !ok || raster.Width != 600 || raster.Height != 400 || len(raster.Classes) != 240_000 {
		t.Fatalf("%dx%d, %d classes, %v", raster.Width, raster.Height, len(raster.Classes), ok)
	}
	if report.Unmatched != 0 || len(report.Samples) != 0 {
		t.Errorf("%d pixels unmatched with the provider's own table, samples %v", report.Unmatched, report.Samples)
	}
	counts := map[int8]int{}
	for _, c := range raster.Classes {
		counts[c]++
	}
	if counts[NoData] == 0 || counts[1]+counts[2]+counts[3] == 0 {
		t.Errorf("classes %v; the image has clear sky and rain in it", counts)
	}
	for c := range counts {
		if c < NoData || c > 6 {
			t.Errorf("class %d; radar has six and no data", c)
		}
	}
	if raster.Preset != uint8(colour.Radar) || raster.Projection != uint8(WebMercator) || raster.West != -85.4 {
		t.Errorf("%+v", raster)
	}
	if s.OwnedBytes() < 240_000 {
		t.Errorf("the store owns %d bytes; the class bytes are its own, one a pixel (D-36)", s.OwnedBytes())
	}
}

// TestUnmatchedCountedAndSampled is plan task 10.18, with 10.23's tolerance
// and transparent pixels, and TestOneBytePerPixel of 10.19.
func TestUnmatchedCountedAndSampled(t *testing.T) {
	table := []TableEntry{{Colour: colour.RGB{G: 200}, Value: 15}, {Colour: colour.RGB{R: 250, G: 250}, Value: 35}, {Colour: colour.RGB{R: 250}, Value: 55}}
	picture := picturePNG(t, 40, 1, func(x, _ int) color.Color {
		switch {
		case x < 10:
			return color.NRGBA{G: 200, A: 255} // in the table
		case x < 20:
			return color.NRGBA{G: 204, B: 3, A: 255} // near green: inside a tolerance of 10
		case x < 30:
			return color.NRGBA{R: uint8(x), G: uint8(7 * x), B: 255, A: 255} // ten different blues, in no table
		}
		return color.NRGBA{R: 9, G: 9, B: 9, A: 0} // fully transparent: nothing there
	})
	o := Overlay{ID: "radar", Valid: noon, Keeps: time.Hour, Image: &Image{PNG: picture, West: -90, South: 30, East: -80, North: 40,
		Projection: PlateCarree, Table: table, Type: Type{Preset: "radar", Unit: "dBZ"}}}
	s := store(t)
	if _, err := s.Set(o); err != nil {
		t.Fatal(err)
	}
	s.PrepareJob("radar", 6).Run(context.Background())
	raster, report, _ := s.Raster("radar")
	if len(raster.Classes) != 40 {
		t.Fatalf("%d class bytes for 40 pixels", len(raster.Classes))
	}
	if raster.Classes[0] != 1 || raster.Classes[15] != 1 {
		t.Errorf("classes %v; 15 dBZ is class 1, and a colour within the tolerance takes its neighbour's", raster.Classes[:20])
	}
	if raster.Classes[25] != NoData || raster.Classes[35] != NoData {
		t.Error("an unmatched pixel and a transparent one are both no data")
	}
	if report.Unmatched != 10 || len(report.Samples) != 10 {
		t.Errorf("%d unmatched, %d samples; the transparent pixels are not counted", report.Unmatched, len(report.Samples))
	}
	w := s.TakeWarnings()
	if len(w) != 1 || w[0].Kind != fault.UnmatchedImageColours || w[0].Count != 10 {
		t.Errorf("warnings %+v", w)
	}
	// With no tolerance the near colour is unmatched too.
	o.Image.Exact = true
	s.Set(o)
	s.PrepareJob("radar", 6).Run(context.Background())
	if _, exact, _ := s.Raster("radar"); exact.Unmatched != 20 || len(exact.Samples) != 11 {
		t.Errorf("exact: %d unmatched, %d distinct samples", exact.Unmatched, len(exact.Samples))
	}
	// At most sixteen samples, however many colours there are.
	many := picturePNG(t, 64, 1, func(x, _ int) color.Color { return color.NRGBA{R: uint8(4 * x), B: 255, A: 255} })
	o.Image.PNG = many
	s.Set(o)
	s.PrepareJob("radar", 6).Run(context.Background())
	if _, r, _ := s.Raster("radar"); r.Unmatched != 64 || len(r.Samples) != 16 {
		t.Errorf("%d unmatched, %d samples; at most 16 are kept", r.Unmatched, len(r.Samples))
	}
}

// TestBrokenImageIsRefusedByTheJob: a PNG whose header is fine and whose body
// is not fails in the job, as the library's own error, and leaves nothing.
func TestBrokenImageIsRefusedByTheJob(t *testing.T) {
	s := store(t)
	o := radar(t, append(headerOnly(20, 20, 8, 6), []byte("no image data follows")...))
	if _, err := s.Set(o); err != nil {
		t.Fatalf("the header is a fine one: %v", err)
	}
	if err := s.PrepareJob("radar", 6).Run(context.Background()); !isKind(err, fault.ImageRefused) {
		t.Errorf("%v; want the image-refused kind", err)
	}
	if _, _, ok := s.Raster("radar"); ok {
		t.Error("a raster was kept from an image that could not be decoded")
	}
}
