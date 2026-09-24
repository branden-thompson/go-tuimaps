package overlay

import (
	"sort"
	"strconv"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// Moment is one valid time on a map's timeline: every loop's frame times,
// merged. It is a gap when no loop has a picture at that time, and a forecast
// when any loop's frame at that time is one.
type Moment struct {
	Valid         time.Time
	Gap, Forecast bool
}

// Timeline is every loop's frame times, merged and in order (D-67). A single
// picture is not on it: it has no frames to step through.
func (s *Store) Timeline() []Moment {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	byTime := map[int64]*Moment{}
	for _, h := range s.current {
		img := h.overlay.Image
		if img == nil {
			continue
		}
		for _, f := range img.Frames {
			key := f.Valid.UnixNano()
			m, ok := byTime[key]
			if !ok {
				m = &Moment{Valid: f.Valid, Gap: true}
				byTime[key] = m
			}
			m.Gap = m.Gap && f.Gap
			m.Forecast = m.Forecast || f.Forecast
		}
	}
	out := make([]Moment, 0, len(byTime))
	for _, m := range byTime {
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Valid.Before(out[j].Valid) })
	return out
}

// RightNow is the newest observed picture of every loop: where a loop opens,
// and where Reset returns (D-67). It is the zero time when no loop is set.
func (s *Store) RightNow() time.Time {
	if s == nil {
		return time.Time{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var newest time.Time
	for _, h := range s.current {
		img := h.overlay.Image
		if img == nil || len(img.Frames) == 0 {
			continue
		}
		if i := shown(img); i >= 0 && img.Frames[i].Valid.After(newest) {
			newest = img.Frames[i].Valid
		}
	}
	return newest
}

// frameShownAt is the frame a loop shows at a moment: its newest picture at or
// before it, gaps held over by the picture before them (L-1.2); or, when the
// loop has nothing so early, the nearest it still holds, its oldest picture
// (D-67). The zero moment is "right now".
func frameShownAt(img *Image, at time.Time) int {
	if at.IsZero() || len(img.Frames) == 0 {
		return shown(img)
	}
	oldest := -1
	for i := len(img.Frames) - 1; i >= 0; i-- {
		f := img.Frames[i]
		if f.Gap {
			continue
		}
		if !f.Valid.After(at) {
			return i
		}
		oldest = i
	}
	return oldest
}

// RasterAt is the picture an overlay shows at a moment on the timeline: for
// a loop, the frame frameAt names; for a single picture, that picture.
func (s *Store) RasterAt(id string, at time.Time) (scene.Raster, Report, bool) {
	if s == nil || id == "" {
		return scene.Raster{}, Report{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.current[id]
	if !ok || h.overlay.Image == nil {
		return scene.Raster{}, Report{}, false
	}
	pictures, i := s.pictures[id], frameShownAt(h.overlay.Image, at)
	if i < 0 || i >= len(pictures) || !pictures[i].ready {
		return scene.Raster{}, Report{}, false
	}
	return pictures[i].raster, pictures[i].report, true
}

// Valid is the time an overlay's freshness is judged by: for a loop, its
// newest observed picture - never the frame shown, never a gap and never a
// forecast (L-1.3, D-67); otherwise, or for a loop with no observed picture,
// the valid time it was handed in with.
func Valid(o Overlay) time.Time {
	if o.Image == nil {
		return o.Valid
	}
	for i := len(o.Image.Frames) - 1; i >= 0; i-- {
		if f := o.Image.Frames[i]; !f.Gap && !f.Forecast {
			return f.Valid
		}
	}
	return o.Valid
}

// Label is what a loop's frame at a moment says of itself on the map: the
// frame the moment falls on, gaps included - so that a gap reads as a gap,
// with its own time, while the picture before it is drawn (L-1.2) - and
// whether it is a forecast. A moment before every frame is the oldest
// picture's. The zero moment is "right now".
type Label struct {
	Valid         time.Time
	Gap, Forecast bool
}

// LabelAt is the label of the loop whose frame at the moment is newest, when
// two loops play (L-1.10g), and false when no loop is set.
func (s *Store) LabelAt(at time.Time) (Label, bool) {
	if s == nil {
		return Label{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var best Label
	found := false
	for _, h := range s.current {
		img := h.overlay.Image
		if img == nil || len(img.Frames) == 0 {
			continue
		}
		i := frameShownAt(img, at)
		if i < 0 {
			continue
		}
		if !at.IsZero() {
			for j := len(img.Frames) - 1; j >= 0; j-- {
				if !img.Frames[j].Valid.After(at) {
					i = max(i, j) // the gap the moment falls on, if it falls on one
					break
				}
			}
		}
		f := img.Frames[i]
		if !found || f.Valid.After(best.Valid) {
			best, found = Label{Valid: f.Valid, Gap: f.Gap, Forecast: f.Forecast}, true
		}
	}
	return best, found
}

// CheckFrameTimes refuses a loop whose observed frame is dated past the host's
// clock by more than clocks differ: only a forecast frame may carry a future
// valid time (D-67). A refusal names the frame. A zero clock checks nothing.
func CheckFrameTimes(img *Image, now time.Time) error {
	if img == nil || now.IsZero() {
		return nil
	}
	for i, f := range img.Frames {
		if f.Forecast || !f.Valid.After(now.Add(aheadAllowed)) {
			continue
		}
		return fault.Make(fault.ImageRefused, textsafe.Const("the image was refused"),
			textsafe.Const("an observed frame is dated more than five minutes past the map's clock; only a forecast frame may be"),
			textsafe.Const("mark a forecast frame as one, or check the frame's time")).
			Of(textsafe.Join(textsafe.Const("frame "), textsafe.Clean(strconv.Itoa(i+1))))
	}
	return nil
}
