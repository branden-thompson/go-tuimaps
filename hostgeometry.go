package tuimaps

// This file is the seam between a host's own geometry and the library's.
//
// **A host keeps its shapes in its own type**, and should not have to change
// that to draw them. The first integration made the case: its point type
// carries JSON tags because it is also serialised, and its package sits below
// the renderer in its own layering, so it cannot name a library type at all.
// Anything that matched host points by memory layout would have refused it,
// and anything that demanded the library's own type would have inverted its
// dependencies.
//
// So a point is asked what it is, rather than inspected for what it looks
// like: field order, field names and struct tags all stay the host's business.

// Positioned is a host's own point: anything that can say where it is.
//
// The method answers in plain numbers and names nothing from this library, so
// a host's data model can satisfy it without importing the renderer - which is
// the whole point of it being a method rather than a shape.
type Positioned interface {
	// LonLat is the position in degrees, longitude first - the order
	// GeoJSON writes it in, and the order this library reads it in.
	LonLat() (lon, lat float64)
}

// Ring converts one run of a host's positions - a polygon's outline, one of
// its holes, or a line - into the library's own.
//
// **It copies.** The positions are read once, here, rather than on every
// frame: what the renderer then holds is its own and does not change under it
// if the host edits its copy. For the geometry this library is built for that
// is a few hundred kilobytes when an overlay changes, against nothing at all
// per frame.
func Ring[P Positioned](run []P) []LonLat {
	if len(run) == 0 {
		return nil
	}
	out := make([]LonLat, len(run))
	for i, p := range run {
		lon, lat := p.LonLat()
		out[i] = LonLat{Lon: lon, Lat: lat}
	}
	return out
}

// Rings converts one area of a host's own geometry: its outline first, then
// any holes in it, which is the order a Feature reads them in.
//
// **One area, not several.** A hazard that covers separate ground is several
// features, because within a feature the first ring is the outline and every
// ring after it is a hole - so pouring two areas into one Feature would make
// the second a hole in the first. A host with many areas calls this once per
// area and makes a Feature of each.
func Rings[P Positioned](area [][]P) [][]LonLat {
	if len(area) == 0 {
		return nil
	}
	out := make([][]LonLat, 0, len(area))
	for _, run := range area {
		out = append(out, Ring(run))
	}
	return out
}
