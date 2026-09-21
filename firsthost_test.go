package tuimaps

import (
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/assets"
)

// TestTokenNamesAreWhatAPaletteTakes is the first host's theming gap (task
// 14.19): a host matching the map to its own design system has to know what
// there is to match, and every name here is one SetPalette accepts.
func TestTokenNamesAreWhatAPaletteTakes(t *testing.T) {
	names := TokenNames()
	if len(names) < 10 {
		t.Fatalf("only %d token names", len(names))
	}
	m, err := New(WithSize(80, 24), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	palette := map[string]RGB{}
	for _, n := range names {
		palette[n] = RGB{R: 1, G: 2, B: 3}
	}
	refused, err := m.SetPalette(palette)
	if err != nil {
		t.Fatal(err)
	}
	if len(refused) != 0 {
		t.Errorf("the palette refused names this package published: %v", refused)
	}
}

// TestDeepestZoomIsTheTilesOwn answers what a host cannot work out for itself:
// how deep the map it built can actually draw. Zoom takes anything up to
// MaxZoom, so without this a host asks for a zoom it cannot be given and gets
// an empty picture with no error (task 14.19).
func TestDeepestZoomIsTheTilesOwn(t *testing.T) {
	m, err := New(WithSize(80, 24), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if got := m.DeepestZoom(); got != float64(assets.MaxZoom) {
		t.Errorf("the map draws to zoom %v; its tiles stop at %d", got, assets.MaxZoom)
	}
	if MaxZoom <= assets.MaxZoom {
		t.Fatal("this test assumes a map may be zoomed deeper than its tiles reach")
	}
}

// TestTheNoTilesNoticeNamesTheRealCause is the wording the first host was
// misled by. With tiles to draw from and no work run, the map told the host to
// name a source or pass the embedded tiles - both of which it had already done
// - and said nothing of the one thing that would have helped (task 14.19).
func TestTheNoTilesNoticeNamesTheRealCause(t *testing.T) {
	m, err := New(WithSize(80, 24), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	frame, err := m.Render(Size{Cols: 80, Rows: 24}, drawnAt) // nothing has run the work
	if err != nil {
		t.Fatal(err)
	}
	if frame.Status != NoTiles {
		t.Skipf("the frame is %v, so there is no notice to read", frame.Status)
	}
	text := strings.Join(frame.Lines, "\n")
	if !strings.Contains(text, "Work") && !strings.Contains(text, "Settle") {
		t.Errorf("the map has tiles and no work has run, and the notice says nothing of either:\n%s", text)
	}
	if strings.Contains(text, "name a source") {
		t.Error("the notice tells a host with tiles to name a source")
	}
}
