package tuimaps_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// contractKinds reads the two closed lists out of the contract, the document
// a host actually reads, so that the package and the document cannot drift.
func contractKinds(t *testing.T) (errs, warnings []string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("06_docs", "02_features", "go-tuimaps", "03-architecture-design", "contract.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		if !strings.HasPrefix(line, "| Kinds ") {
			continue
		}
		cells := strings.Split(line, " | ")
		if len(cells) != 3 {
			t.Fatalf("the contract's kinds row has %d cells, want 3", len(cells))
		}
		return names(cells[1]), names(cells[2])
	}
	t.Fatal("the contract has no kinds row")
	return nil, nil
}

func names(cell string) []string {
	var out []string
	for _, n := range strings.Split(strings.TrimSuffix(strings.TrimSpace(cell), " |"), "·") {
		if n = strings.TrimSpace(n); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// TestErrorKindsClosed is plan task 12.21: the kinds the package hands a
// host are exactly the contract's, in its order, and a host can name them
// without reaching inside the library.
func TestErrorKindsClosed(t *testing.T) {
	wantErrs, wantWarnings := contractKinds(t)
	var gotErrs []string
	for _, k := range tuimaps.Kinds() {
		gotErrs = append(gotErrs, k.String())
	}
	if strings.Join(gotErrs, " ") != strings.Join(wantErrs, " ") {
		t.Errorf("the error kinds are\n  %s\nthe contract says\n  %s", strings.Join(gotErrs, " "), strings.Join(wantErrs, " "))
	}
	var gotWarnings []string
	for _, k := range tuimaps.WarningKinds() {
		gotWarnings = append(gotWarnings, k.String())
	}
	if strings.Join(gotWarnings, " ") != strings.Join(wantWarnings, " ") {
		t.Errorf("the warning kinds are\n  %s\nthe contract says\n  %s", strings.Join(gotWarnings, " "), strings.Join(wantWarnings, " "))
	}
}

// TestEveryErrorCarriesAKind: a host can tell what went wrong from the error
// itself, without reading its words, and an error from somewhere else says
// plainly that it is not the library's.
func TestEveryErrorCarriesAKind(t *testing.T) {
	m := world(t, 80, 24)
	cases := map[string]error{
		"a place off the world":  second(m.AddPlace(tuimaps.Place{Name: "x", At: tuimaps.LonLat{Lat: 100}})),
		"an overlay with no id":  second(m.Set(tuimaps.Overlay{})),
		"a zoom that is nowhere": m.Zoom(99),
		"a language that is not": m.LabelLanguage("a whole sentence"),
		"no disk cache":          second(m.Purge()),
		"no size":                second(tuimaps.New(tuimaps.Embed(nil, 0))),
	}
	for name, err := range cases {
		kind, ok := tuimaps.KindOf(err)
		if !ok {
			t.Errorf("%s: %v carries no kind", name, err)
			continue
		}
		if kind < tuimaps.InvalidCoordinates || kind > tuimaps.Internal {
			t.Errorf("%s: kind %v is outside the closed list", name, kind)
		}
	}
	if _, ok := tuimaps.KindOf(errors.New("a host's own error")); ok {
		t.Error("an error from somewhere else was claimed as the library's")
	}
	if _, ok := tuimaps.KindOf(nil); ok {
		t.Error("no error at all was claimed as the library's")
	}
	// The kind survives wrapping, which is how a host's own layers pass it on.
	wrapped := errors.Join(errors.New("while updating the map"), m.Zoom(99))
	if _, ok := tuimaps.KindOf(wrapped); !ok {
		t.Error("the kind was lost when the error was wrapped")
	}
	_ = context.Background
}
