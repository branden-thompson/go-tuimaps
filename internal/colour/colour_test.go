package colour

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// documentedTokens reads the token table of the constants document, which
// was written first: every name in backticks, a range "x.1 to x.N" filled
// in, and the alert pair repeated for each severity the row names.
func documentedTokens(t *testing.T) []string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "06_docs", "02_features", "go-tuimaps", "03-architecture-design", "constants.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, after, found := strings.Cut(string(body), "### The semantic tokens (D-63)")
	if !found {
		t.Fatal("the constants document has no token section")
	}
	section, _, _ := strings.Cut(after, "\nThe sixteen-colour depth")
	tick := regexp.MustCompile("`([^`]+)`")
	span := regexp.MustCompile("`([a-z]+)\\.1` to `[a-z]+\\.(\\d+)`")
	var out []string
	for _, line := range strings.Split(section, "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 4 || strings.Contains(cells[1], "Group") || strings.Contains(cells[1], "---") {
			continue
		}
		list := cells[2]
		if m := span.FindStringSubmatch(list); m != nil {
			n, _ := strconv.Atoi(m[2])
			for i := 1; i <= n; i++ {
				out = append(out, m[1]+"."+strconv.Itoa(i))
			}
			continue
		}
		names := tick.FindAllStringSubmatch(list, -1)
		if strings.HasPrefix(strings.TrimSpace(list), "`alert.") {
			for _, severity := range append([]string{"extreme"}, flatten(names[2:])...) {
				out = append(out, "alert."+severity+".outline", "alert."+severity+".tint")
			}
			continue
		}
		out = append(out, flatten(names)...)
	}
	return out
}

func flatten(m [][]string) []string {
	var out []string
	for _, x := range m {
		out = append(out, x[1])
	}
	return out
}

// TestTokensMatchTheDocumentedList is plan task 08.1 (D-63).
func TestTokensMatchTheDocumentedList(t *testing.T) {
	want := documentedTokens(t)
	if len(want) != 60 {
		t.Fatalf("read %d tokens from the document, want 60: %v", len(want), want)
	}
	var got []string
	for _, tok := range Tokens() {
		got = append(got, tok.Name())
	}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("tokens\n got %v\nwant %v", got, want)
	}
	for _, name := range want {
		tok, ok := ParseToken(name)
		if !ok || tok.Name() != name {
			t.Errorf("%q does not parse back to itself", name)
		}
	}
	if _, ok := ParseToken("road.motorway"); ok {
		t.Error("a name that is no token parsed")
	}
	if Token(0).Name() != "" || Token(250).Name() != "" {
		t.Error("a value that is no token has a name")
	}
}

// TestPaletteOverridesToken and TestUnsetTokenFallsBack are plan task 08.2.
func TestPaletteOverridesToken(t *testing.T) {
	p, unknown := NewPalette(map[string]RGB{"road.major": {255, 0, 0}, "water.fill": {0, 0, 80}, "road.motorway": {1, 2, 3}})
	if len(unknown) != 1 || unknown[0] != "road.motorway" {
		t.Errorf("unknown names %v", unknown)
	}
	if got, ok := p.Resolve(RoadMajor, Dark); !ok || got != (RGB{255, 0, 0}) {
		t.Errorf("road.major resolved to %v, %v", got, ok)
	}
	if got, _ := p.Resolve(RoadMajor, Light); got != (RGB{255, 0, 0}) {
		t.Errorf("a host's colour is its colour on either ground: %v", got)
	}
}

func TestUnsetTokenFallsBack(t *testing.T) {
	p, _ := NewPalette(map[string]RGB{"road.major": {255, 0, 0}})
	dark, ok := p.Resolve(WaterFill, Dark)
	if !ok || dark != (RGB{24, 44, 72}) {
		t.Errorf("water.fill on dark: %v, %v", dark, ok)
	}
	light, _ := p.Resolve(WaterFill, Light)
	if light == dark {
		t.Error("the library's defaults differ by ground")
	}
	var none Palette
	if got, ok := none.Resolve(Ground, Dark); !ok || got != (RGB{16, 22, 28}) {
		t.Errorf("the zero palette: ground %v, %v", got, ok)
	}
	if got, _ := none.Resolve(Ground, Light); got != (RGB{245, 245, 240}) {
		t.Errorf("light ground %v", got)
	}
	for _, tok := range Tokens() {
		if tok > TrackLabel {
			continue
		}
		for _, g := range []GroundKind{Dark, Light} {
			if _, ok := none.Resolve(tok, g); !ok && basemapOrFurniture(tok) {
				t.Errorf("%s has no default on ground %v", tok.Name(), g)
			}
		}
	}
	if _, ok := none.Resolve(Token(0), Dark); ok {
		t.Error("a value that is no token resolved")
	}
}

// TestContrastFormula is plan task 08.3: published WCAG pairs.
func TestContrastFormula(t *testing.T) {
	cases := []struct {
		a, b RGB
		want float64
	}{
		{RGB{0, 0, 0}, RGB{255, 255, 255}, 21},
		{RGB{255, 255, 255}, RGB{255, 255, 255}, 1},
		{RGB{119, 119, 119}, RGB{255, 255, 255}, 4.48}, // #777 on white: the well-known near miss of 4.5:1
		{RGB{118, 118, 118}, RGB{255, 255, 255}, 4.54}, // #767676: the lightest grey that passes
		{RGB{255, 0, 0}, RGB{255, 255, 255}, 4.0},
		{RGB{0, 0, 255}, RGB{255, 255, 255}, 8.59},
	}
	for _, c := range cases {
		if got := Contrast(c.a, c.b); math.Abs(got-c.want) > 0.01 {
			t.Errorf("%v on %v: %.3f, want %.2f", c.a, c.b, got, c.want)
		}
		if Contrast(c.a, c.b) != Contrast(c.b, c.a) {
			t.Errorf("%v and %v: the ratio depends on the order", c.a, c.b)
		}
	}
	if got := Luminance(RGB{255, 255, 255}); math.Abs(got-1) > 1e-9 {
		t.Errorf("white's luminance %v", got)
	}
}

// TestForegroundRule is plan task 08.4 (FR-16, D-77), over every default
// background a line can cross; TestKeepColourWherePasses is 08.5.
func TestForegroundRule(t *testing.T) {
	var none Palette
	lines := []Token{Coast, BorderCountry, BorderRegion, River, RoadMajor, RoadMinor, Rail, Park, Runway, WaterLine}
	for _, g := range []GroundKind{Dark, Light} {
		var backgrounds []RGB
		for _, tok := range []Token{Ground, WaterFill} {
			c, _ := none.Resolve(tok, g)
			backgrounds = append(backgrounds, c)
		}
		for _, bg := range backgrounds {
			for _, tok := range lines {
				own, _ := none.Resolve(tok, g)
				got := Foreground(own, bg, LineContrast)
				if Contrast(got, bg) < LineContrast {
					t.Errorf("%s on %v: %v is %.2f:1", tok.Name(), bg, got, Contrast(got, bg))
				}
				if Contrast(own, bg) >= LineContrast && got != own {
					t.Errorf("%s on %v meets 3:1 and was not kept", tok.Name(), bg)
				}
				if Contrast(own, bg) < LineContrast {
					black, white := Contrast(RGB{}, bg), Contrast(RGB{255, 255, 255}, bg)
					want := RGB{255, 255, 255}
					if black > white {
						want = RGB{}
					}
					if got != want {
						t.Errorf("%s on %v: fell back to %v; black gives %.2f and white %.2f", tok.Name(), bg, got, black, white)
					}
				}
			}
		}
	}
}

func TestKeepColourWherePasses(t *testing.T) {
	alertTint := RGB{112, 36, 28} // specimen 20: where coloured lines failed
	orange := RGB{255, 128, 80}
	if got := Foreground(orange, alertTint, LineContrast); got != orange {
		t.Errorf("the outline on its own tint is %.2f:1 and was replaced by %v", Contrast(orange, alertTint), got)
	}
	grey := RGB{95, 95, 95}
	if got := Foreground(grey, alertTint, LineContrast); got != (RGB{255, 255, 255}) {
		t.Errorf("a grey road crossing the tint (%.2f:1) became %v, want white", Contrast(grey, alertTint), got)
	}
	label := RGB{140, 140, 140}
	light := RGB{245, 245, 240}
	if Contrast(label, light) >= TextContrast {
		t.Fatal("the case no longer tests the text threshold")
	}
	if got := Foreground(label, light, TextContrast); got != (RGB{}) {
		t.Errorf("text under 4.5:1 on a light ground became %v, want black", got)
	}
	if got := Foreground(label, light, LineContrast); got != (RGB{}) && Contrast(label, light) < LineContrast {
		t.Errorf("a line under 3:1 became %v", got)
	}
}

// Plan task 08.16 (D-64).
func TestGroundPaintedByDefault(t *testing.T) {
	var g GroundChoice
	var none Palette
	colour, painted := g.InEffect(none)
	if !painted || colour != (RGB{16, 22, 28}) {
		t.Errorf("the default ground: %v painted=%v; want the dark ground, painted", colour, painted)
	}
	host, _ := NewPalette(map[string]RGB{"ground": {250, 250, 250}})
	if colour, painted := g.InEffect(host); !painted || colour != (RGB{250, 250, 250}) {
		t.Errorf("a host's ground token: %v painted=%v", colour, painted)
	}
}

func TestDeclaredGroundNotPainted(t *testing.T) {
	g, err := Declared(RGB{255, 255, 255})
	if err != nil {
		t.Fatal(err)
	}
	var none Palette
	colour, painted := g.InEffect(none)
	if painted || colour != (RGB{255, 255, 255}) {
		t.Errorf("declared: %v painted=%v", colour, painted)
	}
	if g.Kind(none) != Light {
		t.Error("a declared white ground is a light ground")
	}
}

func TestStyleByGroundLuminance(t *testing.T) {
	var none Palette
	cases := map[RGB]GroundKind{{16, 22, 28}: Dark, {245, 245, 240}: Light, {0, 0, 0}: Dark, {255, 255, 255}: Light, {100, 100, 100}: Dark, {128, 128, 128}: Light}
	for c, want := range cases {
		g, _ := Declared(c)
		if got := g.Kind(none); got != want {
			t.Errorf("%v: %v; a ground is light when black contrasts with it more than white does (black %.2f, white %.2f)", c, got, Contrast(RGB{}, c), Contrast(RGB{255, 255, 255}, c))
		}
	}
	lightTheme, _ := NewPalette(map[string]RGB{"ground": {245, 245, 240}})
	var painted GroundChoice
	if painted.Kind(lightTheme) != Light {
		t.Error("a painted ground's kind follows the ground token")
	}
	road, _ := lightTheme.Resolve(RoadMajor, painted.Kind(lightTheme))
	if Contrast(road, RGB{245, 245, 240}) < LineContrast {
		t.Errorf("the bright style's road %v fails on the light ground", road)
	}
}

func basemapOrFurniture(t Token) bool {
	return t <= MarkerLabel || t == Track || t == TrackLabel
}
