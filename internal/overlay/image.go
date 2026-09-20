package overlay

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"image/png"
	"math"

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
	// maxTable, defaultTolerance, maxTolerance and maxSamples are FR-9's bounds.
	maxTable         = 256
	defaultTolerance = 10.0
	maxTolerance     = 25.0
	maxSamples       = 16
)

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
	West, South, East, North float64
	Projection               Projection
	Table                    []TableEntry
	Tolerance                float64 // a colour difference from 0 to 25; zero means the default, 10
	Exact                    bool    // match the table's colours exactly, with no tolerance at all
	Type                     Type
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
	err := checkPNG(img.PNG, imageCap)
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

// rasterise decodes an image - directly, never through a registry of formats -
// and turns its colours into classes, one byte a pixel.
func rasterise(img *Image, kind Kind) (scene.Raster, Report, error) {
	if img == nil {
		return scene.Raster{}, Report{}, imageRefused(textsafe.Const("there is no image"), textsafe.Const("hand in a PNG"))
	}
	decoded, err := png.Decode(bytes.NewReader(img.PNG))
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
func (s *Store) readPicture(img *Image, kind Kind) (scene.Raster, Report, error) {
	if s == nil {
		return scene.Raster{}, Report{}, imageRefused(textsafe.Const("there is no store to read a picture into"), textsafe.Const("this is a defect in the library; report it"))
	}
	key, keyed := ImageKey(img, kind)
	if keyed {
		if raster, report, ok := s.caps.Classified.Read(key); ok {
			return raster, report, nil
		}
	}
	raster, report, err := rasterise(img, kind)
	if err == nil && keyed {
		s.caps.Classified.Keep(key, raster, report)
	}
	return raster, report, err
}

// Raster is an overlay's prepared image and what matching its colours found,
// once a job has decoded it.
func (s *Store) Raster(id string) (scene.Raster, Report, bool) {
	if s == nil || id == "" {
		return scene.Raster{}, Report{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rasters[id]
	return r, s.reports[id], ok
}

// keepRaster stores a prepared image, if the image it was made from is still
// the overlay's, and warns of colours that matched nothing.
func (s *Store) keepRaster(r *Reader, raster scene.Raster, report Report) {
	if r == nil || r.h == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.h.retired {
		return
	}
	s.rasters[r.id], s.reports[r.id] = raster, report
	if report.Unmatched > 0 {
		s.warnCountLocked(fault.UnmatchedImageColours, textsafe.Quote(r.id), report.Unmatched)
	}
}
