package tuimaps

// footer_gap_test.go — WP-L11, L11.3 (D-83; watchpost UAT-1 U1-1): the
// footer's scale mark and the credit never touch, so "50 km" and the credit
// never read as one word. The credit keeps its width (FR-14); the scale mark
// takes what is left, or is not drawn.

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
)

func TestTheScaleMarkAndTheCreditNeverTouch(t *testing.T) {
	for _, cols := range []int{69, 70, 72, 76, 80, 100} {
		m, err := New(WithSize(cols, 12), Embed(assets.Tile, assets.MaxZoom))
		if err != nil {
			t.Fatal(err)
		}
		if err := m.Recentre(LonLat{Lon: -117.38, Lat: 33.2}); err != nil {
			t.Fatal(err)
		}
		if err := m.Zoom(6); err != nil {
			t.Fatal(err)
		}
		at := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
		if _, err := m.Render(Size{Cols: cols, Rows: 12}, at); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, _ = m.Settle(ctx) // the scale mark is drawn over a settled picture
		cancel()
		f, err := m.Render(Size{Cols: cols, Rows: 12}, at)
		m.Close()
		if err != nil {
			t.Fatal(err)
		}
		last := regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(f.Lines[len(f.Lines)-1], "")
		i := strings.Index(last, " km")
		if i < 0 {
			i = strings.Index(last, " mi")
		}
		if i < 0 {
			if cols >= 100 {
				t.Errorf("%d columns: no scale mark, with room for one: %q", cols, last)
			}
			continue // no room beside the whole credit: no scale mark, and nothing touches
		}
		after := []rune(last[i+3:])
		if len(after) > 0 && after[0] != ' ' && after[0] != '⠀' {
			t.Errorf("%d columns: the scale mark and the credit touch: %q", cols, last)
		}
	}
}
