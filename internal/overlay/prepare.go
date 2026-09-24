package overlay

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// circleSides is how many sides a circle is drawn with.
	circleSides = 64
	// earthRadiusKm is the mean radius the library measures with (constants).
	earthRadiusKm = 6371.0088
)

// ringOf is a circle as a ring of positions, each the radius from the centre
// along the ground.
func ringOf(centre project.LonLat, radiusKm float64) []project.LonLat {
	if !(radiusKm > 0) {
		return nil
	}
	lat1, lon1 := centre.Lat*math.Pi/180, centre.Lon*math.Pi/180
	d := radiusKm / earthRadiusKm
	ring := make([]project.LonLat, 0, circleSides+1)
	for i := range circleSides {
		bearing := 2 * math.Pi * float64(i) / circleSides
		lat2 := math.Asin(math.Sin(lat1)*math.Cos(d) + math.Cos(lat1)*math.Sin(d)*math.Cos(bearing))
		lon2 := lon1 + math.Atan2(math.Sin(bearing)*math.Sin(d)*math.Cos(lat1), math.Cos(d)-math.Sin(lat1)*math.Sin(lat2))
		lon := math.Mod(lon2*180/math.Pi+540, 360) - 180
		ring = append(ring, project.LonLat{Lon: lon, Lat: math.Max(-project.MaxLatitude, math.Min(project.MaxLatitude, lat2*180/math.Pi))})
	}
	return append(ring, ring[0])
}

// prepareRing makes the drawn form of one ring: cut at the antimeridian,
// projected, and for a line or an area simplified for the bucket. A ring
// smaller than a dot gives nothing.
func prepareRing(ring []project.LonLat, kind scene.ShapeKind, tolerance float64) ([][]Vertex, error) {
	if len(ring) == 0 {
		return nil, nil
	}
	closed := ring
	if kind == scene.ShapeArea && ring[0] != ring[len(ring)-1] {
		closed = append(append(make([]project.LonLat, 0, len(ring)+1), ring...), ring[0])
	}
	parts := [][]project.LonLat{closed}
	if kind != scene.ShapePoint {
		split, err := Split(closed)
		if err != nil {
			return nil, err
		}
		parts = split
	}
	var out [][]Vertex
	for _, part := range parts {
		vertices, err := Project(part)
		if err != nil {
			return nil, err
		}
		if kind == scene.ShapePoint {
			out = append(out, vertices)
			continue
		}
		simplified := Simplify(vertices, tolerance)
		if kind == scene.ShapeArea {
			kept, ok := Keep(simplified)
			if !ok {
				continue
			}
			simplified = kept
		}
		out = append(out, simplified)
	}
	return out, nil
}

// Prepare makes an overlay's drawn form for one zoom bucket: the library's
// own copy of each shape, simplified to the bucket's tolerance and unclipped.
// It is the slow step, and runs inside a Work call, never while drawing.
func Prepare(o Overlay, bucket int) ([]scene.Shape, error) {
	if o.ID == "" || len(o.Features) == 0 || o.Grid != nil || o.Image != nil {
		return nil, fault.Make(fault.Internal, textsafe.Const("an overlay could not be prepared"),
			textsafe.Const("it was never accepted: it has no id or no features"), textsafe.Const("this is a defect in the library; report it"))
	}
	tolerance := Tolerance(bucket)
	shapes := make([]scene.Shape, 0, len(o.Features))
	for _, f := range o.Features {
		kind := map[FeatureKind]scene.ShapeKind{Point: scene.ShapePoint, Line: scene.ShapeLine, Polygon: scene.ShapeArea, Circle: scene.ShapeArea}[f.Kind]
		rings := f.Rings
		if f.Kind == Circle {
			rings = [][]project.LonLat{ringOf(f.Centre, f.RadiusKm)}
		}
		shape := scene.Shape{Kind: kind, Role: uint8(f.Role), Label: LabelOf(f)}
		if kind == scene.ShapeArea {
			shape.Mark = SeverityOf(f).Digit()
		}
		for _, ring := range rings {
			prepared, err := prepareRing(ring, kind, tolerance)
			if err != nil {
				return nil, err
			}
			shape.Rings = append(shape.Rings, prepared...)
		}
		shapes = append(shapes, shape)
	}
	return shapes, nil
}
