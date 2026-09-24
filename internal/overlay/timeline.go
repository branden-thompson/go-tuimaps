package overlay

import (
	"sort"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
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
