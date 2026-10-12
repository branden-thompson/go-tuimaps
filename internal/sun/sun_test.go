package sun

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// noaaElevation is the sun's elevation at a place by NOAA's solar-position
// spreadsheet (after Meeus), the oracle these helpers are held to (watchpost
// A-13): written here from its published equations, apart from the helpers'
// own.
func noaaElevation(lat, lon float64, at time.Time) float64 {
	at = at.UTC()
	r := math.Pi / 180
	jd := float64(at.Unix())/86400 + 2440587.5
	t := (jd - 2451545) / 36525
	l0 := math.Mod(280.46646+t*(36000.76983+t*0.0003032), 360)
	m := 357.52911 + t*(35999.05029-0.0001537*t)
	e := 0.016708634 - t*(0.000042037+0.0000001267*t)
	c := math.Sin(m*r)*(1.914602-t*(0.004817+0.000014*t)) + math.Sin(2*m*r)*(0.019993-0.000101*t) + math.Sin(3*m*r)*0.000289
	app := l0 + c - 0.00569 - 0.00478*math.Sin((125.04-1934.136*t)*r)
	obl := 23 + (26+(21.448-t*(46.815+t*(0.00059-t*0.001813)))/60)/60 + 0.00256*math.Cos((125.04-1934.136*t)*r)
	decl := math.Asin(math.Sin(obl*r) * math.Sin(app*r))
	y := math.Pow(math.Tan(obl*r/2), 2)
	eot := 4 / r * (y*math.Sin(2*l0*r) - 2*e*math.Sin(m*r) + 4*e*y*math.Sin(m*r)*math.Cos(2*l0*r) - 0.5*y*y*math.Sin(4*l0*r) - 1.25*e*e*math.Sin(2*m*r))
	minutes := float64(at.Hour()*60+at.Minute()) + float64(at.Second())/60
	ha := (minutes+eot+4*lon)/4 - 180 // the hour angle, degrees
	cosZ := math.Sin(lat*r)*math.Sin(decl) + math.Cos(lat*r)*math.Cos(decl)*math.Cos(ha*r)
	return 90 - math.Acos(math.Max(-1, math.Min(1, cosZ)))/r
}

// yearOfTimes are moments across a year, every five days at a changing hour,
// the equinoxes and solstices among them.
func yearOfTimes() []time.Time {
	var out []time.Time
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for d := 0; d < 366; d += 5 {
		out = append(out, start.AddDate(0, 0, d).Add(time.Duration(d%24)*time.Hour+17*time.Minute))
	}
	return append(out, time.Date(2026, 3, 20, 14, 46, 0, 0, time.UTC), time.Date(2026, 6, 21, 8, 24, 0, 0, time.UTC),
		time.Date(2026, 9, 23, 0, 5, 0, 0, time.UTC), time.Date(2026, 12, 21, 20, 50, 0, 0, time.UTC))
}

// TestTheTerminatorMatchesTheEphemeris is L-4.2 (watchpost A-13): at every
// longitude, across a year, the terminator lies within 0.5° of NOAA's - the
// sun on NOAA's reckoning within half a degree of the horizon at each of its
// points, a distance that holds where the line runs north-south, as near an
// equinox, where its latitude alone could not.
func TestTheTerminatorMatchesTheEphemeris(t *testing.T) {
	for _, at := range yearOfTimes() {
		line := Line(at)
		if len(line) != 361 || line[0][0] != -180 || line[360][0] != 180 {
			t.Fatalf("%v: the line has %d points, from %v to %v", at, len(line), line[0], line[len(line)-1])
		}
		for _, p := range line {
			if e := noaaElevation(p[1], p[0], at); math.Abs(e) > 0.5 || math.IsNaN(p[1]) {
				t.Fatalf("%v: the terminator at %v°E, %v°N has the sun %.2f° from the horizon", at, p[0], p[1], e)
			}
		}
	}
}

// inRing reports whether a point lies inside a ring of longitudes and
// latitudes, by the crossings of a ray to the east.
func inRing(lon, lat float64, ring [][2]float64) bool {
	in := false
	for i, j := 0, len(ring)-1; i < len(ring); j, i = i, i+1 {
		a, b := ring[i], ring[j]
		if (a[1] > lat) != (b[1] > lat) && lon < (b[0]-a[0])*(lat-a[1])/(b[1]-a[1])+a[0] {
			in = !in
		}
	}
	return in
}

// TestNightSideIsTheComplement is L-4.2: a place inside the night side has
// the sun below the horizon on NOAA's reckoning, and one outside has it
// above - away from the line itself, within the half degree it is held to.
func TestNightSideIsTheComplement(t *testing.T) {
	for _, at := range yearOfTimes() {
		ring := Night(at)
		for lon := -175.0; lon < 180; lon += 10 {
			for lat := -85.0; lat < 90; lat += 10 {
				e := noaaElevation(lat, lon, at)
				if math.Abs(e) < 0.5 {
					continue
				}
				if night := inRing(lon, lat, ring); night != (e < 0) {
					t.Fatalf("%v: %v°E %v°N is night %v with the sun at %.2f°", at, lon, lat, night, e)
				}
			}
		}
	}
}

// TestADaysPlaceIsInDayOrNight is L-4.4's data: a place's elevation agrees
// with NOAA's to half a degree, so day and night are said as the map draws
// them.
func TestADaysPlaceIsInDayOrNight(t *testing.T) {
	for _, at := range yearOfTimes() {
		for _, p := range [][2]float64{{37.7, -97.3}, {-33.9, 151.2}, {64.8, -147.7}, {0, 0}} {
			if got, want := Elevation(p[0], p[1], at), noaaElevation(p[0], p[1], at); math.Abs(got-want) > 0.5 {
				t.Fatalf("%v at %v: elevation %.2f°, NOAA's %.2f°", p, at, got, want)
			}
		}
	}
}

// TestAPlaceOffTheGlobeHasNoElevation: a latitude past a pole, or a place
// that is no number, has no sun to answer for.
func TestAPlaceOffTheGlobeHasNoElevation(t *testing.T) {
	at := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	for _, p := range [][2]float64{{91, 0}, {-90.5, 10}, {math.NaN(), 0}, {0, math.NaN()}} {
		if e := Elevation(p[0], p[1], at); !math.IsNaN(e) {
			t.Errorf("%v has the sun at %v°; want no answer", p, e)
		}
	}
}
