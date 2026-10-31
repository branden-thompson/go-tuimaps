// Package sun is where the sun stands over the Earth at a moment: the point
// it stands over, the line between day and night, the night side, and a
// place's day or night (go-tuiMaps v0.3.0 L-4.2, L-4.4).
package sun

import (
	"math"
	"time"
)

// Subsolar is the latitude and longitude, degrees, of the point the sun
// stands over at a moment: its declination, and its right ascension less
// Greenwich mean sidereal time; about 0.01° from a full ephemeris.
//
// Source: the Astronomical Almanac's low-precision formulas for the Sun,
// as the US Naval Observatory publishes them; GMST from the same.
func Subsolar(at time.Time) (lat, lon float64) {
	r := math.Pi / 180
	n := float64(at.UTC().UnixNano())/86400e9 + 2440587.5 - 2451545 // days from J2000.0
	l := 280.460 + 0.9856474*n                                      // mean longitude
	g := (357.528 + 0.9856003*n) * r                                // mean anomaly
	ecl := (l + 1.915*math.Sin(g) + 0.020*math.Sin(2*g)) * r        // ecliptic longitude
	obl := (23.439 - 0.0000004*n) * r
	ra := math.Atan2(math.Cos(obl)*math.Sin(ecl), math.Cos(ecl)) / r
	gmst := 280.46061837 + 360.98564736629*n
	return math.Asin(math.Sin(obl)*math.Sin(ecl)) / r, wrap180(ra - gmst)
}

// wrap180 is a longitude in [-180, 180).
func wrap180(lon float64) float64 { return math.Mod(math.Mod(lon+180, 360)+360, 360) - 180 }

// Elevation is the sun's height above the horizon at a place, degrees: the
// sun's centre, without refraction; no number for a place off the globe.
//
// Source: spherical trigonometry from the subsolar point.
func Elevation(lat, lon float64, at time.Time) float64 {
	if lat != lat || lon != lon {
		return math.NaN() // a place that is no number has no sun
	}
	if lat < -90 || lat > 90 {
		return math.NaN() // nor one past a pole
	}
	slat, slon := Subsolar(at)
	r := math.Pi / 180
	sin := math.Sin(lat*r)*math.Sin(slat*r) + math.Cos(lat*r)*math.Cos(slat*r)*math.Cos((lon-slon)*r)
	return math.Asin(math.Max(-1, math.Min(1, sin))) / r
}

// Line is the line between day and night at a moment: its latitude at
// every degree of longitude, from 180°W to 180°E.
//
// Source: the subsolar point (Subsolar); the great circle 90° from it.
func Line(at time.Time) [][2]float64 {
	slat, slon := Subsolar(at)
	if math.Abs(slat) < 1e-6 {
		slat = math.Copysign(1e-6, slat) // at an equinox the line is a meridian: kept off a division by zero
	}
	r := math.Pi / 180
	out := make([][2]float64, 0, 361)
	for lon := -180; lon <= 180; lon++ { // a degree a point (P10-02)
		lat := math.Atan(-math.Cos((float64(lon)-slon)*r)/math.Tan(slat*r)) / r
		out = append(out, [2]float64{float64(lon), lat})
	}
	return out
}

// Night is the night at a moment as one ring: the terminator, closed
// through the pole the sun does not reach.
//
// Source: as Line.
func Night(at time.Time) [][2]float64 {
	slat, _ := Subsolar(at)
	ring := Line(at)
	pole := -90.0 // the sun north of the equator: the south pole is in night
	if slat < 0 {
		pole = 90
	}
	ring = append(ring, [2]float64{180, pole}, [2]float64{-180, pole}, ring[0])
	return ring
}

// Turn is when the terminator next passes a place - the sun's next rising or
// setting there - to the second; the zero time where it does not within two
// days (a polar day or night) or the place is off the globe.
//
// Source: Elevation, stepped ten minutes at a time and halved to a second.
func Turn(lat, lon float64, at time.Time) time.Time {
	up := Elevation(lat, lon, at) > 0
	if lat != lat || lat < -90 || lat > 90 {
		return time.Time{}
	}
	const step = 10 * time.Minute
	for i := 1; i <= 2*24*6; i++ { // two days, ten minutes at a time (P10-02)
		t := at.Add(time.Duration(i) * step)
		if (Elevation(lat, lon, t) > 0) == up {
			continue
		}
		lo, hi := t.Add(-step), t
		for range 10 { // ten halvings take ten minutes under a second (P10-02)
			mid := lo.Add(hi.Sub(lo) / 2)
			if (Elevation(lat, lon, mid) > 0) == up {
				lo = mid
			} else {
				hi = mid
			}
		}
		return hi
	}
	return time.Time{}
}
