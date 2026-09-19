package project

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// EarthRadiusKm is the mean radius of the Earth. Distances are great-circle
// distances on a sphere of this radius: within about 0.5% of the ellipsoid,
// and finer than a terminal cell can show.
const EarthRadiusKm = 6371.0088

// onGlobe reports whether p is a position on the globe: both numbers
// finite, the latitude between the poles.
func onGlobe(p LonLat) error {
	if !(p.Lon >= -math.MaxFloat64 && p.Lon <= math.MaxFloat64) { // NaN and both infinities fail this
		return notFinite()
	}
	if !(p.Lat >= -90 && p.Lat <= 90) {
		return notFinite()
	}
	return nil
}

// GreatCircleKm is the shortest distance over the globe between a and b, in
// kilometres. Longitude is circular: the seam at 180 degrees is not an edge.
func GreatCircleKm(a, b LonLat) (float64, error) {
	err := onGlobe(a)
	if err != nil {
		return 0, err
	}
	err = onGlobe(b)
	if err != nil {
		return 0, err
	}
	const rad = math.Pi / 180
	sinLat := math.Sin((b.Lat - a.Lat) * rad / 2)
	sinLon := math.Sin((b.Lon - a.Lon) * rad / 2)
	h := sinLat*sinLat + math.Cos(a.Lat*rad)*math.Cos(b.Lat*rad)*sinLon*sinLon
	return 2 * EarthRadiusKm * math.Asin(math.Sqrt(math.Min(1, h))), nil
}

// Bearing is the direction in which b lies from a at the start of the
// shortest way there: degrees clockwise from north, from 0 up to 360.
func Bearing(a, b LonLat) (float64, error) {
	err := onGlobe(a)
	if err != nil {
		return 0, err
	}
	err = onGlobe(b)
	if err != nil {
		return 0, err
	}
	const rad = math.Pi / 180
	dLon := (b.Lon - a.Lon) * rad
	y := math.Sin(dLon) * math.Cos(b.Lat*rad)
	x := math.Cos(a.Lat*rad)*math.Sin(b.Lat*rad) - math.Sin(a.Lat*rad)*math.Cos(b.Lat*rad)*math.Cos(dLon)
	return math.Mod(math.Atan2(y, x)/rad+360, 360), nil
}

// Compass gives a bearing as one of eight plain words, each covering 45
// degrees centred on its direction: 342 degrees is "north". The words are
// written in full, so a speech engine reads them as words.
func Compass(bearing float64) (textsafe.Text, error) {
	if !(bearing >= -math.MaxFloat64 && bearing <= math.MaxFloat64) {
		return textsafe.Text{}, notFinite()
	}
	turn := math.Mod(math.Mod(bearing, 360)+360, 360)
	switch int(math.Floor((turn+22.5)/45)) % 8 {
	case 0:
		return textsafe.Const("north"), nil
	case 1:
		return textsafe.Const("north-east"), nil
	case 2:
		return textsafe.Const("east"), nil
	case 3:
		return textsafe.Const("south-east"), nil
	case 4:
		return textsafe.Const("south"), nil
	case 5:
		return textsafe.Const("south-west"), nil
	case 6:
		return textsafe.Const("west"), nil
	}
	return textsafe.Const("north-west"), nil
}

// CellSpanKm is the ground one cell covers at the view's centre: a column's
// width east to west, and a row's height north to south. A cell is twice as
// tall as it is wide, so a row covers about twice the ground of a column;
// "on the edge" is judged with the one that matches the bearing (D-67).
func (v View) CellSpanKm() (col, row float64, err error) {
	err = v.Validate()
	if err != nil {
		return 0, 0, err
	}
	cx, cy := float64(v.Cols*DotsPerCol)/2, float64(v.Rows*DotsPerRow)/2
	east, err := v.FromDot(cx+DotsPerCol, cy)
	if err != nil {
		return 0, 0, err
	}
	south, err := v.FromDot(cx, cy+DotsPerRow)
	if err != nil {
		return 0, 0, err
	}
	col, err = GreatCircleKm(v.Centre, east)
	if err != nil {
		return 0, 0, err
	}
	row, err = GreatCircleKm(v.Centre, south)
	if err != nil {
		return 0, 0, err
	}
	return col, row, nil
}
