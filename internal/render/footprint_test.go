package render

import (
	"math/rand/v2"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// rasterClass is an image's class at one position, sampled on its own: the
// oracle under is held to. It is an image's class at a position, in the projection the host
// stated: rows in equal steps of latitude, or in equal steps of the web map's
// height.
func rasterClass(ra *scene.Raster, lon, lat float64) int8 {
	if ra == nil || ra.Width <= 0 || ra.Height <= 0 || len(ra.Classes) != ra.Width*ra.Height {
		return -1
	}
	if !(lon >= ra.West && lon < ra.East && lat > ra.South && lat <= ra.North) {
		return -1
	}
	col := int((lon - ra.West) / (ra.East - ra.West) * float64(ra.Width))
	down := (ra.North - lat) / (ra.North - ra.South)
	if ra.Projection == 2 {
		down = (mercator(ra.North) - mercator(lat)) / (mercator(ra.North) - mercator(ra.South))
	}
	row := int(down * float64(ra.Height))
	return ra.Classes[min(max(row, 0), ra.Height-1)*ra.Width+min(col, ra.Width-1)]
}

// TestAFootprintIsSampledAsEachPointAlone (REVIEW, perf F3): under places
// each of a footprint's rows and columns once, and answers what sampling
// each of its sixteen points alone answers, in both projections.
func TestAFootprintIsSampledAsEachPointAlone(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for _, projection := range []uint8{1, 2} {
		ra := &scene.Raster{West: -100, South: 20, East: -80, North: 45, Width: 97, Height: 61, Projection: projection}
		ra.Classes = make([]int8, ra.Width*ra.Height)
		for i := range ra.Classes {
			ra.Classes[i] = int8(rng.IntN(7)) - 1
		}
		r := &Renderer{}
		for x := range 60 {
			r.lons = append(r.lons, -105+0.5*float64(x))
		}
		for y := range 40 {
			r.lats = append(r.lats, 50-0.8*float64(y))
		}
		for y := range r.lats {
			for x := range r.lons {
				w, h := len(r.lons), len(r.lats)
				halfLon := (r.lons[min(x+1, w-1)] - r.lons[max(x-1, 0)]) / 4
				halfLat := (r.lats[max(y-1, 0)] - r.lats[min(y+1, h-1)]) / 4
				want := int8(-1)
				for i := range maxSpan {
					lat := r.lats[y] + halfLat - 2*halfLat*(float64(i)+0.5)/maxSpan
					for j := range maxSpan {
						lon := r.lons[x] - halfLon + 2*halfLon*(float64(j)+0.5)/maxSpan
						want = max(want, rasterClass(ra, lon, lat))
					}
				}
				if got := r.under(ra, x, y); got != want {
					t.Fatalf("projection %d, dot %d,%d: under %d, each point alone %d", projection, x, y, got, want)
				}
			}
		}
	}
}
