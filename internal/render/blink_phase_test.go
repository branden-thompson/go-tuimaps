package render

import (
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/project"
)

// A BLINK PHASE REDRAWS ONLY WHAT BLINKS (L-29; watchpost W14 P-10): the
// phase flips twice a blink period whether or not anything blinks, and a frame
// whose markers are all steady looks the same in either half - so the flip
// alone is no change. A blinking marker's frame is redrawn.
func TestTheRendererSeesThePhaseOnlyOnABlinkingMarker(t *testing.T) {
	v := fitted(60, 16)
	steady := input(t, v)
	steady.Markers = []Marker{{At: project.LonLat{Lon: 0, Lat: 20}, Shape: MarkerDot, Label: "Here"}}
	r, err := NewRenderer(v.Cols, v.Rows)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Draw(steady); err != nil {
		t.Fatal(err)
	}
	flipped := steady
	flipped.MarkerPhase = !steady.MarkerPhase
	before := r.Redraws()
	if _, err := r.Draw(flipped); err != nil {
		t.Fatal(err)
	}
	if r.Redraws() != before {
		t.Error("the phase flipped with no blinking marker and the frame was redrawn")
	}

	blinking := steady
	blinking.Markers = []Marker{{At: project.LonLat{Lon: 0, Lat: 20}, Shape: MarkerDot, Label: "Here", Blink: true}}
	if _, err := r.Draw(blinking); err != nil {
		t.Fatal(err)
	}
	lit := blinking
	lit.MarkerPhase = !blinking.MarkerPhase
	before = r.Redraws()
	if _, err := r.Draw(lit); err != nil {
		t.Fatal(err)
	}
	if r.Redraws() != before+1 {
		t.Error("a blinking marker's phase flipped and the frame was not redrawn")
	}
}
