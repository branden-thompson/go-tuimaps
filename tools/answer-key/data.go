package main

import "math"

// Field is a scalar field in a scenario: values on a regular grid, rows
// from the north, each row west to east, with the breaks that divide them
// into bands. The breaks are written in the file rather than taken from the
// library, so that this program owes it nothing (D-43).
type Field struct {
	ID                       string    `json:"id"`
	West, South, East, North float64   `json:"-"`
	W                        float64   `json:"west"`
	S                        float64   `json:"south"`
	E                        float64   `json:"east"`
	N                        float64   `json:"north"`
	Cols                     int       `json:"cols"`
	Rows                     int       `json:"rows"`
	Values                   []float64 `json:"values"`
	Breaks                   []float64 `json:"breaks"`
}

// Picture is a classified image in a scenario: one class a pixel, rows from
// the north, with the box it covers.
type Picture struct {
	ID      string  `json:"id"`
	W       float64 `json:"west"`
	S       float64 `json:"south"`
	E       float64 `json:"east"`
	N       float64 `json:"north"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Classes []int   `json:"classes"`
}

// FieldAnswer is what a field says at a place.
type FieldAnswer struct {
	Place  string  `json:"place"`
	Field  string  `json:"field"`
	Value  float64 `json:"value"`
	Band   int     `json:"band"`
	Rises  string  `json:"rises,omitempty"`
	Flat   bool    `json:"flat,omitempty"`
	NoData bool    `json:"no_data,omitempty"`
}

// PictureAnswer is what an image says at a place.
type PictureAnswer struct {
	Place       string  `json:"place"`
	Image       string  `json:"image"`
	Class       int     `json:"class"`
	NoData      bool    `json:"no_data,omitempty"`
	HeavierKm   float64 `json:"heavier_km,omitempty"`
	HeavierWay  string  `json:"heavier_way,omitempty"`
	HeavierWhat int     `json:"heavier_class,omitempty"`
	NoneHeavier bool    `json:"none_heavier,omitempty"`
}

// atField is the field's answer at a place: the value of the sample the
// place falls in, its band, and which way the values rise - worked out by
// comparing the four neighbours, each step written out.
func atField(p Place, f Field) FieldAnswer {
	out := FieldAnswer{Place: p.Name, Field: f.ID}
	col, row, ok := sampleOf(p, f)
	if !ok {
		out.NoData = true
		return out
	}
	here, ok := sample(f, col, row)
	if !ok {
		out.NoData = true
		return out
	}
	out.Value, out.Band = round(here, 3), bandOf(here, f.Breaks)
	x, y := 0.0, 0.0
	if v, ok := sample(f, col+1, row); ok {
		x += v - here
	}
	if v, ok := sample(f, col-1, row); ok {
		x -= v - here
	}
	if v, ok := sample(f, col, row-1); ok {
		y += v - here
	}
	if v, ok := sample(f, col, row+1); ok {
		y -= v - here
	}
	if x == 0 && y == 0 {
		out.Flat = true
		return out
	}
	out.Rises = compass(math.Mod(math.Atan2(x, y)*180/math.Pi+360, 360))
	return out
}

// sampleOf is the column and row a place falls in.
func sampleOf(p Place, f Field) (int, int, bool) {
	if f.Cols <= 0 || f.Rows <= 0 || f.N <= f.S {
		return 0, 0, false
	}
	span := east(f.W, f.E)
	if span <= 0 {
		span += 360
	}
	from := east(f.W, p.Lon)
	if from < 0 {
		from += 360
	}
	if from > span || p.Lat < f.S || p.Lat > f.N {
		return 0, 0, false
	}
	col := int(from / span * float64(f.Cols))
	row := int((f.N - p.Lat) / (f.N - f.S) * float64(f.Rows))
	if col >= f.Cols {
		col = f.Cols - 1
	}
	if row >= f.Rows {
		row = f.Rows - 1
	}
	return col, row, true
}

// sample is one value of a field, or false where there is none.
func sample(f Field, col, row int) (float64, bool) {
	if col < 0 || row < 0 || col >= f.Cols || row >= f.Rows {
		return 0, false
	}
	at := row*f.Cols + col
	if at >= len(f.Values) {
		return 0, false
	}
	v := f.Values[at]
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

// bandOf is how many breaks a value is at or above.
func bandOf(v float64, breaks []float64) int {
	n := 0
	for _, b := range breaks {
		if v >= b {
			n++
		}
	}
	return n
}

// atPicture is the image's answer at a place: the class here, and the
// nearest pixel of a heavier one, found by looking at every pixel.
func atPicture(p Place, pic Picture) PictureAnswer {
	out := PictureAnswer{Place: p.Name, Image: pic.ID}
	here, ok := classOf(p, pic)
	if !ok {
		out.NoData, here = true, -1
	}
	out.Class = here
	if out.NoData {
		out.Class = 0
	}
	bestKm, bestWay, bestClass, found := math.Inf(1), "", 0, false
	for y := range pic.Height {
		for x := range pic.Width {
			at := y*pic.Width + x
			if at >= len(pic.Classes) || pic.Classes[at] <= here {
				continue
			}
			lon, lat := pixelCentre(pic, x, y)
			km := haversine(p.Lat, p.Lon, lat, lon)
			if km < bestKm {
				bestKm, bestWay, bestClass, found = km, compass(bearingTo(p.Lat, p.Lon, lat, lon)), pic.Classes[at], true
			}
		}
	}
	if !found {
		out.NoneHeavier = true
		return out
	}
	out.HeavierKm, out.HeavierWay, out.HeavierWhat = round(bestKm, 1), bestWay, bestClass
	return out
}

// classOf is the class of the pixel a place falls in.
func classOf(p Place, pic Picture) (int, bool) {
	col, row, ok := sampleOf(p, Field{W: pic.W, S: pic.S, E: pic.E, N: pic.N, Cols: pic.Width, Rows: pic.Height})
	if !ok {
		return 0, false
	}
	at := row*pic.Width + col
	if at >= len(pic.Classes) || pic.Classes[at] < 0 {
		return 0, false
	}
	return pic.Classes[at], true
}

// pixelCentre is the middle of one pixel.
func pixelCentre(pic Picture, x, y int) (float64, float64) {
	span := east(pic.W, pic.E)
	if span <= 0 {
		span += 360
	}
	lon := pic.W + span*(float64(x)+0.5)/float64(pic.Width)
	lat := pic.N - (pic.N-pic.S)*(float64(y)+0.5)/float64(pic.Height)
	return math.Mod(lon+540, 360) - 180, lat
}
