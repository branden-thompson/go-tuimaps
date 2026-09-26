package tuimaps

// standin_test.go — v0.2.0 WP-L11, L11.5 (D-83; watchpost UAT-1 U1-28): a
// replaced overlay is drawn as it was until its replacement is prepared, so
// an alert handed in again never blinks out of the frame.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
)

func severeSquare(west float64) Overlay {
	ring := []LonLat{{Lon: west, Lat: 33.0}, {Lon: west + 0.5, Lat: 33.0}, {Lon: west + 0.5, Lat: 33.4}, {Lon: west, Lat: 33.4}, {Lon: west, Lat: 33.0}}
	return Overlay{ID: "alert/a", Valid: time.Date(2026, 9, 25, 11, 0, 0, 0, time.UTC), Keeps: 6 * time.Hour,
		Features: []Feature{{Kind: Polygon, Rings: [][]LonLat{ring}, Role: AlertSevere, Severity: SeveritySevere, ID: "a"}}}
}

func TestAReplacedOverlayIsDrawnUntilItsReplacementIsReady(t *testing.T) {
	m, err := New(WithSize(100, 30), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Recentre(LonLat{Lon: -117.35, Lat: 33.2}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(7); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	size := Size{Cols: 100, Rows: 30}
	settle := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = m.Settle(ctx)
	}
	digits := func() int {
		f, err := m.Render(size, at)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Count(strings.Join(f.Lines, ""), "3") // the severity digit along the outline (D-65)
	}
	_ = digits()
	settle()
	base := digits() // the basemap's own figures: names, the scale
	if _, err := m.Set(severeSquare(-117.6)); err != nil {
		t.Fatal(err)
	}
	_ = digits()
	settle()
	if digits() <= base {
		t.Fatal("the settled overlay is not drawn: this measures nothing")
	}
	for _, west := range []float64{-117.6, -117.55} { // the same again, then moved a little
		if _, err := m.Set(severeSquare(west)); err != nil {
			t.Fatal(err)
		}
		if digits() <= base {
			t.Errorf("handed in again at %v, the overlay dropped out of the frame before its replacement was prepared", west)
		}
		settle()
		if digits() <= base {
			t.Errorf("at %v, the replacement is not drawn once prepared", west)
		}
	}
	if _, err := m.Remove("alert/a"); err != nil {
		t.Fatal(err)
	}
	settle()
	if n := digits(); n != base {
		t.Errorf("a removed overlay is still drawn (%d digits, the basemap's own %d): its stand-in outlived it", n, base)
	}
}
