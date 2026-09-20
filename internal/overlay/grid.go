package overlay

import (
	"sort"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// NoData is the class of a value that is no number.
	NoData = -1
	// maxClasses is the most classes a host's own type may have: the most an
	// ordered, colour-vision-safe scale was found to carry (FR-7).
	maxClasses = 21
)

// Type is how a host says what a grid's values are: the name of a preset,
// which supplies breaks and colours and may be overridden, or breaks of the
// host's own (D-69).
type Type struct {
	Preset string // "temperature", "radar", or empty for the host's own type
	Unit   string
	Breaks []float64 // the host's own, or an override of the preset's
}

// Kind is a Type resolved: the breaks in effect, and the preset if there is one.
type Kind struct {
	Preset colour.Preset // zero for the host's own type
	Breaks []float64
}

// Classify is a value's class: the number of breaks at or below it, from 0
// to the number of breaks. A value that is no number is no data.
func Classify(value float64, breaks []float64) int {
	if !(value >= -1.797e308 && value <= 1.797e308) {
		return NoData
	}
	return sort.Search(len(breaks), func(i int) bool { return breaks[i] > value })
}

func badBreaks(why textsafe.Text) error {
	return fault.Make(fault.UnsortedBreaks, textsafe.Const("the overlay's type was refused"), why,
		textsafe.Const("give the class breaks as numbers in rising order, at most 20 of them, or name a preset"))
}

func unknownPreset(why textsafe.Text) error {
	return fault.Make(fault.UnknownPreset, textsafe.Const("the overlay's type was refused"), why,
		textsafe.Const("the presets are \"temperature\", in C or F, and \"radar\", in dBZ"))
}

// checkBreaks holds a host's breaks to: numbers, rising, no repeats, and no
// more classes than a safe scale carries.
func checkBreaks(breaks []float64) error {
	if len(breaks) == 0 {
		return badBreaks(textsafe.Const("a type of the host's own needs its class breaks, and none were given"))
	}
	if len(breaks)+1 > maxClasses {
		return badBreaks(textsafe.Const("it has more than 21 classes, the most a scale can carry and stay readable to everyone"))
	}
	for i, b := range breaks {
		if !(b >= -1.797e308 && b <= 1.797e308) {
			return badBreaks(textsafe.Const("one of its breaks is not a number"))
		}
		if i > 0 && b <= breaks[i-1] {
			return badBreaks(textsafe.Const("its breaks are not in rising order, or one repeats"))
		}
	}
	return nil
}

// presetBreaks are a preset's own breaks for a unit.
func presetBreaks(preset, unit string) (colour.Preset, []float64, error) {
	switch {
	case preset == "temperature" && unit == "C":
		return colour.Temperature, colour.TemperatureBreaks(colour.Celsius), nil
	case preset == "temperature" && unit == "F":
		return colour.Temperature, colour.TemperatureBreaks(colour.Fahrenheit), nil
	case preset == "radar" && unit == "dBZ":
		return colour.Radar, colour.RadarFloors(), nil
	case preset == "temperature" || preset == "radar":
		return 0, nil, unknownPreset(textsafe.Const("the preset is not defined in that unit"))
	}
	return 0, nil, unknownPreset(textsafe.Const("there is no preset of that name"))
}

// ResolveType turns what the host said into the breaks in effect.
func ResolveType(t Type) (Kind, error) {
	if t.Preset == "" {
		err := checkBreaks(t.Breaks)
		if err != nil {
			return Kind{}, err
		}
		return Kind{Breaks: append([]float64(nil), t.Breaks...)}, nil
	}
	preset, breaks, err := presetBreaks(t.Preset, t.Unit)
	if err != nil {
		return Kind{}, err
	}
	if len(t.Breaks) == 0 {
		return Kind{Preset: preset, Breaks: breaks}, nil
	}
	err = checkBreaks(t.Breaks)
	if err != nil {
		return Kind{}, err
	}
	return Kind{Preset: preset, Breaks: append([]float64(nil), t.Breaks...)}, nil
}
