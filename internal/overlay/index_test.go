package overlay

import (
	"context"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

func ringOverlay(id string, vertices int) Overlay {
	return alert(id, circle(project.LonLat{Lon: -95, Lat: 38}, 3, vertices-1))
}

// TestRunIndexBuiltInsideSet is plan task 10.26 (D-92): for a shape that might
// have to be drawn straight from the host's memory, Set itself makes one pass
// and builds the index that drawing needs. The count is the most vertices
// whose unsimplified form and index fit the shape cap together.
func TestRunIndexBuiltInsideSet(t *testing.T) {
	s := store(t)
	if s.IndexFrom() != 30303 {
		t.Fatalf("the index is built above %d vertices, want 30,303 at the default cap of 250,000 bytes", s.IndexFrom())
	}
	s.HandIn(ringOverlay("at", 30303))
	s.HandIn(ringOverlay("over", 30304))
	at, _ := s.Read("at")
	over, _ := s.Read("over")
	if at.Index() != nil {
		t.Error("at the count the fallback can never be needed, and an index was built")
	}
	index := over.Index()
	if want := (30304 + 63) / 64; len(index) != want {
		t.Fatalf("%d boxes, want %d: one for each run of 64", len(index), want)
	}
	// Every vertex lies in its run's box, and the box after it starts where it ends.
	ring, _ := Project(over.Overlay().Features[0].Rings[0])
	for i, v := range ring {
		b := index[min(i/64, len(index)-1)]
		if v.X < b.MinX || v.X > b.MaxX || v.Y < b.MinY || v.Y > b.MaxY {
			t.Fatalf("vertex %d lies outside the box of its run", i)
		}
	}
	at.Done()
	over.Done()
	// It lives with the overlay: a replace drops the old one and is released
	// at once on one goroutine; a remove drops it.
	if res, _ := s.HandIn(ringOverlay("over", 40000)); res.Created || !res.Released {
		t.Errorf("%+v", res)
	}
	again, _ := s.Read("over")
	if len(again.Index()) != (40000+63)/64 {
		t.Errorf("after the replace: %d boxes", len(again.Index()))
	}
	again.Done()
	s.Drop("over")
	if s.OwnedBytes() > 8192 {
		t.Errorf("%d bytes owned after the big overlay was removed", s.OwnedBytes())
	}
	small, _ := NewStore(Caps{ShapeBytes: 8250})
	if small.IndexFrom() != 1000 {
		t.Errorf("at a cap of 8,250 bytes the count is %d, want 1,000", small.IndexFrom())
	}
	if _, err := NewStore(Caps{ShapeBytes: 300_000}); err == nil {
		t.Error("the shape cap may be lowered, never raised")
	}
}

// TestFallbackChosenByWholeCap and TestShapePinsNeverEvicted are plan task
// 10.10 (D-90): a prepared form is cached when its vertices times 8.25 fit
// the shape cap, and drawn from the host's memory when they do not; what a
// live view draws is never evicted.
func TestFallbackChosenByWholeCap(t *testing.T) {
	s, _ := NewStore(Caps{ShapeBytes: 8250}) // room for 1,000 vertices
	s.HandIn(alert("small", circle(project.LonLat{Lon: -95, Lat: 38}, 3, 400)))
	s.HandIn(alert("huge", circle(project.LonLat{Lon: -95, Lat: 38}, 3, 5000)))
	for id, want := range map[string]Path{"small": Cached, "huge": FromMemory} {
		job := s.PrepareJob(id, 12)
		if job.Kind() != scene.KindOverlayPrepare || job.Key() == "" {
			t.Fatalf("%s: a job of kind %d, key %q", id, job.Kind(), job.Key())
		}
		if err := job.Run(context.Background()); err != nil {
			t.Fatal(err)
		}
		got, bucket, path := s.Drawn(id, 12)
		if path != want {
			t.Errorf("%s: path %v, want %v", id, path, want)
		}
		if want == Cached && (len(got) == 0 || bucket != 12) {
			t.Errorf("%s: %d shapes at bucket %d", id, len(got), bucket)
		}
	}
	// On a miss the nearest prepared bucket is drawn while the right one is prepared.
	if got, bucket, path := s.Drawn("small", 9); path != Cached || bucket != 12 || len(got) == 0 {
		t.Errorf("a miss at bucket 9: path %v, bucket %d", path, bucket)
	}
	if _, _, path := s.Drawn("never-set", 9); path != NotReady {
		t.Errorf("an overlay that is not set: %v", path)
	}
	// A replace keeps the old prepared form as a stand-in, drawn until the new
	// one is prepared (contract section 4; v0.2.0 L11.5), and a job still
	// prepares the new one: the path is not Cached.
	s.HandIn(alert("small", circle(project.LonLat{Lon: -90, Lat: 38}, 3, 400)))
	if got, _, path := s.Drawn("small", 12); path != StandIn || len(got) == 0 {
		t.Errorf("after a replace the old prepared form is not drawn until the new one is: path %v, %d shapes", path, len(got))
	}
}

func TestShapePinsNeverEvicted(t *testing.T) {
	s, _ := NewStore(Caps{ShapeBytes: 8250})
	ids := []string{"a", "b", "c"}
	for _, id := range ids {
		s.HandIn(Overlay{ID: id, Valid: noon, Keeps: time.Hour, Features: []Feature{{Kind: Polygon, Role: colour.AlertMinorOutline,
			Rings: [][]project.LonLat{circle(project.LonLat{Lon: -95, Lat: 38}, 3, 600)}}}})
	}
	view := s.Register()
	view.Publish(12)
	for _, id := range ids {
		if err := s.PrepareJob(id, 12).Run(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range ids {
		if _, _, path := s.Drawn(id, 12); path != Cached {
			t.Errorf("%s is needed by a live view and was evicted (%v)", id, path)
		}
	}
	use := s.ShapeUse()
	if use.Need <= use.Cap || use.Held != use.Need {
		t.Errorf("%+v: over its cap the cache holds exactly the need", use)
	}
	w := s.TakeWarnings()
	if len(w) != 1 || w[0].Kind != fault.CacheUnderNeed {
		t.Errorf("warnings %+v; need alone over the cap is told once", w)
	}
	// Another bucket is a spare, and has no room.
	s.PrepareJob("a", 5).Run(context.Background())
	if _, bucket, _ := s.Drawn("a", 5); bucket == 5 {
		t.Error("a spare bucket was kept with need already over the cap")
	}
	view.Withdraw()
	s.PrepareJob("a", 5).Run(context.Background())
	if use := s.ShapeUse(); use.Held > use.Cap || use.Need != 0 {
		t.Errorf("%+v after the view was withdrawn", use)
	}
}

// TestPrepareJobReportsRelease: the job is the reader, and a release that Set
// could not report is reported by the Work call that ran the job (D-86).
func TestPrepareJobReportsRelease(t *testing.T) {
	s := store(t)
	s.HandIn(alert("warnings", square(-95, 38, 1)))
	if got := s.TakeReleased(); len(got) != 0 {
		t.Errorf("%v released with nothing replaced", got)
	}
	reader, _ := s.Read("warnings")
	s.HandIn(alert("warnings", square(-96, 38, 1)))
	reader.Done()
	if got := s.TakeReleased(); len(got) != 1 || got[0] != "warnings" {
		t.Errorf("released %v", got)
	}
	if got := s.TakeReleased(); len(got) != 0 {
		t.Errorf("a release is reported once: %v", got)
	}
	if err := s.PrepareJob("gone", 3).Run(context.Background()); err != nil {
		t.Errorf("a job for an overlay that was removed meanwhile is nothing to do, not a failure: %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.PrepareJob("warnings", 3).Run(cancelled); !isKind(err, fault.Cancelled) {
		t.Errorf("%v", err)
	}
}

// TestOldShapeDrawsFromOwnCopy finishes plan task 10.5 (D-86, FR-11): once a
// replace is reported released, the host may write over the memory it handed
// in. A shape already prepared from that memory goes on drawing as it was,
// because what was prepared is the library's own copy of the geometry, in the
// fixed point the renderer draws in.
func TestOldShapeDrawsFromOwnCopy(t *testing.T) {
	s := store(t)
	ring := square(-95, 38, 1)
	s.HandIn(alert("warnings", ring))
	if err := s.PrepareJob("warnings", 5).Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	shapes, _, path := s.Drawn("warnings", 5)
	if path != Cached || len(shapes) == 0 || len(shapes[0].Rings) == 0 {
		t.Fatalf("path %v, %d shapes; want a prepared form to draw", path, len(shapes))
	}
	was := append([]scene.Vertex(nil), shapes[0].Rings[0]...)
	res, err := s.HandIn(alert("warnings", square(-20, -20, 1)))
	if err != nil || !res.Released {
		t.Fatalf("a replace with nothing reading the old geometry: %+v, %v; want it released at once", res, err)
	}
	for i := range ring { // the host reuses what it was told it may
		ring[i] = project.LonLat{Lon: 170, Lat: -80}
	}
	if len(shapes[0].Rings[0]) != len(was) {
		t.Fatalf("the prepared ring is now %d vertices, was %d", len(shapes[0].Rings[0]), len(was))
	}
	for i, v := range shapes[0].Rings[0] {
		if v != was[i] {
			t.Fatalf("vertex %d moved to %v from %v when the host reused its own memory", i, v, was[i])
		}
	}
	// UNTIL THE REPLACEMENT IS PREPARED, THE OLD SHAPE IS DRAWN FROM THIS SAME
	// COPY (contract section 4, v0.2.0 L11.5): the host's reuse of its memory
	// above cannot reach it.
	stand, _, path := s.Drawn("warnings", 5)
	if path != StandIn || len(stand) == 0 || len(stand[0].Rings[0]) != len(was) {
		t.Fatalf("path %v for the replaced overlay; want its old copy drawn until the new shape is prepared", path)
	}
	for i, v := range stand[0].Rings[0] {
		if v != was[i] {
			t.Fatalf("the stand-in's vertex %d moved to %v from %v", i, v, was[i])
		}
	}
}

// TestALargeReplacementGetsNoStandIn is D-92 beside L11.5: a replacement
// large enough to be drawn from the host's memory is drawn so, and the old
// shape's copy is dropped, not kept as a stand-in.
func TestALargeReplacementGetsNoStandIn(t *testing.T) {
	s, err := NewStore(Caps{ShapeBytes: 8250}) // an index from 1,000 vertices
	if err != nil {
		t.Fatal(err)
	}
	s.HandIn(alert("big", circle(project.LonLat{Lon: -90, Lat: 38}, 3, 40)))
	if err := s.PrepareJob("big", 12).Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, _, path := s.Drawn("big", 12); path != Cached {
		t.Fatalf("the small shape is not prepared: %v", path)
	}
	s.HandIn(alert("big", circle(project.LonLat{Lon: -90, Lat: 38}, 3, 5000)))
	if _, _, path := s.Drawn("big", 12); path == StandIn {
		t.Error("a replacement drawn from memory kept the old shape's copy as a stand-in (D-92 drops it)")
	}
}

// TestAStandInIsLetGoOnceItsWorkIsDone is L11.5's accounting: the stand-in's
// bytes are held while it stands in, and let go when the replacement is
// prepared or the overlay is removed - never a copy the cap goes on counting.
func TestAStandInIsLetGoOnceItsWorkIsDone(t *testing.T) {
	prepare := func(s *Store, id string) {
		t.Helper()
		if err := s.PrepareJob(id, 12).Run(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	s, _ := NewStore(Caps{})
	s.HandIn(alert("a", circle(project.LonLat{Lon: -90, Lat: 38}, 3, 400)))
	prepare(s, "a")
	old := s.ShapeUse().Held
	s.HandIn(alert("a", circle(project.LonLat{Lon: -90, Lat: 38}, 2, 200)))
	if got := s.ShapeUse().Held; got != old {
		t.Errorf("while it stands in the old copy holds %d bytes, want its %d", got, old)
	}
	prepare(s, "a")
	fresh, _ := NewStore(Caps{})
	fresh.HandIn(alert("a", circle(project.LonLat{Lon: -90, Lat: 38}, 2, 200)))
	prepare(fresh, "a")
	if got, want := s.ShapeUse().Held, fresh.ShapeUse().Held; got != want {
		t.Errorf("with the replacement prepared %d bytes are held, want the replacement's %d alone", got, want)
	}
	s.HandIn(alert("a", circle(project.LonLat{Lon: -90, Lat: 38}, 3, 400))) // a stand-in again
	if _, err := s.Drop("a"); err != nil {
		t.Fatal(err)
	}
	if got := s.ShapeUse().Held; got != 0 {
		t.Errorf("a removed overlay's stand-in still holds %d bytes", got)
	}
}
