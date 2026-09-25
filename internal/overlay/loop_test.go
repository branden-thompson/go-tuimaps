package overlay

// loop_test.go — v0.2.0 WP-L2: loops at hand-in. Frames, their checks, the
// store's own copy of every slice, and a decode that reads the header first.

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

var (
	rain  = color.NRGBA{R: 200, A: 255}
	storm = color.NRGBA{G: 200, A: 255}
)

// loopTable is a two-colour table: light rain and a storm.
func loopTable() []TableEntry {
	return []TableEntry{{Colour: colour.RGB{R: 200}, Value: 25}, {Colour: colour.RGB{G: 200}, Value: 55}}
}

// frameAt is a frame i steps of five minutes before noon, one colour all over.
func frameAt(t testing.TB, stepsBack int, c color.Color) LoopFrame {
	t.Helper()
	return LoopFrame{Valid: noon.Add(-time.Duration(stepsBack) * 5 * time.Minute), PNG: picturePNG(t, 8, 6, func(int, int) color.Color { return c })}
}

// gapAt is a gap i steps before noon.
func gapAt(stepsBack int) LoopFrame {
	return LoopFrame{Valid: noon.Add(-time.Duration(stepsBack) * 5 * time.Minute), Gap: true}
}

func loopOf(frames ...LoopFrame) Overlay {
	return Overlay{ID: "loop", Valid: noon, Keeps: 10 * time.Minute, Credit: "the host's radar",
		Image: &Image{Frames: frames, West: -90, South: 30, East: -80, North: 40, Projection: PlateCarree,
			Table: loopTable(), Exact: true, Type: Type{Preset: "radar", Unit: "dBZ"}}}
}

// TestAnImageCarriesFrames is L2.1 (L-1.1): an image may carry frames, and a
// single picture is PNG with no frames; one with both is refused.
func TestAnImageCarriesFrames(t *testing.T) {
	s := store(t)
	if _, err := s.HandIn(loopOf(frameAt(t, 2, rain), frameAt(t, 1, storm), frameAt(t, 0, rain))); err != nil {
		t.Fatalf("a loop of three frames: %v", err)
	}
	both := loopOf(frameAt(t, 0, rain))
	both.Image.PNG = frameAt(t, 0, rain).PNG
	_, err := s.HandIn(both)
	if !isKind(err, fault.ImageRefused) || !strings.Contains(err.Error(), "both") {
		t.Errorf("a picture and frames together: %v; want a refusal that says it has both", err)
	}
}

// TestEveryFrameIsCheckedAtHandIn is L2.2 (L-1.9, L-1.15): every frame is
// checked, times rise, a gap has no bytes, there are at most MaxFrames, gaps
// included, and a refusal names the frame.
func TestEveryFrameIsCheckedAtHandIn(t *testing.T) {
	good := func() []LoopFrame {
		var frames []LoopFrame
		for i := 11; i >= 0; i-- {
			frames = append(frames, frameAt(t, i, rain))
		}
		return frames
	}
	tooMany := []LoopFrame{frameAt(t, MaxFrames, rain)}
	for i := MaxFrames - 1; i >= 0; i-- {
		tooMany = append(tooMany, gapAt(i))
	}
	cases := []struct {
		name   string
		frames []LoopFrame
		kind   fault.Kind
		says   string
	}{
		{"out of order", func() []LoopFrame { f := good(); f[3], f[4] = f[4], f[3]; return f }(), fault.ImageRefused, "frame 5"},
		{"a duplicate time", func() []LoopFrame { f := good(); f[6].Valid = f[5].Valid; return f }(), fault.ImageRefused, "frame 7"},
		{"a gap with bytes", func() []LoopFrame { f := good(); f[2].Gap = true; return f }(), fault.ImageRefused, "frame 3"},
		{"a picture frame with no bytes", func() []LoopFrame { f := good(); f[8].PNG = nil; return f }(), fault.ImageRefused, "frame 9"},
		{"a frame with no time", func() []LoopFrame { f := good(); f[0].Valid = time.Time{}; return f }(), fault.ImageRefused, "frame 1"},
		{"a bad PNG in frame 7", func() []LoopFrame {
			f := good()
			f[6].PNG = []byte("not a picture at all, not even close to one")
			return f
		}(), fault.ImageRefused, "frame 7"},
		{"every frame a gap", []LoopFrame{gapAt(1), gapAt(0)}, fault.ImageRefused, "gap"},
		{"frame 73", tooMany, fault.OverImageCap, "72"},
	}
	for _, c := range cases {
		_, err := store(t).HandIn(loopOf(c.frames...))
		if !isKind(err, c.kind) || !strings.Contains(err.Error(), c.says) {
			t.Errorf("%s: %v; want %v saying %q", c.name, err, c.kind, c.says)
		}
	}
	if _, err := store(t).HandIn(loopOf(append(good(), gapAt(-1))...)); err != nil {
		t.Errorf("twelve frames and a gap: %v", err)
	}
	if _, err := store(t).HandIn(loopOf(tooMany[1:]...)); err == nil || !strings.Contains(err.Error(), "gap") {
		t.Errorf("72 gaps and no picture: %v", err)
	}
	exactly := append([]LoopFrame{frameAt(t, MaxFrames-1, rain)}, tooMany[2:]...)
	if len(exactly) != MaxFrames {
		t.Fatalf("the case holds %d frames, not %d", len(exactly), MaxFrames)
	}
	if _, err := store(t).HandIn(loopOf(exactly...)); err != nil {
		t.Errorf("exactly %d frames, gaps included: %v", MaxFrames, err)
	}
}

// classesOf prepares an overlay and returns the classes of the picture shown.
func classesOf(t *testing.T, s *Store, id string) []int8 {
	t.Helper()
	if err := s.PrepareJob(id, 0).Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	raster, _, ok := s.Raster(id)
	if !ok {
		t.Fatalf("%s: nothing prepared", id)
	}
	return append([]int8(nil), raster.Classes...)
}

// TestTheStoreKeepsItsOwnCopy is L2.3 (L-1.14): the host's PNG and table,
// for the single image and every frame, are never read after Set returns.
func TestTheStoreKeepsItsOwnCopy(t *testing.T) {
	for name, make := range map[string]func() Overlay{
		"a single picture": func() Overlay { o := loopOf(); o.Image.PNG = frameAt(t, 0, storm).PNG; return o },
		"a loop":           func() Overlay { return loopOf(frameAt(t, 1, rain), frameAt(t, 0, storm)) },
	} {
		want := classesOf(t, storeWith(t, make()), "loop")

		o := make()
		s := storeWith(t, o)
		for i := range o.Image.PNG {
			o.Image.PNG[i] = 0
		}
		for _, f := range o.Image.Frames {
			for i := range f.PNG {
				f.PNG[i] = 0
			}
		}
		for i := range o.Image.Table {
			o.Image.Table[i].Colour = colour.RGB{B: 1}
			o.Image.Table[i].Value = -1
		}
		got := classesOf(t, s, "loop")
		if !bytes.Equal(int8Bytes(want), int8Bytes(got)) {
			t.Errorf("%s: the host's slices, changed after Set, changed what is drawn", name)
		}
	}
}

func storeWith(t *testing.T, o Overlay) *Store {
	t.Helper()
	s := store(t)
	if _, err := s.HandIn(o); err != nil {
		t.Fatal(err)
	}
	return s
}

func int8Bytes(in []int8) []byte {
	out := make([]byte, len(in))
	for i, v := range in {
		out[i] = byte(v)
	}
	return out
}

// TestDecodeReadsTheHeaderBeforeItAllocates is L2.4 (L-1.14): a picture
// whose header claims four times the cap is refused by the decode itself,
// before the decoder allocates for it.
func TestDecodeReadsTheHeaderBeforeItAllocates(t *testing.T) {
	big := image.NewGray(image.Rect(0, 0, 1000, 1000)) // 1,000,000 pixels: four times the 250,000 cap
	var buf bytes.Buffer
	if err := png.Encode(&buf, big); err != nil {
		t.Fatal(err)
	}
	img := &Image{West: -90, South: 30, East: -80, North: 40, Projection: PlateCarree, Table: loopTable(), Exact: true}
	kind, err := ResolveType(Type{Preset: "radar", Unit: "dBZ"})
	if err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	_, _, err = rasterise(img, buf.Bytes(), kind, defaultImageBytes)
	runtime.ReadMemStats(&after)
	if !isKind(err, fault.OverImageCap) {
		t.Errorf("a picture of four times the cap: %v; want it refused as over the cap", err)
	}
	if grew := after.TotalAlloc - before.TotalAlloc; grew > 256<<10 {
		t.Errorf("refusing it allocated %d bytes; the header alone should have refused it", grew)
	}
}

// TestARefreshDecodesOnlyWhatIsNew is L2.5 (L-1.7): a frame's key is its
// valid time and what its reading depends on, and a re-Set keeps the decoded
// frames whose keys it already holds. A refresh that drops the oldest frame
// and adds a newest decodes exactly one.
func TestARefreshDecodesOnlyWhatIsNew(t *testing.T) {
	var frames []LoopFrame
	for i := 12; i >= 0; i-- {
		c := rain
		if i%3 == 0 {
			c = storm
		}
		frames = append(frames, frameAt(t, i, c))
	}
	s := storeWith(t, loopOf(frames[:12]...))
	classesOf(t, s, "loop")
	if s.decodes != 12 {
		t.Fatalf("the first hand-in decoded %d frames, want 12", s.decodes)
	}
	if _, err := s.HandIn(loopOf(frames[1:]...)); err != nil {
		t.Fatal(err)
	}
	classesOf(t, s, "loop")
	if got := s.decodes - 12; got != 1 {
		t.Errorf("a refresh with 11 frames the same and 1 new decoded %d, want exactly 1", got)
	}
	// The same bytes at a different time are a different frame.
	moved := loopOf(frames[1:]...)
	moved.Image.Frames[0].Valid = moved.Image.Frames[0].Valid.Add(-time.Minute)
	if _, err := s.HandIn(moved); err != nil {
		t.Fatal(err)
	}
	classesOf(t, s, "loop")
	if got := s.decodes - 13; got != 1 {
		t.Errorf("one frame moved in time: %d decoded, want 1", got)
	}
	// Two refreshes before any work: the second still keeps what the first
	// kept, and decodes only its own new frame.
	next := append(append([]LoopFrame(nil), frames[2:]...), frameAt(t, -1, storm))
	if _, err := s.HandIn(loopOf(frames[1:]...)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.HandIn(loopOf(next...)); err != nil {
		t.Fatal(err)
	}
	classesOf(t, s, "loop")
	if got := s.decodes - 14; got != 1 {
		t.Errorf("two refreshes before any work: %d decoded, want 1", got)
	}
	// A new table changes every reading.
	retabled := loopOf(frames[1:]...)
	retabled.Image.Table[1].Value = 60
	if _, err := s.HandIn(retabled); err != nil {
		t.Fatal(err)
	}
	classesOf(t, s, "loop")
	if got := s.decodes - 15; got != 12 {
		t.Errorf("a new table: %d decoded, want all 12", got)
	}
	// A new type changes every reading too.
	retyped := loopOf(frames[1:]...)
	retyped.Image.Table[1].Value = 60
	retyped.Image.Type = Type{Preset: "temperature", Unit: "C"}
	if _, err := s.HandIn(retyped); err != nil {
		t.Fatal(err)
	}
	classesOf(t, s, "loop")
	if got := s.decodes - 27; got != 12 {
		t.Errorf("a new type: %d decoded, want all 12", got)
	}
}

// TestTheKeyHoldsAValuesExactBits is L2.9 (L-12.6): two tables that differ
// only at a value too large for a fixed-point key get different keys.
func TestTheKeyHoldsAValuesExactBits(t *testing.T) {
	kind, err := ResolveType(Type{Preset: "radar", Unit: "dBZ"})
	if err != nil {
		t.Fatal(err)
	}
	a := loopOf().Image
	a.PNG = frameAt(t, 0, rain).PNG
	b := *a
	a.Table = []TableEntry{{Colour: colour.RGB{R: 200}, Value: 25}, {Colour: colour.RGB{G: 200}, Value: 1e18}}
	b.Table = []TableEntry{{Colour: colour.RGB{R: 200}, Value: 25}, {Colour: colour.RGB{G: 200}, Value: 2e18}}
	ka, _ := ImageKey(a, kind)
	kb, _ := ImageKey(&b, kind)
	if ka == kb {
		t.Error("two tables that differ at 1e18 share a key")
	}
	c := *a
	c.South = a.South + 1e-9
	if kc, _ := ImageKey(&c, kind); kc == ka {
		t.Error("two images a nanodegree apart share a key")
	}
}

// TestTheBudgetRefusesAndSaysByHowMuch is L2.6 (L-12.1, D-72): a hand-in
// over the budget is refused, saying by how much; lowering the budget below
// what is held drops nothing, and refuses the next hand-in that does not fit.
func TestTheBudgetRefusesAndSaysByHowMuch(t *testing.T) {
	s := store(t)
	loop := loopOf(frameAt(t, 1, rain), frameAt(t, 0, storm))
	need := imageCharge(loop.Image)
	s.SetBudget(need - 10)
	_, err := s.HandIn(loop)
	if !isKind(err, fault.OverImageCap) || !strings.Contains(err.Error(), "10 over") {
		t.Fatalf("a loop 10 bytes over the budget: %v; want a refusal that says it is 10 over", err)
	}
	s.SetBudget(need)
	if _, err := s.HandIn(loop); err != nil {
		t.Fatalf("a loop that fits exactly: %v", err)
	}
	classesOf(t, s, "loop")
	s.SetBudget(1)
	if _, _, ok := s.Raster("loop"); !ok {
		t.Error("lowering the budget dropped what was held")
	}
	if s.ImageUse() != need {
		t.Errorf("after lowering the budget the store counts %d, want the %d it held", s.ImageUse(), need)
	}
	other := loopOf(frameAt(t, 0, rain))
	other.ID = "second"
	if _, err := s.HandIn(other); !isKind(err, fault.OverImageCap) {
		t.Errorf("the next hand-in over a lowered budget: %v", err)
	}
	// A refresh of the same id is counted without the version it replaces.
	s.SetBudget(need)
	if _, err := s.HandIn(loop); err != nil {
		t.Errorf("a refresh of the same size under a budget that holds one: %v", err)
	}
}

// TestTheBudgetCountsEveryPart is L2.7 (L-12.4, L-12.6): the budget counts
// each frame's retained PNG bytes and its classified pixels, a grid's field,
// and the shared classified set, checked against a hand count.
func TestTheBudgetCountsEveryPart(t *testing.T) {
	shared := NewClassified(0)
	s, err := NewStore(Caps{Classified: shared})
	if err != nil {
		t.Fatal(err)
	}
	frames := []LoopFrame{frameAt(t, 2, rain), gapAt(1), frameAt(t, 0, storm)}
	if _, err := s.HandIn(loopOf(frames...)); err != nil {
		t.Fatal(err)
	}
	byHand := int64(len(frames[0].PNG) + len(frames[2].PNG) + 2*8*6) // two pictures of 8x6, one byte a pixel; the gap is nothing
	if got := s.ImageUse(); got != byHand {
		t.Errorf("a loop before any work: counted %d, by hand %d", got, byHand)
	}
	classesOf(t, s, "loop")
	held, _ := shared.Bytes()
	if held == 0 {
		t.Fatal("the shared set holds nothing, so this proves nothing")
	}
	byHand += held
	grid := Overlay{ID: "temps", Valid: noon, Keeps: time.Hour, Grid: &Grid{West: -90, South: 30, East: -80, North: 40, Cols: 5, Rows: 4, Values: make([]float64, 20), Type: Type{Preset: "temperature", Unit: "C"}}}
	if _, err := s.HandIn(grid); err != nil {
		t.Fatal(err)
	}
	byHand += 5 * 4
	if got := s.ImageUse(); got != byHand {
		t.Errorf("a loop, the shared set and a grid: counted %d, by hand %d", got, byHand)
	}
}

// TestAFramesFileIsCappedByItsPixels is L2.8 (L-12.4): a picture's file may
// be at most four bytes a pixel and 64 KiB; a padded one is refused, and an
// honest one of the same size passes.
func TestAFramesFileIsCappedByItsPixels(t *testing.T) {
	honest := picturePNG(t, 500, 500, func(x, y int) color.Color {
		if (x/50+y/50)%2 == 0 {
			return rain
		}
		return storm
	})
	if err := checkPNG(honest, defaultImageBytes); err != nil {
		t.Fatalf("an honest 250,000-pixel picture: %v", err)
	}
	padded := append(append([]byte(nil), honest...), make([]byte, 4*250_000+64<<10+1-len(honest))...)
	if err := checkPNG(padded, defaultImageBytes); !isKind(err, fault.OverImageCap) {
		t.Errorf("a 250,000-pixel picture padded one byte past four bytes a pixel and 64 KiB: %v", err)
	}
	if err := checkPNG(padded[:len(padded)-1], defaultImageBytes); err != nil {
		t.Errorf("padded to exactly the limit: %v", err)
	}
}

// TestPurgeDropsOnlyPicturesKeptForReuse is v0.2.0 L9.4 (L-9.3): a purge
// empties the shared set of readings and a replaced loop's spare frames, so
// the refresh after it decodes every frame again; the picture an overlay
// shows now is the host's, and stays.
func TestPurgeDropsOnlyPicturesKeptForReuse(t *testing.T) {
	var frames []LoopFrame
	for i := 12; i >= 0; i-- {
		c := rain
		if i%2 == 0 {
			c = storm
		}
		frames = append(frames, frameAt(t, i, c))
	}
	shared := NewClassified(0)
	s, err := NewStore(Caps{Classified: shared})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.HandIn(loopOf(frames[:12]...)); err != nil {
		t.Fatal(err)
	}
	classesOf(t, s, "loop")
	first := s.decodes
	if held, _ := shared.Bytes(); held == 0 || first == 0 {
		t.Fatal("nothing was decoded or shared, so this proves nothing")
	}
	s.Purge()
	if held, _ := shared.Bytes(); held != 0 {
		t.Errorf("the shared set holds %d bytes after a purge", held)
	}
	if _, _, ok := s.Raster("loop"); !ok {
		t.Error("a purge took away the picture the overlay shows")
	}
	// A refresh keeps its spare frames; a purge before its work drops them,
	// and it decodes each of the two pictures again, as the first hand-in
	// did. With the spare frames kept it would decode only the new frame.
	if _, err := s.HandIn(loopOf(frames[1:]...)); err != nil {
		t.Fatal(err)
	}
	s.Purge()
	classesOf(t, s, "loop")
	if got := s.decodes - first; got != first {
		t.Errorf("the refresh after a purge decoded %d, want %d, as the first hand-in did", got, first)
	}
}
