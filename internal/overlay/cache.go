package overlay

import (
	"context"
	"math"
	"strconv"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// defaultShapeBytes is the shape cache's default cap (D-85).
	defaultShapeBytes = 250_000
	// runLength is how many vertices one box of the run index covers.
	runLength = 64
	// bytesPerVertex is what a vertex costs the cache: 8 bytes for itself and
	// its share of a 16-byte box for each run of 64 (constants, section 3).
	bytesPerVertex = 8.25
)

// Box is one entry of a run index: the bounds of a run of 64 vertices, in
// the world's fractions. It is 16 bytes, and it is the renderer's own
// type, so that drawing from the host's memory copies no index (D-92).
type Box = scene.Run

// Path is how an overlay is drawn now.
type Path uint8

// The paths (D-90, D-92).
const (
	NotReady   Path = iota // nothing is prepared yet; a job will prepare it
	Cached                 // from the library's own simplified copy
	FromMemory             // straight from the host's memory, culled by the run index: its simplified form is larger than the whole shape cap
)

type prepared struct {
	shapes []scene.Shape
	bytes  int64
	used   uint64
}

// Use is what the shape cache reports of itself (D-90).
type Use struct{ Need, Held, Cap int64 }

// ShapeView is one live view's say in what the shape cache must keep: the
// bucket it draws at.
type ShapeView struct {
	store *Store
}

// IndexFrom is the vertex count above which Set builds a run index: the most
// vertices whose unsimplified form and index fit the shape cap together. At
// or below it the fallback can never be needed.
func (s *Store) IndexFrom() int {
	if s == nil {
		return 0
	}
	return int(float64(s.caps.ShapeBytes) / bytesPerVertex)
}

// buildIndex makes one linear pass over an overlay's rings (D-92): a box for
// each run of 64 vertices, ring after ring. A run ends on the vertex the next
// begins with, so no segment falls between two boxes. The projection rises
// with longitude and falls with latitude and never turns back, so a run's
// box is found from its extremes of longitude and latitude and only its two
// corners are projected: two projections a run, not sixty-four.
func buildIndex(o Overlay) []Box {
	var index []Box
	for _, f := range o.Features {
		for _, ring := range f.Rings {
			for start := 0; start < len(ring); start += runLength {
				end := min(start+runLength, len(ring)-1)
				west, east, south, north := 180.0, -180.0, 90.0, -90.0
				for _, p := range ring[start : end+1] {
					west, east = math.Min(west, p.Lon), math.Max(east, p.Lon)
					south, north = math.Min(south, p.Lat), math.Max(north, p.Lat)
				}
				index = append(index, boxOf(west, south, east, north))
			}
		}
	}
	return index
}

// boxOf projects a box's two corners. North is the smaller fraction.
func boxOf(west, south, east, north float64) Box {
	x0, y0, err := project.ToTile(project.LonLat{Lon: west, Lat: north}, 0)
	if err != nil {
		return Box{MaxX: math.MaxUint32, MaxY: math.MaxUint32} // cannot be placed: never culled
	}
	x1, y1, err := project.ToTile(project.LonLat{Lon: east, Lat: south}, 0)
	if err != nil {
		return Box{MaxX: math.MaxUint32, MaxY: math.MaxUint32}
	}
	return Box{MinX: fixed(x0), MinY: fixed(y0), MaxX: fixed(x1), MaxY: fixed(y1)}
}

// Index is the run index of the version this reader holds, or nil if the
// overlay is small enough never to need one.
func (r *Reader) Index() []Box {
	if r == nil || r.h == nil {
		return nil
	}
	return r.h.index
}

// Register adds a live view of the overlays. It needs nothing until it
// publishes the bucket it draws at.
func (s *Store) Register() *ShapeView {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v := &ShapeView{store: s}
	s.views[v] = math.MinInt
	return v
}

// Publish says which bucket the view draws at: prepared forms of that bucket
// are need, and are never evicted (D-90).
func (v *ShapeView) Publish(bucket int) {
	if v == nil || v.store == nil {
		return
	}
	s := v.store
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, live := s.views[v]; live {
		s.views[v] = bucket
		s.settleLocked()
	}
}

// Withdraw ends the view: what only it needed becomes spare.
func (v *ShapeView) Withdraw() {
	if v == nil || v.store == nil {
		return
	}
	s := v.store
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.views, v)
	s.settleLocked()
}

func (s *Store) neededLocked(bucket int) bool {
	for _, b := range s.views {
		if b == bucket {
			return true
		}
	}
	return false
}

// settleLocked drops the spare used longest ago until the cache is under its
// cap or holds need alone, and notes when need alone went over the cap.
func (s *Store) settleLocked() {
	need := int64(0)
	for _, buckets := range s.prepared {
		for b, p := range buckets {
			if s.neededLocked(b) {
				need += p.bytes
			}
		}
	}
	for range 1 << 16 {
		if s.shapeHeld <= int64(s.caps.ShapeBytes) {
			break
		}
		oldID, oldBucket, found := "", 0, false
		for id, buckets := range s.prepared {
			for b, p := range buckets {
				if !s.neededLocked(b) && (!found || p.used < s.prepared[oldID][oldBucket].used) {
					oldID, oldBucket, found = id, b, true
				}
			}
		}
		if !found {
			break
		}
		s.shapeHeld -= s.prepared[oldID][oldBucket].bytes
		delete(s.prepared[oldID], oldBucket)
	}
	over := need > int64(s.caps.ShapeBytes)
	if over && !s.overNeed {
		s.warnLocked(fault.CacheUnderNeed, textsafe.Const("the shape cache"))
	}
	s.overNeed, s.shapeNeed = over, need
}

// ShapeUse reports the bytes live views need, the bytes held, and the cap.
func (s *Store) ShapeUse() Use {
	if s == nil {
		return Use{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settleLocked()
	return Use{Need: s.shapeNeed, Held: s.shapeHeld, Cap: int64(s.caps.ShapeBytes)}
}

// dropPreparedLocked forgets what was prepared from an overlay's old geometry.
func (s *Store) dropPreparedLocked(id string) {
	for _, p := range s.prepared[id] {
		s.shapeHeld -= p.bytes
	}
	delete(s.prepared, id)
	delete(s.fromMemory, id)
	delete(s.fields, id)
	delete(s.pictures, id)
}

// Drawn is what is drawn for an overlay at a bucket now: its prepared form at
// that bucket, or at the nearest bucket prepared while the right one is being
// prepared (FR-11); or word that it is drawn from the host's memory; or that
// nothing is ready.
func (s *Store) Drawn(id string, bucket int) ([]scene.Shape, int, Path) {
	if s == nil || id == "" {
		return nil, 0, NotReady
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	buckets := s.prepared[id]
	have := make([]int, 0, len(buckets))
	for b := range buckets {
		have = append(have, b)
	}
	if nearest, ok := Nearest(bucket, have); ok {
		s.clock++
		buckets[nearest].used = s.clock
		return buckets[nearest].shapes, nearest, Cached
	}
	if s.fromMemory[id] {
		return nil, bucket, FromMemory
	}
	return nil, 0, NotReady
}

// TakeReleased returns, once, the ids whose replaced or removed geometry has
// been read for the last time since it was last called: what a Work or Settle
// call reports among its results (D-86).
func (s *Store) TakeReleased() []string {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.released
	s.released = nil
	return out
}

// prepareJob prepares one overlay for one bucket, inside a Work call.
type prepareJob struct {
	store  *Store
	id     string
	bucket int
}

// PrepareJob is the job that prepares an overlay for a bucket.
func (s *Store) PrepareJob(id string, bucket int) scene.Job {
	return &prepareJob{store: s, id: id, bucket: bucket}
}

// Kind says this is an overlay job.
func (j *prepareJob) Kind() scene.JobKind { return scene.KindOverlayPrepare }

// Key names the work: an overlay and a bucket.
func (j *prepareJob) Key() string {
	if j == nil {
		return ""
	}
	return "overlay/" + j.id + "/" + strconv.Itoa(j.bucket)
}

// Run reads the host's geometry, which it holds for as long as it reads, and
// keeps the result if the geometry is still the overlay's. The store's lock
// is not held while it simplifies.
func (j *prepareJob) Run(ctx context.Context) error {
	if j == nil || j.store == nil || ctx == nil {
		return fault.Make(fault.Internal, textsafe.Const("an overlay job could not run"), textsafe.Const("it has no store or no context"), textsafe.Const("this is a defect in the library; report it"))
	}
	if ctx.Err() != nil {
		return fault.Make(fault.Cancelled, textsafe.Const("preparing an overlay was abandoned"), textsafe.Const("the work it was part of was cancelled or ran out of time"), textsafe.Const("nothing; it is prepared again if it is still wanted"))
	}
	reader, ok := j.store.Read(j.id)
	if !ok {
		return nil // removed meanwhile: nothing to do
	}
	if img := reader.Overlay().Image; img != nil {
		pictures, err := j.store.readPictures(ctx, j.id, img, reader.h.kind)
		if err == nil {
			j.store.keepPictures(reader, pictures)
		}
		reader.Done()
		return err
	}
	if grid := reader.Overlay().Grid; grid != nil {
		j.store.keepField(reader, classify(grid, reader.h.kind))
		reader.Done()
		return nil
	}
	shapes, err := Prepare(reader.Overlay(), j.bucket)
	if err == nil {
		j.store.keep(reader, j.bucket, shapes)
	}
	reader.Done()
	return err
}

// keepField stores a prepared grid, if the grid it was made from is still the
// overlay's.
func (s *Store) keepField(r *Reader, field scene.Field) {
	if r == nil || r.h == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !r.h.retired {
		s.fields[r.id] = field
	}
}

// keep stores a prepared form, if the geometry it was made from is still the
// overlay's and the form fits the whole cap; otherwise the overlay is drawn
// from the host's memory.
func (s *Store) keep(r *Reader, bucket int, shapes []scene.Shape) {
	if r == nil || r.h == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.h.retired {
		return
	}
	vertices := 0
	for _, shape := range shapes {
		for _, ring := range shape.Rings {
			vertices += len(ring)
		}
	}
	bytes := int64(math.Ceil(float64(vertices) * bytesPerVertex))
	if bytes > int64(s.caps.ShapeBytes) {
		s.fromMemory[r.id] = true
		return
	}
	if s.prepared[r.id] == nil {
		s.prepared[r.id] = map[int]*prepared{}
	}
	if old, ok := s.prepared[r.id][bucket]; ok {
		s.shapeHeld -= old.bytes
	}
	s.clock++
	s.prepared[r.id][bucket] = &prepared{shapes: shapes, bytes: bytes, used: s.clock}
	s.shapeHeld += bytes
	s.settleLocked()
}
