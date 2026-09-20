package main

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/png" // the pictures a provider sends
	"math"
)

// Entry is one row of a provider's colour table, as the scenario file
// carries it: a colour and what it stands for.
type Entry struct {
	R       uint8   `json:"r"`
	G       uint8   `json:"g"`
	B       uint8   `json:"b"`
	Value   float64 `json:"value"`
	Missing bool    `json:"missing"`
}

// Painted is a picture as the provider sends it: the image itself and the
// table that says what its colours mean. **Both this program and the
// library are given these two**, so the comparison is of two readings of
// one input and not of two different inputs.
type Painted struct {
	ID        string  `json:"id"`
	W         float64 `json:"west"`
	S         float64 `json:"south"`
	E         float64 `json:"east"`
	N         float64 `json:"north"`
	PNG       string  `json:"png_base64"`
	Table     []Entry `json:"table"`
	Tolerance float64 `json:"tolerance"`
}

// classify turns a painted picture into classes, by the rule the scenario
// states: each pixel takes the table entry nearest it, and a pixel further
// than the tolerance from every entry is no data. The rule is written out
// here from the description, not taken from the library.
func classify(p Painted) (Picture, bool) {
	body, err := base64.StdEncoding.DecodeString(p.PNG)
	if err != nil {
		return Picture{}, false
	}
	picture, _, err := image.Decode(bytes.NewReader(body))
	if err != nil {
		return Picture{}, false
	}
	tolerance := p.Tolerance
	if tolerance <= 0 {
		tolerance = 10 // the library's own default, written down rather than imported
	}
	box := picture.Bounds()
	out := Picture{ID: p.ID, W: p.W, S: p.S, E: p.E, N: p.N, Width: box.Dx(), Height: box.Dy()}
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			r, g, b, _ := picture.At(x, y).RGBA()
			out.Classes = append(out.Classes, classOfColour(uint8(r>>8), uint8(g>>8), uint8(b>>8), p.Table, tolerance))
		}
	}
	return out, true
}

// classOfColour is the class a colour stands for: the nearest entry of the
// table within the tolerance, counting the entries that are not "no data"
// in the order they are written; -1 for a colour that matches none and for
// one whose entry means no data.
func classOfColour(r, g, b uint8, table []Entry, tolerance float64) int {
	best, bestAt := math.Inf(1), -1
	class := 0
	classOf := make([]int, len(table))
	for i, e := range table {
		if e.Missing {
			classOf[i] = -1
			continue
		}
		classOf[i] = class
		class++
	}
	for i, e := range table {
		d := math.Sqrt(float64(diff(r, e.R)*diff(r, e.R) + diff(g, e.G)*diff(g, e.G) + diff(b, e.B)*diff(b, e.B)))
		if d < best {
			best, bestAt = d, i
		}
	}
	if bestAt < 0 || best > tolerance {
		return -1
	}
	return classOf[bestAt]
}

// diff is the difference between two channel values.
func diff(a, b uint8) int {
	if a > b {
		return int(a) - int(b)
	}
	return int(b) - int(a)
}
