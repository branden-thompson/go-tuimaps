package tuimaps

import (
	"github.com/branden-thompson/go-tuimaps/internal/describe"
	"github.com/branden-thompson/go-tuimaps/internal/overlay"
)

// PlantPanic makes the next call of a given name panic, so that the library's
// own tests can show that no panic escapes a public call (contract, section
// 6, rule 4). It is in a test file: no build of the library carries it.
func PlantPanic(m *Map, call string) {
	m.planted = func(at string) {
		if at == call {
			panic("planted: " + at)
		}
	}
}

// ClearPanic takes the planted panic away again.
func ClearPanic(m *Map) { m.planted = nil }

// ReportRemembered reports whether a report for these places is already
// worked out and would be returned from memory.
func ReportRemembered(m *Map, asked []Place) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.rememberedReport(asked)
	return ok
}

// ImageUse is what the map's images hold, the shared set of readings among
// it, as the image budget counts it.
func ImageUse(m *Map) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store.ImageUse()
}

// AreasMeasured is how many times the map has measured a place against an
// alert's areas - the description's costliest sum - since it was made.
func AreasMeasured(m *Map) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.areas.measured
}

// KeptAtRender reports whether a render leaves the preparation of overlay id,
// at the bucket in view (or that bucket plus off), running: the predicate a
// render's plan keeps work by.
func KeptAtRender(m *Map, id string, off int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stillWanted(m.store.PrepareJob(id, m.bucket()+off).Key())
}

// Redraws is how many frames the map's renderer has redrawn rather than
// reused.
func Redraws(m *Map) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.renderer.Redraws()
}

// TrackedFrames is how many frames of a loop the map keeps as motion last
// measured them, and the first of them, to be compared, or nil.
func TrackedFrames(m *Map, id string) (int, *describe.Frame) {
	m.mu.Lock()
	defer m.mu.Unlock()
	frames := m.tracked[id]
	if len(frames) == 0 {
		return 0, nil
	}
	return len(frames), &frames[0]
}

// DrawnFromMemory reports whether the frame draws an overlay straight from
// the host's memory (D-92) at the view's bucket.
func DrawnFromMemory(m *Map, id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, _, path := m.store.Drawn(id, m.bucket())
	return path == overlay.FromMemory
}

// DrawnBucket is the bucket an overlay's shapes are drawn from, and the
// bucket the view is at.
func DrawnBucket(m *Map, id string) (drawn, inView int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, drawn, _ = m.store.Drawn(id, m.bucket())
	return drawn, m.bucket()
}
