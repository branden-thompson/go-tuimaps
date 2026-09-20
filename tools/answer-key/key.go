// Command answer-key works out, from a scenario's own data, where a place
// is relative to each of its shapes: inside or outside, how far the nearest
// edge is, and which way. It is **the independent half of M1b** (D-43,
// D-67): it shares no code with the library, so that the library's
// description is checked against something, not against itself.
//
// Everything here is written from the definitions rather than taken from
// the library: the winding number rather than the library's even-odd count,
// haversine written out, and the bearing formula as the textbooks give it.
// Where the two agree, they agree for a reason.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
)

// earthKm is the mean radius of the Earth, the figure the constants
// document names. It is the one number this program and the library share,
// and it is shared by being written down in both, not by being imported.
const earthKm = 6371.0088

// Place is a place in a scenario.
type Place struct {
	Name string  `json:"name"`
	Lon  float64 `json:"lon"`
	Lat  float64 `json:"lat"`
}

// Shape is one shape in a scenario: a ring for an area, a run for a line,
// one position for a point.
type Shape struct {
	ID    string        `json:"id"`
	Kind  string        `json:"kind"` // "area", "line" or "point"
	Rings [][][]float64 `json:"rings"`
}

// Scenario is what a key is worked out from.
type Scenario struct {
	Name     string    `json:"name"`
	Places   []Place   `json:"places"`
	Shapes   []Shape   `json:"shapes"`
	Fields   []Field   `json:"fields"`
	Pictures []Picture `json:"images"`
	// Views are the map sizes the scenario is drawn at, so that the key can
	// say how many cells from the edge a place is at each of them.
	Views []View `json:"views"`
}

// View is one map size, with how far a cell spans there.
type View struct {
	Name      string  `json:"name"`
	CellKm    float64 `json:"cell_km"`     // east to west
	CellRowKm float64 `json:"cell_row_km"` // north to south; a cell is taller than wide
}

// Answer is one place against one shape.
type Answer struct {
	Place   string             `json:"place"`
	Shape   string             `json:"shape"`
	Inside  *bool              `json:"inside,omitempty"`
	Km      float64            `json:"km"`
	Bearing float64            `json:"bearing_deg"`
	Compass string             `json:"compass"`
	Cells   map[string]float64 `json:"cells,omitempty"`
	Frame   map[string]string  `json:"frame_answer,omitempty"`
}

func main() {
	in := flag.String("scenario", "", "the scenario file to work a key out from")
	out := flag.String("out", "", "where to write the key; standard output if empty")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "answer-key: name a scenario file with -scenario")
		os.Exit(2)
	}
	body, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "answer-key:", err)
		os.Exit(1)
	}
	var scenario Scenario
	if err := json.Unmarshal(body, &scenario); err != nil {
		fmt.Fprintln(os.Stderr, "answer-key:", err)
		os.Exit(1)
	}
	key, err := json.MarshalIndent(WholeKey(scenario), "", " ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "answer-key:", err)
		os.Exit(1)
	}
	key = append(key, '\n')
	if *out == "" {
		os.Stdout.Write(key)
		return
	}
	if err := os.WriteFile(*out, key, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "answer-key:", err)
		os.Exit(1)
	}
}

// Keys is a whole key: the shapes, the fields and the images.
type Keys struct {
	Shapes   []Answer        `json:"shapes,omitempty"`
	Fields   []FieldAnswer   `json:"fields,omitempty"`
	Pictures []PictureAnswer `json:"images,omitempty"`
}

// WholeKey is every answer a scenario has: shapes, fields and images.
func WholeKey(s Scenario) Keys {
	out := Keys{Shapes: Key(s)}
	for _, p := range s.Places {
		if !onGlobe(p) {
			continue
		}
		for _, f := range s.Fields {
			out.Fields = append(out.Fields, atField(p, f))
		}
		for _, pic := range s.Pictures {
			out.Pictures = append(out.Pictures, atPicture(p, pic))
		}
	}
	return out
}

// Key is the whole key for a scenario: every place against every shape. A
// scenario with no places or no shapes has no key, and says so by being
// empty rather than by being half a key.
func Key(s Scenario) []Answer {
	if len(s.Places) == 0 || len(s.Shapes) == 0 {
		return nil
	}
	var out []Answer
	for _, place := range s.Places {
		if !onGlobe(place) {
			continue // a place that is nowhere is compared with nothing
		}
		for _, shape := range s.Shapes {
			if len(shape.Rings) == 0 {
				continue // a shape with no geometry is nothing to measure
			}
			out = append(out, answer(place, shape, s.Views))
		}
	}
	return out
}

// onGlobe reports whether a place in the file is a place on the world.
func onGlobe(p Place) bool {
	if p.Lat < -90 || p.Lat > 90 {
		return false
	}
	return p.Lon >= -180 && p.Lon <= 180
}

// answer is one place against one shape.
func answer(p Place, shape Shape, views []View) Answer {
	out := Answer{Place: p.Name, Shape: shape.ID}
	if len(shape.Rings) == 0 {
		return out
	}
	if shape.Kind == "area" {
		in := wound(p, shape.Rings)
		out.Inside = &in
	}
	km, bearing := nearest(p, shape)
	out.Km = round(km, 1)
	out.Bearing = math.Round(bearing)
	out.Compass = compass(bearing)
	if len(views) == 0 {
		return out
	}
	out.Cells, out.Frame = map[string]float64{}, map[string]string{}
	for _, v := range views {
		cells := cellsAcross(km, bearing, v)
		out.Cells[v.Name] = round(cells, 2)
		out.Frame[v.Name] = frameAnswer(cells, out.Inside)
	}
	return out
}

// wound reports whether a place is inside an area, **by the winding
// number**: the turning of the boundary as seen from the place, which is a
// different rule from the library's crossing count and must agree with it.
func wound(p Place, rings [][][]float64) bool {
	if len(rings) == 0 {
		return false
	}
	turn := 0.0
	for _, ring := range rings {
		for i := range ring {
			a, b := ring[i], ring[(i+1)%len(ring)]
			if len(a) < 2 || len(b) < 2 {
				continue
			}
			turn += angleAt(p, a, b)
		}
	}
	return math.Abs(turn) > math.Pi
}

// angleAt is the angle one edge subtends at the place, signed.
func angleAt(p Place, a, b []float64) float64 {
	if len(a) < 2 || len(b) < 2 {
		return 0
	}
	ax, ay := east(p.Lon, a[0]), a[1]-p.Lat
	bx, by := east(p.Lon, b[0]), b[1]-p.Lat
	angle := math.Atan2(bx, by) - math.Atan2(ax, ay)
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// nearest is how far the nearest part of a shape is, and which way, walking
// every edge and taking the nearest of a hundred steps along each. It is
// slower than the library's arithmetic and owes it nothing.
func nearest(p Place, shape Shape) (float64, float64) {
	if len(shape.Rings) == 0 {
		return 0, 0
	}
	bestKm, bestBearing := math.Inf(1), 0.0
	for _, ring := range shape.Rings {
		last := len(ring)
		if shape.Kind != "area" {
			last = len(ring) - 1 // a line is not closed
		}
		for i := 0; i < last; i++ {
			a, b := ring[i], ring[(i+1)%len(ring)]
			if len(a) < 2 || len(b) < 2 {
				continue
			}
			for step := 0; step <= 100; step++ {
				along := float64(step) / 100
				lon := a[0] + along*east(a[0], b[0])
				lat := a[1] + along*(b[1]-a[1])
				km := haversine(p.Lat, p.Lon, lat, lon)
				if km < bestKm {
					bestKm, bestBearing = km, bearingTo(p.Lat, p.Lon, lat, lon)
				}
			}
		}
		// Every vertex is a candidate whatever the shape is. A line of one
		// position is a point, and a shape with no walkable edge at all -
		// which the steps above pass over - still has somewhere to measure to.
		for _, at := range ring {
			if len(at) < 2 {
				continue
			}
			km := haversine(p.Lat, p.Lon, at[1], at[0])
			if km < bestKm {
				bestKm, bestBearing = km, bearingTo(p.Lat, p.Lon, at[1], at[0])
			}
		}
	}
	if math.IsInf(bestKm, 1) {
		return 0, 0 // nothing in the shape could be measured to at all
	}
	return bestKm, bestBearing
}

// haversine is the great-circle distance, written out.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	rad := math.Pi / 180
	dLat, dLon := (lat2-lat1)*rad, east(lon1, lon2)*rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earthKm * math.Asin(math.Min(1, math.Sqrt(a)))
}

// bearingTo is the initial bearing from one place to another, in degrees
// clockwise from north.
func bearingTo(lat1, lon1, lat2, lon2 float64) float64 {
	rad := math.Pi / 180
	dLon := east(lon1, lon2) * rad
	y := math.Sin(dLon) * math.Cos(lat2*rad)
	x := math.Cos(lat1*rad)*math.Sin(lat2*rad) - math.Sin(lat1*rad)*math.Cos(lat2*rad)*math.Cos(dLon)
	return math.Mod(math.Atan2(y, x)/rad+360, 360)
}

// east is how far east one longitude is from another, the short way.
func east(from, to float64) float64 {
	return math.Mod(to-from+540, 360) - 180
}

// compass is the word for a bearing, in full.
func compass(bearing float64) string {
	words := []string{"north", "north-east", "east", "south-east", "south", "south-west", "west", "north-west"}
	turn := math.Mod(math.Mod(bearing, 360)+360, 360)
	return words[int(math.Floor((turn+22.5)/45))%8]
}

// cellsAcross is how many cells of a view a distance is, **measured in the
// dimension the bearing points along** (S17-2): a column's width for
// something east or west, a row's height for something north or south. A
// cell is twice as tall as it is wide in ground terms, so saying "cells"
// without saying which way is saying half of it.
func cellsAcross(km, bearing float64, v View) float64 {
	turn := math.Mod(math.Mod(bearing, 360)+360, 360)
	northSouth := turn < 45 || turn >= 315 || (turn >= 135 && turn < 225)
	span := v.CellKm
	if northSouth && v.CellRowKm > 0 {
		span = v.CellRowKm
	}
	if span <= 0 {
		return 0
	}
	return km / span
}

// frameAnswer is what the drawn frame can say at that size: under one cell
// from the edge, the picture cannot settle it (D-67).
func frameAnswer(cells float64, inside *bool) string {
	if cells <= 0 {
		return "unknown" // a view that says nothing about its cells settles nothing
	}
	if cells < 1 {
		return "on the edge"
	}
	if inside != nil && *inside {
		return "inside"
	}
	return "outside"
}

// round is a number to a number of decimal places.
func round(v float64, places int) float64 {
	scale := math.Pow(10, float64(places))
	return math.Round(v*scale) / scale
}
