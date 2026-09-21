package overlay

import (
	"strconv"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// maxGridCells is the most cells a grid may have, as for an image's pixels.
const maxGridCells = 1_048_576

// Grid is a scalar grid as a host hands it in (FR-7): a regular longitude and
// latitude grid of values, rows from the north, with a type that says what
// the values are. The values are the host's memory, read when the grid is
// prepared and not kept.
type Grid struct {
	West, South, East, North float64
	Cols, Rows               int
	Values                   []float64
	Type                     Type
}

// checkGrid validates a grid on hand-in and resolves its type.
func checkGrid(g *Grid) (Kind, error) {
	if g == nil {
		return Kind{}, refused(fault.SizeMismatch, textsafe.Const("there is no grid in it"), textsafe.Const("hand in a grid, features or an image"))
	}
	if g.Cols <= 0 || g.Rows <= 0 {
		return Kind{}, refused(fault.SizeMismatch, textsafe.Const("its grid has no columns or no rows"),
			textsafe.Const("give the number of columns and of rows; the values are rows from the north, each west to east"))
	}
	if g.Cols > maxGridCells || g.Rows > maxGridCells || g.Cols*g.Rows > maxGridCells {
		return Kind{}, refused(fault.OverImageCap, textsafe.Const("its grid has more than 1,048,576 cells"),
			textsafe.Const("hand in a coarser grid: a terminal map shows a few thousand cells at most"))
	}
	if len(g.Values) != g.Cols*g.Rows {
		return Kind{}, refused(fault.SizeMismatch,
			textsafe.Join(textsafe.Const("its grid says "), textsafe.Clean(grouped(g.Cols*g.Rows)), textsafe.Const(" cells and carries "), textsafe.Clean(grouped(len(g.Values))), textsafe.Const(" values")),
			textsafe.Const("give exactly columns times rows values, rows from the north, each west to east"))
	}
	if !(g.West >= -180 && g.East <= 180 && g.South >= -90 && g.North <= 90) {
		return Kind{}, refused(fault.InvalidCoordinates, textsafe.Const("one of its grid's bounds is not a number, or is off the globe"),
			textsafe.Const("give west and east from -180 to 180, south and north from -90 to 90"))
	}
	if !(g.West < g.East && g.South < g.North) {
		return Kind{}, refused(fault.InvalidCoordinates, textsafe.Const("its grid's west is not west of its east, or its south is not south of its north"),
			textsafe.Const("check the order of the four bounds"))
	}
	return ResolveType(g.Type)
}

// implausible reports whether every value that is a number makes no sense in
// the unit declared: temperatures all above 60 declared as Celsius are very
// likely Fahrenheit. Such a grid is accepted, with a warning.
func implausible(g *Grid) bool {
	if g == nil || g.Type.Preset != "temperature" {
		return false
	}
	low, high := -90.0, 60.0
	if g.Type.Unit == "F" {
		low, high = -130, 140
	}
	numbers, outside := 0, 0
	for _, v := range g.Values {
		if Classify(v, nil) == NoData {
			continue
		}
		numbers++
		if v < low || v > high {
			outside++
		}
	}
	return numbers > 0 && outside == numbers
}

// classify makes a grid's prepared form: each value's class, once.
func classify(g *Grid, kind Kind) scene.Field {
	field := scene.Field{West: g.West, South: g.South, East: g.East, North: g.North, Cols: g.Cols, Rows: g.Rows,
		Classes: make([]int8, len(g.Values)), Preset: uint8(kind.Preset), ClassCount: len(kind.Breaks) + 1,
		Labels: bandLabels(kind.Breaks)}
	for i, v := range g.Values {
		field.Classes[i] = int8(Classify(v, kind.Breaks))
	}
	return field
}

// Field is an overlay's prepared grid, once a job has classified it.
func (s *Store) Field(id string) (scene.Field, bool) {
	if s == nil || id == "" {
		return scene.Field{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.fields[id]
	return f, ok
}

// bandLabels are the breaks as text, each placed at the class above it, so
// that a contour drawn between two bands carries the value it divides them at.
// **Without these a field drawn with no colour is bare lines**, which say where
// a band changes but not to what - the defect the first acceptance sitting
// found in scenario 4, and the reason D-35 asked for them.
func bandLabels(breaks []float64) []string {
	if len(breaks) == 0 {
		return nil
	}
	out := make([]string, len(breaks)+1) // class 0 lies below the first break and has no boundary of its own
	for i, b := range breaks {
		out[i+1] = strconv.FormatFloat(b, 'f', -1, 64)
	}
	return out
}
