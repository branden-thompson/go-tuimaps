package textsafe

import (
	"strings"
	"testing"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/mattn/go-runewidth"
)

// samples are strings a map meets: plain names, accents, wide scripts,
// emoji with joiners and selectors, ambiguous-width symbols, and damage.
var samples = []string{
	"Fort Wayne", "Z\u00FCrich", "\u6771\u4EAC", "\uC11C\uC6B8", "\u0628\u064A\u0631\u0648\u062A",
	"e\u0301cole", "\U0001F1FA\U0001F1F8 USA", "\U0001F468\u200D\U0001F469\u200D\U0001F467 park",
	"\u2764\uFE0F", "\u00B0C \u00B1 2", "\u2191 N", "\u2588\u2591\u2592\u2593", "a\xffb", "\u0915\u094D\u0937\u093F", "",
}

// TestWidthMatchesHostTable is plan task 02.5: the library measures exactly
// as the first host's width table does, in the fixed narrow condition.
func TestWidthMatchesHostTable(t *testing.T) {
	narrow := runewidth.Condition{EastAsianWidth: false, StrictEmojiNeutral: true}
	for _, s := range samples {
		text := Clean(s)
		if got, want := Width(text), narrow.StringWidth(text.String()); got != want {
			t.Errorf("Width(%q) = %d; the host's table says %d", text, got, want)
		}
	}
}

func TestEveryClusterIsOneOrTwoCells(t *testing.T) {
	for _, s := range samples {
		clusters := graphemes.FromString(Clean(s).String())
		for clusters.Next() {
			if w := Width(Clean(clusters.Value())); w != 1 && w != 2 {
				t.Errorf("cluster %q of %q measures %d cells; a cluster occupies one or two", clusters.Value(), s, w)
			}
		}
	}
}

// TestFitDropsWholeCluster is plan task 02.6.
func TestFitDropsWholeCluster(t *testing.T) {
	cases := []struct {
		in    string
		cells int
		want  string
	}{
		{"Fort Wayne", 4, "Fort"},
		{"Fort Wayne", 99, "Fort Wayne"},
		{"Fort Wayne", 0, ""},
		{"Fort Wayne", -3, ""},
		{"\u6771\u4EAC", 3, "\u6771"},                           // the second character is two cells wide and only one is left
		{"\u6771\u4EAC", 1, ""},                                 // not even the first fits
		{"ab\u6771", 3, "ab"},                                   // never half a wide character
		{"e\u0301cole", 1, "e\u0301"},                           // the accent stays with its letter
		{"x\U0001F468\u200D\U0001F469\u200D\U0001F467", 2, "x"}, // a joined family is one cluster, kept or dropped whole
	}
	for _, c := range cases {
		got := Fit(Clean(c.in), c.cells)
		if got.String() != c.want {
			t.Errorf("Fit(%q, %d) = %q, want %q", c.in, c.cells, got, c.want)
		}
		if c.cells > 0 && Width(got) > c.cells {
			t.Errorf("Fit(%q, %d) is %d cells wide", c.in, c.cells, Width(got))
		}
	}
}

// TestQuoteCutTo64Clusters is plan task 02.7.
func TestQuoteCutTo64Clusters(t *testing.T) {
	short := "a name\x1b[2J from a tile"
	if got := Quote(short).String(); got != "a name[2J from a tile" {
		t.Errorf("Quote(%q) = %q; quoted text is cleaned", short, got)
	}
	long := strings.Repeat("e\u0301", 200)
	got := Quote(long).String()
	if !strings.HasPrefix(got, strings.Repeat("e\u0301", QuoteClusters)) || strings.HasPrefix(got, strings.Repeat("e\u0301", QuoteClusters+1)) {
		t.Errorf("Quote kept the wrong number of clusters: %d bytes", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("a cut quotation must say it was cut: %q", got[len(got)-8:])
	}
	exact := strings.Repeat("x", QuoteClusters)
	if got := Quote(exact).String(); got != exact {
		t.Errorf("text of exactly %d clusters was changed: %q", QuoteClusters, got)
	}
}

// TestIDValidation is plan task 02.8: ids are never cleaned, so cleaning can
// never make two ids collide. An id that would need cleaning is refused.
func TestIDValidation(t *testing.T) {
	for _, id := range []string{"alerts", "radar/KIWX", "zone:INZ009", "temp 2m", "\u6771\u4EAC-1", "e\u0301"} {
		got, err := ID(id)
		if err != nil {
			t.Errorf("ID(%q) refused: %v", id, err)
			continue
		}
		if got.String() != id {
			t.Errorf("ID(%q) returned %q; an id comes back byte for byte", id, got)
		}
	}
	bad := map[string]string{
		"":                                "empty",
		"a\x1bb":                          "an escape byte",
		"a\u200Bb":                        "a zero-width space",
		"a\u202Eb":                        "a bidirectional control",
		"a\xffb":                          "invalid UTF-8",
		"line\nbreak":                     "a line break",
		strings.Repeat("x", MaxIDBytes+1): "too long",
	}
	for id, why := range bad {
		if _, err := ID(id); err == nil {
			t.Errorf("ID(%q) accepted; it holds %s", id, why)
		}
	}
	if _, err := ID(strings.Repeat("x", MaxIDBytes)); err != nil {
		t.Errorf("an id of exactly %d bytes was refused: %v", MaxIDBytes, err)
	}
}
