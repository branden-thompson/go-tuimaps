package tuimaps_test

// tables_test.go — v0.2.0 WP-L6: the provider-table seam (L-2.5) and the
// tables the library carries.

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/colour"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
)

// iemPublished is IEM's N0Q table as the provider publishes it: index,
// dBZ, red, green, blue and the colour as hex (spec 29's input).
func iemPublished(t *testing.T) [][]string {
	t.Helper()
	body, err := os.ReadFile("06_docs/02_features/radar-loops/02-analysis/programs/inputs/spec29/n0q-table-raw.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows [][]string
	if err := json.Unmarshal(body, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

// TestTheIEMTableIsThePublishedOne is L6.2: the library's IEM table is the
// published N0Q table, entry for entry.
func TestTheIEMTableIsThePublishedOne(t *testing.T) {
	rows := iemPublished(t)
	var derived []overlay.TableEntry
	for _, row := range rows {
		v, _ := strconv.ParseFloat(row[1], 64)
		r, _ := strconv.Atoi(row[2])
		g, _ := strconv.Atoi(row[3])
		b, _ := strconv.Atoi(row[4])
		derived = append(derived, overlay.TableEntry{Colour: colour.RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, Value: v})
	}
	writeTable(t, "internal/overlay/provider_iem.go", "ProviderIEM", "IEM",
		"the Iowa Environmental Mesonet's N0Q table as it publishes it, -32 to 95 dBZ in steps of 0.5 (spec 29's input)", false, false, derived)
	table, ok := overlay.TableOf(overlay.ProviderIEM)
	if !ok {
		t.Fatal("IEM has no table")
	}
	if len(table.Entries) != len(rows) {
		t.Fatalf("%d entries; the published table has %d", len(table.Entries), len(rows))
	}
	for i, row := range rows {
		v, _ := strconv.ParseFloat(row[1], 64)
		r, _ := strconv.Atoi(row[2])
		g, _ := strconv.Atoi(row[3])
		b, _ := strconv.Atoi(row[4])
		e := table.Entries[i]
		if e.Value != v || int(e.Colour.R) != r || int(e.Colour.G) != g || int(e.Colour.B) != b || e.Missing {
			t.Errorf("entry %d: %+v; published %v", i, e, row)
		}
	}
	if table.Approximate || table.Unverified {
		t.Error("IEM's table is published, not approximate or unverified")
	}
}

// TestAnImageNamingAProviderDrawsWithItsTable is L6.1: an image with no
// table of its own that names a provider is read with that provider's
// table; one naming no provider the library has is refused.
func TestAnImageNamingAProviderDrawsWithItsTable(t *testing.T) {
	rows := iemPublished(t)
	var forty [3]uint8
	for _, row := range rows {
		if row[1] == "40" {
			r, _ := strconv.Atoi(row[2])
			g, _ := strconv.Atoi(row[3])
			b, _ := strconv.Atoi(row[4])
			forty = [3]uint8{uint8(r), uint8(g), uint8(b)}
		}
	}
	m := world(t, 80, 24)
	must(t, m.Recentre(tuimaps.LonLat{Lon: -91, Lat: 35}))
	img := tuimaps.Image{West: -95, South: 30, East: -85, North: 40, Projection: tuimaps.PlateCarree, Provider: tuimaps.ProviderIEM,
		PNG: solidPNG(t, 8, 8, color.NRGBA{R: forty[0], G: forty[1], B: forty[2], A: 255})}
	mustSet(t, m, tuimaps.RadarImage("radar", img, noon))
	settle(t, m)
	r, err := m.Report([]tuimaps.Place{{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: -91, Lat: 35}}})
	if err != nil {
		t.Fatal(err)
	}
	var answer tuimaps.Answer
	for _, a := range r.Places[0].Answers {
		if a.Overlay == "radar" {
			answer = a
		}
	}
	var entry tuimaps.LegendEntry
	for _, e := range m.Legend() {
		if e.ID == "radar" {
			entry = e
		}
	}
	if answer.NoData || answer.Class >= len(entry.Classes) || !strings.Contains(entry.Classes[answer.Class].Label, "40") {
		t.Errorf("IEM's 40 dBZ colour read as class %d (%+v); want the class that begins at 40", answer.Class, entry.Classes)
	}
	if w := m.Warnings(); len(w) != 0 {
		t.Errorf("IEM's own colour matched nothing: %v", w)
	}
	img.Provider = tuimaps.Provider(99)
	if _, err := m.Set(tuimaps.RadarImage("other", img, noon)); !isKind(err, fault.MissingTable) {
		t.Errorf("an image naming no provider the library has: %v", err)
	}
}

// TestAProviderIsOneFileOfItsOwn is L6.1's architectural test (D-70): each
// provider's table is registered by a file of its own, which names no other
// provider, so a third is added without editing the other two; and the seam
// itself names none of them.
func TestAProviderIsOneFileOfItsOwn(t *testing.T) {
	files, err := filepath.Glob("internal/overlay/provider_*.go")
	if err != nil {
		t.Fatal(err)
	}
	named := map[string]string{} // provider constant -> the file that registers it
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		var mentions []string
		ast.Inspect(file, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && strings.HasPrefix(id.Name, "Provider") && id.Name != "Provider" && id.Name != "ProviderTable" {
				mentions = append(mentions, id.Name)
			}
			return true
		})
		distinct := map[string]bool{}
		for _, m := range mentions {
			distinct[m] = true
		}
		if len(distinct) != 1 {
			t.Errorf("%s names %v; a provider's file names its own provider and no other", name, mentions)
			continue
		}
		for p := range distinct {
			named[p] = name
		}
	}
	if len(named) < 2 {
		t.Fatalf("%d provider files; IEM and MRMS each have one", len(named))
	}
	seam, err := os.ReadFile("internal/overlay/providers.go")
	if err != nil {
		t.Fatal(err)
	}
	for p := range named {
		if strings.Count(string(seam), p) > 1 {
			t.Errorf("the seam names %s beyond declaring it", p)
		}
	}
}

// writeTable writes a provider's registration file from its evidence when
// TUIMAPS_WRITE_TABLES is set, so that the table the library carries and
// the test that holds it to its evidence can never be two different things.
func writeTable(t *testing.T, path, constant, name, source string, approximate, unverified bool, entries []overlay.TableEntry) {
	t.Helper()
	if os.Getenv("TUIMAPS_WRITE_TABLES") == "" {
		return
	}
	var b strings.Builder
	b.WriteString("package overlay\n\n// Code generated by tables_test.go with TUIMAPS_WRITE_TABLES set; DO NOT EDIT.\n\n")
	b.WriteString("import \"github.com/branden-thompson/go-tuimaps/internal/colour\"\n\n")
	b.WriteString("// " + name + "'s table: " + source + ".\n")
	b.WriteString("func init() {\n\tregister(" + constant + ", ProviderTable{Name: \"" + name + "\", Approximate: " + strconv.FormatBool(approximate) +
		", Unverified: " + strconv.FormatBool(unverified) + ", Entries: []TableEntry{\n")
	for _, e := range entries {
		b.WriteString("\t\t{Colour: colour.RGB{R: " + strconv.Itoa(int(e.Colour.R)) + ", G: " + strconv.Itoa(int(e.Colour.G)) + ", B: " + strconv.Itoa(int(e.Colour.B)) +
			"}, Value: " + strconv.FormatFloat(e.Value, 'g', -1, 64) + "},\n")
	}
	b.WriteString("\t}})\n}\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %s", path)
}

// mrmsObserved is MRMS's palette as observed (wave 2, D-19): every colour of
// every recorded frame, valued by the nearest pixel of the provider's own
// legend - 500 pixels across, -20 to 70 dBZ - in the Lab the measure used.
func mrmsObserved(t *testing.T) []overlay.TableEntry {
	t.Helper()
	dir := "06_docs/02_features/radar-loops/02-analysis/programs/inputs/mrms/"
	decode := func(name string) *image.NRGBA {
		f, err := os.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		img, err := png.Decode(f)
		if err != nil {
			t.Fatal(err)
		}
		out := image.NewNRGBA(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				out.Set(x, y, color.NRGBAModel.Convert(img.At(x, y)))
			}
		}
		return out
	}
	lab := func(c color.NRGBA) [3]float64 {
		lin := func(v uint8) float64 {
			x := float64(v) / 255
			if x <= 0.04045 {
				return x / 12.92
			}
			return math.Pow((x+0.055)/1.055, 2.4)
		}
		r, g, b := lin(c.R), lin(c.G), lin(c.B)
		x := (0.4124*r + 0.3576*g + 0.1805*b) / 0.95047
		y := 0.2126*r + 0.7152*g + 0.0722*b
		z := (0.0193*r + 0.1192*g + 0.9505*b) / 1.08883
		f := func(t float64) float64 {
			if t > 0.008856 {
				return math.Cbrt(t)
			}
			return 7.787*t + 16.0/116
		}
		return [3]float64{116*f(y) - 16, 500 * (f(x) - f(y)), 200 * (f(y) - f(z))}
	}
	legend := decode(dir + "legend.png")
	type stop struct {
		l   [3]float64
		dbz float64
	}
	var stops []stop
	for y := 0; y < 10; y++ {
		for x := 0; x < 500; x++ {
			// A column is 0.18 dBZ; the value is rounded to a thousandth so that
			// every machine reads the same number (amd64 and arm64 differ in the
			// last bits of the product: arm64 fuses the multiply and the add).
			stops = append(stops, stop{lab(legend.NRGBAAt(x, y)), math.Round((-20+(float64(x)+0.5)*0.18)*1000) / 1000})
		}
	}
	files, err := filepath.Glob(dir + "nat/*.png")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, dir+"national.png", dir+"region.png", dir+"state.png")
	seen := map[color.NRGBA]bool{}
	for _, name := range files {
		img := decode(name)
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				if c := img.NRGBAAt(x, y); c.A != 0 {
					seen[c] = true
				}
			}
		}
	}
	var out []overlay.TableEntry
	for c := range seen {
		l := lab(c)
		best, value := math.Inf(1), 0.0
		for _, s := range stops {
			d := math.Sqrt((l[0]-s.l[0])*(l[0]-s.l[0]) + (l[1]-s.l[1])*(l[1]-s.l[1]) + (l[2]-s.l[2])*(l[2]-s.l[2]))
			if d < best {
				best, value = d, s.dbz
			}
		}
		out = append(out, overlay.TableEntry{Colour: colour.RGB{R: c.R, G: c.G, B: c.B}, Value: value})
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Value != b.Value {
			return a.Value < b.Value
		}
		return uint32(a.Colour.R)<<16|uint32(a.Colour.G)<<8|uint32(a.Colour.B) < uint32(b.Colour.R)<<16|uint32(b.Colour.G)<<8|uint32(b.Colour.B)
	})
	return out
}

// TestTheMRMSTableIsTheObservedPalette is L6.3 (L-2.1, L-2.4, D-19): the
// library's MRMS table is the 111 colours observed in the recorded frames,
// each valued from the legend; every one of them matches exactly; and the
// entry says it is approximate - and, until a severe day is captured, that
// its heavy end is unverified (L6.7).
func TestTheMRMSTableIsTheObservedPalette(t *testing.T) {
	derived := mrmsObserved(t)
	if len(derived) != 111 {
		t.Fatalf("%d colours observed; wave 2 counted 111", len(derived))
	}
	writeTable(t, "internal/overlay/provider_mrms.go", "ProviderMRMS", "MRMS",
		"the 111 colours observed in fifteen recorded national frames and three views (wave 2), each valued by the nearest pixel of MRMS's own legend; approximate, and its heavy end unverified until a severe day is captured (L-2.3, OW-12)",
		true, true, derived)
	table, ok := overlay.TableOf(overlay.ProviderMRMS)
	if !ok {
		t.Fatal("MRMS has no table")
	}
	if len(table.Entries) != len(derived) {
		t.Fatalf("%d entries; %d observed", len(table.Entries), len(derived))
	}
	for i := range derived {
		if table.Entries[i] != derived[i] {
			t.Errorf("entry %d: %+v; observed %+v", i, table.Entries[i], derived[i])
		}
	}
	if !table.Approximate || !table.Unverified {
		t.Error("MRMS's table must say it is approximate and its heavy end unverified")
	}
	if err := overlay.CheckTable(table.Entries); err != nil {
		t.Errorf("MRMS's table breaks the table rules: %v", err)
	}
}

// TestTheLegendSaysWhenATableIsApproximate is L5.6 (L-13.10) and L6.7: an
// image read with MRMS's table has a legend entry that says it is
// approximate and its heavy end unverified; one read with IEM's says
// neither; and one that brings its own table is the host's, whatever
// provider it names.
func TestTheLegendSaysWhenATableIsApproximate(t *testing.T) {
	m := world(t, 80, 24)
	pixel := solidPNG(t, 4, 4, color.NRGBA{R: 255, G: 106, B: 0, A: 255}) // an observed MRMS colour
	for id, p := range map[string]tuimaps.Provider{"mrms": tuimaps.ProviderMRMS, "iem": tuimaps.ProviderIEM} {
		img := tuimaps.Image{West: -95, South: 30, East: -85, North: 40, Projection: tuimaps.PlateCarree, Provider: p, PNG: pixel}
		mustSet(t, m, tuimaps.RadarImage(id, img, noon))
	}
	own := tuimaps.Image{West: -95, South: 30, East: -85, North: 40, Projection: tuimaps.PlateCarree, Provider: tuimaps.ProviderMRMS, PNG: pixel,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 255, G: 106, B: 0}, Value: 50}}}
	mustSet(t, m, tuimaps.RadarImage("own", own, noon))
	marks := map[string][2]bool{}
	for _, e := range m.Legend() {
		marks[e.ID] = [2]bool{e.Approximate, e.Unverified}
	}
	if marks["mrms"] != [2]bool{true, true} || marks["iem"] != [2]bool{false, false} || marks["own"] != [2]bool{false, false} {
		t.Errorf("approximate and unverified by entry: %v; want MRMS both, IEM and the host's own table neither", marks)
	}
}
