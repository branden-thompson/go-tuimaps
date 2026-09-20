package tuimaps_test

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// clean reports whether a string holds nothing a terminal would read as a
// command: no escape, no control character, no stray surrogate (FR-34).
func clean(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if r == 0x1b || r < 0x20 && r != '\n' || r == 0x7f || (r >= 0x80 && r < 0xa0) {
			return false
		}
	}
	return true
}

// FuzzEveryReturnedStringIsClean is plan task 12.10 (FR-34): whatever a host
// hands in - names, ids, language codes, palette token names - nothing the
// library hands back holds an escape or a control character, and no public
// call panics. A frame's own rows are checked with the colour sequences
// taken out, since those are the library's and not text.
func FuzzEveryReturnedStringIsClean(f *testing.F) {
	f.Add("Miami", "home", "en", "water.fill", uint8(80), uint8(24), uint8(0))
	f.Add("\x1b[31mred", "\x00id", "", "", uint8(1), uint8(1), uint8(3))
	f.Add("𐀀", "a\tb", "pt-BR", "not.a.token", uint8(200), uint8(60), uint8(1))
	f.Fuzz(func(t *testing.T, name, id, language, token string, cols, rows, depth uint8) {
		size := tuimaps.Size{Cols: int(cols)%160 + 1, Rows: int(rows)%50 + 1}
		m, err := tuimaps.New(tuimaps.WithSize(size.Cols, size.Rows), tuimaps.Embed(assets.Tile, assets.MaxZoom))
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		m.ColourDepth(tuimaps.Depth(depth % 4))
		if unknown, err := m.SetPalette(map[string]tuimaps.RGB{token: {R: 1, G: 2, B: 3}}); err == nil {
			for _, u := range unknown {
				if !clean(u) {
					t.Fatalf("a name handed back from SetPalette is not clean: %q", u)
				}
			}
		}
		_ = m.LabelLanguage(language) // a code that is refused is refused; nothing is returned
		place := tuimaps.Place{ID: id, Name: name, At: tuimaps.LonLat{Lon: -80.19, Lat: 25.77}}
		if got, err := m.AddPlace(place); err == nil && !clean(got) {
			t.Fatalf("the id handed back is not clean: %q", got)
		}
		for _, p := range m.Places() {
			if !clean(p.Name) || !clean(p.ID) || !clean(p.Glyph) {
				t.Fatalf("a place the library holds is not clean: %+v", p)
			}
		}
		if _, err := m.Settle(context.Background()); err != nil {
			t.Fatalf("settle: %v", err)
		}
		frame, err := m.Render(size, time.Time{})
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		if len(frame.Lines) != size.Rows {
			t.Fatalf("%d rows for %d", len(frame.Lines), size.Rows)
		}
		for i, line := range frame.Lines {
			if !clean(colours.ReplaceAllString(line, "")) {
				t.Fatalf("row %d of the frame is not clean: %q", i, line)
			}
		}
		if id, err := m.RemovePlace(id); err == nil && id < 0 {
			t.Fatal("a count of places removed below zero")
		}
		if !clean(frame.Status.String()) || !clean(strings.Join(nothingButText(m), "")) {
			t.Fatal("something the library says about itself is not clean")
		}
	})
}

// nothingButText is every string a map will say about itself that a host
// might print.
func nothingButText(m *tuimaps.Map) []string {
	var out []string
	for _, p := range m.Places() {
		out = append(out, p.ID, p.Name, p.Glyph)
	}
	return out
}
