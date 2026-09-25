package overlay

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
