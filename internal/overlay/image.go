package overlay

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"image/color"
	"image/png"
	"math"
	"strconv"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// maxPixels is the most pixels an image may have, read from its header
	// before anything is decoded (FR-9).
	maxPixels = 1_048_576
	// maxPNGBytes bounds the file itself: a million pixels of four bytes, with room.
	maxPNGBytes = 8 << 20
	// defaultImageBytes is the map's image cap, at one byte a pixel (D-85, D-36).
	defaultImageBytes = 250_000
	// defaultImageBudget is what a map's images may hold in all (L-12.1, D-68).
	defaultImageBudget = 6 << 20
	// fileSlack is what a picture's file may carry beyond four bytes a pixel (L-12.4).
	fileSlack = 64 << 10
	// maxTable, defaultTolerance, maxTolerance and maxSamples are FR-9's bounds.
	maxTable         = 256
	defaultTolerance = 10.0
	maxTolerance     = 25.0
	maxSamples       = 16
)

// MaxFrames is the most frames a loop may have, gaps included (L-1.15, D-68).
const MaxFrames = 72

// Projection is the projection a host says its image is in. There are two,
// and an image in any other is refused.
type Projection uint8

// The projections an image may be in.
const (
	PlateCarree Projection = iota + 1 // equal steps of longitude and latitude
	WebMercator                       // as web map tiles are
)

// TableEntry is one line of an image's colour table: a colour the provider
// draws, and the value it stands for - or that it stands for "missing".
type TableEntry struct {
	Colour  colour.RGB
	Value   float64
	Missing bool
}

// Image is a georeferenced image as a host hands it in (FR-9): one PNG for
// one bounding box, in a stated projection, with the table that says what
// its colours mean. The table is required: the provider's own colours are
// never shown, so without it there is nothing to draw.
type Image struct {
	PNG                      []byte
	Frames                   []LoopFrame // a loop; a single picture is PNG with no frames (L-1.1)
	West, South, East, North float64
	Projection               Projection
	Table                    []TableEntry
	Tolerance                float64 // a colour difference from 0 to 25; zero means the default, 10
	Exact                    bool    // match the table's colours exactly, with no tolerance at all
	Type                     Type
}

// LoopFrame is one frame of a loop: when its picture was valid, and the
// picture. A gap is a frame that is missing, stated as missing and never
// filled from a neighbour (L-1.2); it has no bytes. A forecast frame is one
// the provider forecast rather than observed (D-67).
type LoopFrame struct {
	Valid    time.Time
	PNG      []byte
	Gap      bool
	Forecast bool
}

// Report is what matching an image's colours found (FR-9): how many pixels
// matched nothing, and at most sixteen of their colours, so that a changed
// palette is visible and not silently wrong.
type Report struct {
	Unmatched int
	Samples   []colour.RGB
}

func imageRefused(why, todo textsafe.Text) error {
	return fault.Make(fault.ImageRefused, textsafe.Const("the image was refused"), why, todo)
}

// header reads a PNG's size and depth from its first 33 bytes, which is all
// that is looked at before the image is accepted.
func header(b []byte) (w, h uint32, depth byte, err error) {
	if len(b) < 33 || string(b[:8]) != "\x89PNG\r\n\x1a\n" || binary.BigEndian.Uint32(b[8:]) != 13 || string(b[12:16]) != "IHDR" {
		return 0, 0, 0, imageRefused(textsafe.Const("it is not a PNG"), textsafe.Const("hand in a PNG; no other format is read"))
	}
	return binary.BigEndian.Uint32(b[16:]), binary.BigEndian.Uint32(b[20:]), b[24], nil
}

// checkPNG holds the file to its limits from its header alone: its length,
// its size in pixels, its depth, and the map's image cap at a byte a pixel.
func checkPNG(file []byte, imageCap int) error {
	if len(file) > maxPNGBytes {
		return refused(fault.OverImageCap, textsafe.Const("its PNG is larger than 8 MiB"), textsafe.Const("hand in a smaller image"))
	}
	w, h, depth, err := header(file)
	if err != nil {
		return err
	}
	if w == 0 || h == 0 {
		return imageRefused(textsafe.Const("its header gives it no size"), textsafe.Const("check the image"))
	}
	if depth != 8 {
		return imageRefused(textsafe.Const("it is not 8 bits a channel; a 16-bit image, or one of fewer bits, is not read"), textsafe.Const("hand in an 8-bit PNG"))
	}
	pixels := uint64(w) * uint64(h)
	if w > maxPixels || h > maxPixels || pixels > maxPixels {
		return refused(fault.OverImageCap, textsafe.Const("it has more than 1,048,576 pixels"), textsafe.Const("hand in a smaller image: a terminal map shows a few thousand cells"))
	}
	if uint64(len(file)) > 4*pixels+fileSlack {
		return refused(fault.OverImageCap,
			textsafe.Join(textsafe.Const("its file is "), textsafe.Clean(grouped(len(file))), textsafe.Const(" bytes, more than four bytes a pixel and 64 KiB for its "), textsafe.Clean(grouped(int(pixels))), textsafe.Const(" pixels")),
			textsafe.Const("hand in the picture without the padding or extra chunks it carries"))
	}
	if pixels > uint64(imageCap) {
		return refused(fault.OverImageCap,
			textsafe.Join(textsafe.Const("at one byte a pixel it is "), textsafe.Clean(grouped(int(pixels))), textsafe.Const(" bytes, over the map's image cap of "), textsafe.Clean(grouped(imageCap))),
			textsafe.Const("hand in an image of fewer pixels than the cap has bytes"))
	}
	return nil
}

// checkImage validates an image on hand-in, before a byte of it is decoded.
func checkImage(img *Image, imageCap int) (Kind, error) {
	if img == nil {
		return Kind{}, imageRefused(textsafe.Const("there is no image"), textsafe.Const("hand in a PNG"))
	}
	err := checkPictures(img, imageCap)
	if err != nil {
		return Kind{}, err
	}
	if img.Projection != PlateCarree && img.Projection != WebMercator {
		return Kind{}, imageRefused(textsafe.Const("its projection is not stated, or is not one the library has"),
			textsafe.Const("say whether the image is in equal steps of longitude and latitude, or in the web map projection"))
	}
	if !(img.West >= -180 && img.East <= 180 && img.South >= -90 && img.North <= 90) || !(img.West < img.East && img.South < img.North) {
		return Kind{}, refused(fault.InvalidCoordinates, textsafe.Const("its bounds are not numbers on the globe, or its west is not west of its east, or its south not south of its north"),
			textsafe.Const("give west, south, east and north in degrees"))
	}
	if !img.Exact && !(img.Tolerance >= 0 && img.Tolerance <= maxTolerance) {
		return Kind{}, fault.Make(fault.MalformedTable, textsafe.Const("the image was refused"),
			textsafe.Const("its colour tolerance is not a difference from 0 to 25"), textsafe.Const("give a tolerance from 0 to 25, or none for the default of 10"))
	}
	err = CheckTable(img.Table)
	if err != nil {
		return Kind{}, err
	}
	return ResolveType(img.Type)
}

// checkPictures holds an image's picture, or every frame of its loop, to the
// limits (L-1.9): times that rise, each once; a gap with no bytes and a
// picture frame with some; and at least one picture. A refusal names the
// frame, counting from one. The frame limit is held by copied, before this.
func checkPictures(img *Image, imageCap int) error {
	if len(img.Frames) == 0 {
		return checkPNG(img.PNG, imageCap)
	}
	if len(img.PNG) != 0 {
		return imageRefused(textsafe.Const("it has both a single picture and frames"),
			textsafe.Const("hand in PNG for one picture, or Frames for a loop, not both"))
	}
	pictures := 0
	for i, f := range img.Frames {
		frame := textsafe.Join(textsafe.Const("frame "), textsafe.Clean(strconv.Itoa(i+1)))
		var err error
		switch {
		case f.Valid.IsZero():
			err = imageRefused(textsafe.Const("it has no valid time"), textsafe.Const("give every frame the time its picture was valid"))
		case i > 0 && !f.Valid.After(img.Frames[i-1].Valid):
			err = imageRefused(textsafe.Const("its valid time is not after the frame before it"), textsafe.Const("hand the frames in oldest first, each time once"))
		case f.Gap && len(f.PNG) != 0:
			err = imageRefused(textsafe.Const("it is a gap and has a picture; a gap has no bytes"), textsafe.Const("hand in a missing frame as a gap with no picture, or as a picture frame"))
		case !f.Gap && len(f.PNG) == 0:
			err = imageRefused(textsafe.Const("it has no picture; a missing frame is a gap, and is stated as one"), textsafe.Const("mark a missing frame as a gap"))
		case !f.Gap:
			err = checkPNG(f.PNG, imageCap)
			pictures++
		}
		var said *fault.Error
		if errors.As(err, &said) {
			return said.Of(frame)
		}
		if err != nil {
			return err
		}
	}
	if pictures == 0 {
		return imageRefused(textsafe.Const("every frame of it is a gap"), textsafe.Const("hand in at least one frame with a picture"))
	}
	return nil
}

// copied is the store's own copy of an image: its picture, its table, and
// every frame's picture, so that the host's slices are never read after Set
// returns (L-1.14). A loop over the frame limit, MaxFrames gaps included, is
// refused before anything is copied (L-1.15).
func copied(img *Image) (*Image, error) {
	if img == nil {
		return nil, nil
	}
	if len(img.Frames) > MaxFrames {
		return nil, refused(fault.OverImageCap,
			textsafe.Join(textsafe.Const("it has "), textsafe.Clean(strconv.Itoa(len(img.Frames))), textsafe.Const(" frames, over the most a loop may have, 72, gaps included")),
			textsafe.Const("hand in fewer frames"))
	}
	own := *img
	own.PNG = cloneBytes(img.PNG)
	own.Table = append([]TableEntry(nil), img.Table...)
	if img.Frames != nil {
		own.Frames = make([]LoopFrame, len(img.Frames))
	}
	for i, f := range img.Frames {
		f.PNG = cloneBytes(f.PNG)
		own.Frames[i] = f
	}
	return &own, nil
}

// imageCharge is what an image costs the map's budget (L-12.4): every
// picture's retained file, and its classified pixels at one byte a pixel,
// read from the header it was checked by. A gap costs nothing.
func imageCharge(img *Image) int64 {
	if img == nil {
		return 0
	}
	one := func(file []byte) int64 {
		w, h, _, err := header(file)
		if err != nil {
			return int64(len(file))
		}
		return int64(len(file)) + int64(w)*int64(h)
	}
	if len(img.Frames) == 0 {
		return one(img.PNG)
	}
	total := int64(0)
	for _, f := range img.Frames {
		if !f.Gap {
			total += one(f.PNG)
		}
	}
	return total
}

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	return append([]byte(nil), b...)
}

// CheckTable holds a colour table to FR-9: required, at most 256 entries, no
// colour twice, values that are numbers and never fall.
func CheckTable(table []TableEntry) error {
	if len(table) == 0 {
		return fault.Make(fault.MissingTable, textsafe.Const("the image was refused"),
			textsafe.Const("it has no colour table, and a provider's own colours are never shown"),
			textsafe.Const("hand in the table that says what value each of the image's colours stands for"))
	}
	bad := func(why textsafe.Text) error {
		return fault.Make(fault.MalformedTable, textsafe.Const("the image's colour table was refused"), why,
			textsafe.Const("give at most 256 entries, each colour once, with values that are numbers and never fall"))
	}
	if len(table) > maxTable {
		return bad(textsafe.Const("it has more than 256 entries"))
	}
	seen := make(map[colour.RGB]bool, len(table))
	last, values := math.Inf(-1), 0
	for _, e := range table {
		if seen[e.Colour] {
			return bad(textsafe.Const("one colour appears in it twice"))
		}
		seen[e.Colour] = true
		if e.Missing {
			continue
		}
		if !(e.Value >= -1.797e308 && e.Value <= 1.797e308) || e.Value < last {
			return bad(textsafe.Const("one of its values is not a number, or is lower than the one before it"))
		}
		last = e.Value
		values++
	}
	if values == 0 {
		return bad(textsafe.Const("none of its entries has a value"))
	}
	return nil
}

// matcher turns a pixel's colour into a class, remembering each colour it
// has met so that an image of few colours costs few searches.
type matcher struct {
	table     []TableEntry
	breaks    []float64
	tolerance float64
	known     map[colour.RGB]int8
	report    Report
	sampled   map[colour.RGB]bool
}

const unmatched = int8(-2)

func (m *matcher) class(c colour.RGB) int8 {
	if got, ok := m.known[c]; ok {
		return got
	}
	best, bestGap := -1, math.Inf(1)
	for i, e := range m.table {
		if e.Colour == c {
			best = i
			break
		}
		if m.tolerance > 0 {
			if gap := colour.InLab(c, colour.Normal).Distance(colour.InLab(e.Colour, colour.Normal)); gap <= m.tolerance && gap < bestGap {
				best, bestGap = i, gap
			}
		}
	}
	got := unmatched
	if best >= 0 {
		got = int8(NoData)
		if !m.table[best].Missing {
			got = int8(Classify(m.table[best].Value, m.breaks))
		}
	}
	m.known[c] = got
	return got
}

// pixel classifies one pixel. A fully transparent pixel is nothing at all,
// and is not counted as unmatched.
func (m *matcher) pixel(c color.NRGBA) int8 {
	if c.A == 0 {
		return NoData
	}
	rgb := colour.RGB{R: c.R, G: c.G, B: c.B}
	got := m.class(rgb)
	if got != unmatched {
		return got
	}
	m.report.Unmatched++
	if !m.sampled[rgb] && len(m.report.Samples) < maxSamples {
		m.sampled[rgb] = true
		m.report.Samples = append(m.report.Samples, rgb)
	}
	return NoData
}

// rasterise decodes one picture of an image - directly, never through a
// registry of formats - and turns its colours into classes, one byte a pixel.
// It reads the size from the picture's own header and re-checks the caps
// before the decoder allocates anything (L-1.14).
func rasterise(img *Image, file []byte, kind Kind, imageCap int) (scene.Raster, Report, error) {
	if img == nil {
		return scene.Raster{}, Report{}, imageRefused(textsafe.Const("there is no image"), textsafe.Const("hand in a PNG"))
	}
	err := checkPNG(file, imageCap)
	if err != nil {
		return scene.Raster{}, Report{}, err
	}
	decoded, err := png.Decode(bytes.NewReader(file))
	if err != nil {
		return scene.Raster{}, Report{}, imageRefused(textsafe.Const("its header is a PNG's and the rest of it could not be decoded"), textsafe.Const("check the image; it may have been cut short"))
	}
	bounds := decoded.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 || w*h > maxPixels {
		return scene.Raster{}, Report{}, imageRefused(textsafe.Const("its size as decoded is not the size its header gave"), textsafe.Const("check the image"))
	}
	tolerance := img.Tolerance
	if tolerance == 0 {
		tolerance = defaultTolerance
	}
	if img.Exact {
		tolerance = 0
	}
	m := &matcher{table: img.Table, breaks: kind.Breaks, tolerance: tolerance, known: map[colour.RGB]int8{}, sampled: map[colour.RGB]bool{}}
	raster := scene.Raster{West: img.West, South: img.South, East: img.East, North: img.North, Projection: uint8(img.Projection),
		Width: w, Height: h, Classes: make([]int8, w*h), Preset: uint8(kind.Preset), ClassCount: len(kind.Breaks) + 1}
	for y := range h {
		for x := range w {
			raster.Classes[y*w+x] = m.pixel(color.NRGBAModel.Convert(decoded.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.NRGBA))
		}
	}
	return raster, m.report, nil
}

// readPicture is the reading of one picture: the set every map of a shared
// group draws from, where there is one, and the work itself where there is
// not. Two maps showing the same radar frame then decode it once between
// them rather than once each (D-116).
func (s *Store) readPicture(img *Image, file []byte, kind Kind) (scene.Raster, Report, error) {
	if s == nil {
		return scene.Raster{}, Report{}, imageRefused(textsafe.Const("there is no store to read a picture into"), textsafe.Const("this is a defect in the library; report it"))
	}
	one := *img
	one.PNG, one.Frames = file, nil
	key, keyed := ImageKey(&one, kind)
	if keyed {
		if raster, report, ok := s.caps.Classified.Read(key); ok {
			return raster, report, nil
		}
	}
	s.mu.Lock()
	s.decodes++
	s.mu.Unlock()
	raster, report, err := rasterise(img, file, kind, s.caps.ImageBytes)
	if err == nil && keyed {
		s.caps.Classified.Keep(key, raster, report)
	}
	return raster, report, err
}

// picture is one decoded picture of an image: its single picture, or one
// frame of its loop. A gap is never decoded, and stays not ready.
type picture struct {
	key    [32]byte
	raster scene.Raster
	report Report
	ready  bool
}

// shown is the frame an image shows: its single picture, or the newest
// observed frame that is not a gap - "right now", where a loop opens (D-67) -
// or, with no observed frame, the newest that is not a gap.
func shown(img *Image) int {
	if img == nil || len(img.Frames) == 0 {
		return 0
	}
	newest := -1
	for i := len(img.Frames) - 1; i >= 0; i-- {
		f := img.Frames[i]
		if f.Gap {
			continue
		}
		if !f.Forecast {
			return i
		}
		if newest < 0 {
			newest = i
		}
	}
	return newest
}

// Raster is the picture an overlay shows and what matching its colours
// found, once a job has decoded it.
func (s *Store) Raster(id string) (scene.Raster, Report, bool) {
	if s == nil || id == "" {
		return scene.Raster{}, Report{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.current[id]
	if !ok {
		return scene.Raster{}, Report{}, false
	}
	pictures, at := s.pictures[id], shown(h.overlay.Image)
	if at < 0 || at >= len(pictures) || !pictures[at].ready {
		return scene.Raster{}, Report{}, false
	}
	return pictures[at].raster, pictures[at].report, true
}

// readPictures decodes every picture of an image: its single picture, or
// each frame of its loop that is not a gap. A frame whose key the overlay's
// last version already decoded is kept, not decoded again (L-1.7).
func (s *Store) readPictures(ctx context.Context, id string, img *Image, kind Kind) ([]picture, error) {
	if len(img.Frames) == 0 {
		raster, report, err := s.readPicture(img, img.PNG, kind)
		return []picture{{raster: raster, report: report, ready: err == nil}}, err
	}
	pictures := make([]picture, len(img.Frames))
	for i, f := range img.Frames {
		if f.Gap {
			continue
		}
		if ctx.Err() != nil {
			return nil, fault.Make(fault.Cancelled, textsafe.Const("reading a loop was abandoned"), textsafe.Const("the work it was part of was cancelled or ran out of time"), textsafe.Const("nothing; it is read again if it is still wanted"))
		}
		key := frameKey(img, f.PNG, f.Valid, kind)
		if kept, ok := s.spareFrame(id, key); ok {
			pictures[i] = kept
			continue
		}
		raster, report, err := s.readPicture(img, f.PNG, kind)
		if err != nil {
			return nil, err
		}
		pictures[i] = picture{key: key, raster: raster, report: report, ready: true}
	}
	return pictures, nil
}

// spareFrame is a decoded frame of the overlay's replaced version, if it had
// one with this key.
func (s *Store) spareFrame(id string, key [32]byte) (picture, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.spare[id][key]
	return p, ok
}

// keepPictures stores an image's decoded pictures, if the image they were
// made from is still the overlay's, and warns of colours that matched nothing.
func (s *Store) keepPictures(r *Reader, pictures []picture) {
	if r == nil || r.h == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.h.retired {
		return
	}
	s.pictures[r.id] = pictures
	s.landed++
	delete(s.spare, r.id) // what the new version kept, it now holds
	unmatched := 0
	for _, p := range pictures {
		unmatched += p.report.Unmatched
	}
	if unmatched > 0 {
		s.warnCountLocked(fault.UnmatchedImageColours, textsafe.Quote(r.id), unmatched)
	}
}
