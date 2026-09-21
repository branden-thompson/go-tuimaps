package project

import (
	"math"
	"sort"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// MaxViewZoom is the closest a view zooms in, as upstream (P-55).
	MaxViewZoom = 18
	// maxFitItems bounds what one fit-to may be asked to frame.
	maxFitItems = 100000
)

// Box is a bounding box in degrees. West greater than East means the box
// crosses the seam at 180 degrees. An overlay's box is recorded when it is
// handed in, so fit-to needs no Work to have run (D-76).
type Box struct {
	West  float64
	South float64
	East  float64
	North float64
}

// arc is a stretch of longitude, from start eastward, in degrees; start is
// from 0 up to 360. A point is an arc of no width.
type arc struct {
	start float64
	width float64
}

// badFit is the error for a fit-to that cannot be done.
func badFit() error {
	return fault.Make(fault.InvalidCoordinates,
		textsafe.Const("fit-to cannot frame what was named"),
		textsafe.Const("the margin is negative or leaves no room in the view, a box has its south above its north, or too many things were named"),
		textsafe.Const("check the margin against the view's size, and the boxes"))
}

// FitTo returns the view, of the same size as current, that frames every
// point and box with margin cells to spare on each side, at the largest
// zoom that does - solved directly, not stepped. A single point is centred
// at the zoom the view already has; nothing named changes nothing. The
// short way round is taken: longitude is circular.
func Frame(current View, points []LonLat, boxes []Box, margin int) (View, error) {
	err := current.Validate()
	if err != nil {
		return View{}, err
	}
	if margin < 0 || len(points)+len(boxes) > maxFitItems {
		return View{}, badFit()
	}
	roomX := float64((current.Cols - 2*margin) * DotsPerCol)
	roomY := float64((current.Rows - 2*margin) * DotsPerRow)
	if roomX <= 0 || roomY <= 0 {
		return View{}, badFit()
	}
	arcs, south, north, err := extents(points, boxes)
	if err != nil {
		return View{}, err
	}
	if len(arcs) == 0 {
		return current, nil
	}
	west, width := shortestCover(arcs)
	_, top, err := ToTile(LonLat{Lat: north}, 0)
	if err != nil {
		return View{}, err
	}
	_, bottom, err := ToTile(LonLat{Lat: south}, 0)
	if err != nil {
		return View{}, err
	}
	centre, err := FromTile(0.5, (top+bottom)/2, 0)
	if err != nil {
		return View{}, err
	}
	centre.Lon = math.Mod(west+width/2+540, 360) - 180
	fitted := current
	fitted.Centre = centre
	spanX, spanY := width/360, bottom-top
	if spanX > 0 || spanY > 0 { // otherwise a single point: the zoom stays
		zoom := math.Min(math.Log2(roomX/(TileSize*spanX)), math.Log2(roomY/(TileSize*spanY)))
		fitted.Zoom = math.Max(MinViewZoom, math.Min(MaxViewZoom, zoom))
	}
	return fitted, nil
}

// extents gathers the stretches of longitude that points and boxes cover,
// and the range of latitude.
func extents(points []LonLat, boxes []Box) (arcs []arc, south, north float64, err error) {
	south, north = 90, -90
	arcs = make([]arc, 0, len(points)+len(boxes))
	for _, p := range points {
		err = onGlobe(p)
		if err != nil {
			return nil, 0, 0, err
		}
		arcs = append(arcs, arc{start: math.Mod(math.Mod(p.Lon, 360)+360, 360)})
		south, north = math.Min(south, p.Lat), math.Max(north, p.Lat)
	}
	for _, b := range boxes {
		err = onGlobe(LonLat{Lon: b.West, Lat: b.South})
		if err != nil {
			return nil, 0, 0, err
		}
		err = onGlobe(LonLat{Lon: b.East, Lat: b.North})
		if err != nil {
			return nil, 0, 0, err
		}
		if b.South > b.North {
			return nil, 0, 0, badFit()
		}
		start := math.Mod(math.Mod(b.West, 360)+360, 360)
		arcs = append(arcs, arc{start: start, width: math.Mod(math.Mod(b.East, 360)+360-start+360, 360)})
		south, north = math.Min(south, b.South), math.Max(north, b.North)
	}
	return arcs, south, north, nil
}

// shortestCover returns the shortest stretch of longitude that covers every
// arc: where it starts, in degrees from 0 up to 360, and how wide it is. It
// is the whole circle less the widest stretch that nothing covers.
func shortestCover(arcs []arc) (west, width float64) {
	if len(arcs) == 0 {
		return 0, 0
	}
	sorted := append([]arc(nil), arcs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start < sorted[j].start })
	reach := sorted[0].start + sorted[0].width
	gap, after := 0.0, sorted[0].start
	for _, a := range sorted[1:] {
		if a.start-reach > gap {
			gap, after = a.start-reach, a.start
		}
		reach = math.Max(reach, a.start+a.width)
	}
	if wrap := sorted[0].start + 360 - reach; wrap > gap {
		gap, after = wrap, sorted[0].start
	}
	if gap <= 0 {
		return 0, 360 // everything is covered: the whole world
	}
	return after, 360 - gap
}

// FitWorld is upstream's view of the whole world (P-56): latitude 84 to -56,
// longitude 0, centred on the Mercator midpoint of the two, at the smaller of
// the zoom that fits that span's height and the zoom that fits the world's
// width.
func WholeWorld(cols, rows int) (View, error) {
	if cols <= 0 || rows <= 0 {
		return View{}, badView()
	}
	_, top, err := ToTile(LonLat{Lat: 84}, 0)
	if err != nil {
		return View{}, err
	}
	_, bottom, err := ToTile(LonLat{Lat: -56}, 0)
	if err != nil {
		return View{}, err
	}
	centre, err := FromTile(0.5, (top+bottom)/2, 0)
	if err != nil {
		return View{}, err
	}
	w, h := float64(cols*DotsPerCol), float64(rows*DotsPerRow)
	zoom := math.Min(math.Log2(h/((bottom-top)*TileSize)), math.Log2(w/TileSize))
	v := View{Centre: LonLat{Lon: 0, Lat: centre.Lat}, Zoom: math.Max(MinViewZoom, math.Min(MaxViewZoom, zoom)), Cols: cols, Rows: rows}
	return v, v.Validate()
}
