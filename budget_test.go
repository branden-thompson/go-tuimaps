package tuimaps_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// solidPNG is a picture of one colour.
func solidPNG(t testing.TB, w, h int, c color.Color) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// radarLoop is a loop of n frames five minutes apart, ending at noon.
func radarLoop(t *testing.T, id string, n, w, h int) tuimaps.Overlay {
	t.Helper()
	var frames []tuimaps.LoopFrame
	for i := n - 1; i >= 0; i-- {
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(-time.Duration(i) * 5 * time.Minute), PNG: solidPNG(t, w, h, color.NRGBA{R: 200, A: 255})})
	}
	return tuimaps.RadarImage(id, tuimaps.Image{Frames: frames, West: -90, South: 30, East: -80, North: 40, Projection: tuimaps.PlateCarree,
		Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 200}, Value: 25}}, Exact: true}, noon)
}

// TestTheImageBudgetIsTheHosts is L2.6 through the public Map (L-12.1,
// D-72): the default holds a loop at the view's dot grid, a lowered budget
// refuses the next Set and says by how much, and zero puts the default back.
func TestTheImageBudgetIsTheHosts(t *testing.T) {
	m := world(t, 149, 38)
	loop := radarLoop(t, "radar", 12, 298, 152)
	if _, err := m.Set(loop); err != nil {
		t.Fatalf("twelve frames at the dot grid under the default budget: %v", err)
	}
	if err := m.SetImageBudget(1000); err != nil {
		t.Fatal(err)
	}
	second := radarLoop(t, "second", 1, 10, 10)
	_, err := m.Set(second)
	if !isKind(err, fault.OverImageCap) || !strings.Contains(err.Error(), "over its image budget of 1,000") {
		t.Fatalf("a Set over a lowered budget: %v; want a refusal that names the budget", err)
	}
	if ids := m.Overlays(); len(ids) != 1 {
		t.Errorf("lowering the budget changed what is set: %v", ids)
	}
	if err := m.SetImageBudget(0); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Set(second); err != nil {
		t.Errorf("zero is the default budget again: %v", err)
	}
	m.Close()
	if err := m.SetImageBudget(1); !isKind(err, fault.Closed) {
		t.Errorf("SetImageBudget on a closed map: %v", err)
	}
}

// A REFRESHED LOOP KEEPS DRAWING (L-27.1, watchpost U2-47's blink, its C-10
// and D-199): a loop handed in again is drawn from its old frames until the
// new ones are decoded - the frame drawn straight after the hand-in, before
// any work, still shows the radar - as a replaced shape draws its old form
// (L11.5). Live runs saw one frame with no radar on every refresh.
func TestARefreshedLoopKeepsDrawing(t *testing.T) {
	m := world(t, 80, 24)
	m.ColourDepth(tuimaps.Truecolor)
	if _, err := m.Set(radarLoop(t, "radar", 12, 160, 96)); err != nil {
		t.Fatal(err)
	}
	settle(t, m)
	var radar []tuimaps.Class
	for _, e := range m.Legend() {
		if e.Preset == "radar" {
			radar = e.Classes
		}
	}
	if before := groundsOf(radar, momentFrame(t, m, noon)); before == 0 {
		t.Fatal("the settled loop draws no radar")
	}
	if _, err := m.Set(radarLoop(t, "radar", 13, 160, 96)); err != nil { // the refresh: a frame more
		t.Fatal(err)
	}
	if after := groundsOf(radar, momentFrame(t, m, noon)); after == 0 {
		t.Error("the frame drawn after the refresh, before its decoding, shows no radar: the loop blinks")
	}
}
