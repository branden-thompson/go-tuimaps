package overlay

import (
	"math"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
)

// Provider names a source of radar images whose colour table the library
// carries (L-2.5, D-35): an image that names one needs no table of its own.
// Each provider's table is registered by a file of its own, provider_*.go,
// so one is added without touching another (D-70).
type Provider uint8

// The providers the library carries tables for.
const (
	ProviderIEM  Provider = iota + 1 // the Iowa Environmental Mesonet's N0Q composite
	ProviderMRMS                     // NOAA's Multi-Radar Multi-Sensor reflectivity
)

// ProviderTable is one provider's colour table and what is known of it.
// Approximate says the values were read from a legend rather than published
// (D-19); Unverified says part of its range has not yet been seen in data
// (L-2.3, RK-2).
type ProviderTable struct {
	Name                    string
	Entries                 []TableEntry
	Approximate, Unverified bool
	// Gradient is the provider's legend where the observed palette runs out,
	// as stops in rising value: a colour the table does not hold is valued
	// by where it projects onto it (L-2.3, L6.5). Nil for a published table.
	Gradient []TableEntry
}

// fallbackReach is how far a colour may lie from a provider's legend
// gradient and still be valued along it: wave 2 found MRMS's heavy-end
// colours 3 to 30 from any single legend colour (M-A). Beyond it, a colour
// is taken for something that is not rain, and matches nothing.
const fallbackReach = 30.0

// gradient is a provider's legend gradient in Lab, worked out once for an
// image's reading.
type gradient struct {
	at     []colour.Lab
	values []float64
}

func gradientOf(stops []TableEntry) gradient {
	g := gradient{at: make([]colour.Lab, len(stops)), values: make([]float64, len(stops))}
	for i, s := range stops {
		g.at[i], g.values[i] = colour.InLab(s.Colour, colour.Normal), s.Value
	}
	return g
}

// value is where a colour projects onto the gradient: the value there, and
// how far the colour is from it.
func (g gradient) value(c colour.RGB) (float64, float64) {
	p := colour.InLab(c, colour.Normal)
	best, value := math.Inf(1), 0.0
	for i := 0; i+1 < len(g.at); i++ {
		a, b := g.at[i], g.at[i+1]
		ab := [3]float64{b.L - a.L, b.A - a.A, b.B - a.B}
		ap := [3]float64{p.L - a.L, p.A - a.A, p.B - a.B}
		t := 0.0
		if den := ab[0]*ab[0] + ab[1]*ab[1] + ab[2]*ab[2]; den > 0 {
			t = math.Max(0, math.Min(1, (ap[0]*ab[0]+ap[1]*ab[1]+ap[2]*ab[2])/den))
		}
		q := colour.Lab{L: a.L + t*ab[0], A: a.A + t*ab[1], B: a.B + t*ab[2]}
		if d := p.Distance(q); d < best {
			best, value = d, g.values[i]+t*(g.values[i+1]-g.values[i])
		}
	}
	return value, best
}

// tables are the registered providers' tables.
var tables = map[Provider]ProviderTable{}

// register adds one provider's table. It is called once, by that provider's
// own file.
func register(p Provider, t ProviderTable) {
	if _, twice := tables[p]; twice {
		panic("overlay: a provider registered twice")
	}
	tables[p] = t
}

// TableOf is a provider's table, if the library carries one.
func TableOf(p Provider) (ProviderTable, bool) {
	t, ok := tables[p]
	return t, ok
}
