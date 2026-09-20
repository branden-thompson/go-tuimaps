package style

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/jsonsafe"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// maxStyleBytes and maxStyleDepth bound a user's style, which is
	// untrusted input (NFR-10, PL-IS-1).
	maxStyleBytes = 1 << 20
	maxStyleDepth = 64
	// maxHops bounds a chain of refs, and of constants naming constants.
	maxHops = 8
	// maxFilterSteps bounds the work of compiling one filter.
	maxFilterSteps = 1 << 16
)

func malformed(why textsafe.Text) error {
	return fault.Make(fault.MalformedStyle, textsafe.Const("the style was refused"), why,
		textsafe.Const("check the style; this release reads the legacy filter format, and a style that needs expressions cannot be used yet"))
}

func tooLarge() error {
	return fault.Make(fault.OverLimit, textsafe.Const("the style was refused"),
		textsafe.Const("it is larger than 1 MiB or nested deeper than 64"),
		textsafe.Const("check the style; the limits protect the host's memory"))
}

// Parse reads a user's style: the same JSON format as upstream's, with legacy
// filters. The library is handed the bytes and never reads a file (D-94).
func Parse(body []byte) (*Style, error) {
	if len(body) == 0 {
		return nil, malformed(textsafe.Const("it is empty"))
	}
	if !jsonsafe.Within(body, maxStyleBytes, maxStyleDepth) {
		return nil, tooLarge()
	}
	var doc struct {
		Constants map[string]any   `json:"constants"`
		Layers    []map[string]any `json:"layers"`
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber() // a filter tells 3 from 3.0, as upstream does
	err := dec.Decode(&doc)
	if err != nil {
		return nil, malformed(textsafe.Const("it is not JSON with a list of layers"))
	}
	p := &parser{constants: doc.Constants, byID: map[string]map[string]any{}}
	for _, layer := range doc.Layers {
		if id, ok := layer["id"].(string); ok {
			p.byID[id] = layer
		}
	}
	s := &Style{}
	for _, layer := range doc.Layers {
		rule, drawn, err := p.rule(layer)
		if err != nil {
			return nil, err
		}
		if drawn {
			s.rules = append(s.rules, rule)
		}
	}
	return s, nil
}

type parser struct {
	constants map[string]any
	byID      map[string]map[string]any
}

// constant follows an @name to its value. A name that leads back to itself
// is refused; a name the style does not define is left as it is written.
func (p *parser) constant(v any) (any, error) {
	for range maxHops {
		name, ok := v.(string)
		if !ok || len(name) == 0 || name[0] != '@' {
			return v, nil
		}
		next, defined := p.constants[name]
		if !defined {
			return v, nil
		}
		v = next
	}
	return nil, malformed(textsafe.Const("its constants name each other in a ring"))
}

// inherited follows a layer's ref to the layer whose type, source layer,
// zooms and filter it takes (P-43).
func (p *parser) inherited(layer map[string]any) (map[string]any, error) {
	from := layer
	for range maxHops {
		ref, ok := from["ref"].(string)
		if !ok {
			return from, nil
		}
		next, found := p.byID[ref]
		if !found {
			return nil, malformed(textsafe.Const("a layer's ref names a layer that is not there"))
		}
		from = next
	}
	return nil, malformed(textsafe.Const("its refs name each other in a ring"))
}

func number(v any) float64 {
	n, ok := v.(json.Number)
	if !ok {
		return 0
	}
	f, err := n.Float64()
	if err != nil || !(f >= 0 && f <= 30) {
		return 0
	}
	return f
}

// rule reads one layer. A layer of a type the map does not draw - a
// background, a raster - is passed over.
func (p *parser) rule(layer map[string]any) (Rule, bool, error) {
	if layer == nil {
		return Rule{}, false, nil
	}
	from, err := p.inherited(layer)
	if err != nil {
		return Rule{}, false, err
	}
	r := Rule{MinZoom: number(from["minzoom"]), MaxZoom: number(from["maxzoom"])}
	r.ID, _ = layer["id"].(string)
	r.Layer, _ = from["source-layer"].(string)
	switch from["type"] {
	case "line":
		r.Kind = Line
	case "fill":
		r.Kind = Fill
	case "symbol":
		r.Kind = Symbol
	default:
		return Rule{}, false, nil
	}
	if raw, has := from["filter"]; has {
		r.filter, err = p.filter(raw)
		if err != nil {
			return Rule{}, false, err
		}
	}
	paint, _ := layer["paint"].(map[string]any)
	r.colours, err = p.colours(paint)
	if err != nil {
		return Rule{}, false, err
	}
	r.widths, err = p.widths(paint)
	return r, err == nil, err
}

// isExpression reports whether an operator belongs to the newer expression
// language, which this release does not read (FR-20).
func isExpression(op string) bool {
	switch op {
	case "get", "match", "case", "coalesce", "step", "interpolate", "let", "var", "literal",
		"to-number", "to-string", "to-boolean", "concat", "zoom", "geometry-type", "at", "length", "!":
		return true
	}
	return false
}

// frame is a group being compiled: its operator, its parts, and how many of
// them have been taken up.
type frame struct {
	op    string
	parts []any
	next  int
}

// header reads a filter's operator, and refuses what is not a legacy filter.
func header(raw any) (string, []any, error) {
	list, ok := raw.([]any)
	if !ok || len(list) == 0 {
		return "", nil, malformed(textsafe.Const("a filter is not a list that starts with an operator"))
	}
	op, ok := list[0].(string)
	if !ok {
		return "", nil, malformed(textsafe.Const("a filter is not a list that starts with an operator"))
	}
	if isExpression(op) {
		return "", nil, malformed(textsafe.Const("a filter is written as an expression"))
	}
	return op, list, nil
}

// filter compiles a legacy filter into steps. Groups nest; they are followed
// with a list of frames, not by a function calling itself, and the number of
// steps is bounded.
func (p *parser) filter(raw any) ([]step, error) {
	var program []step
	var open []frame
	current, have := raw, true
	for range maxFilterSteps {
		if have {
			op, list, err := header(current)
			if err != nil {
				return nil, err
			}
			have = false
			if op == "all" || op == "any" || op == "none" {
				open = append(open, frame{op: op, parts: list[1:]})
				continue
			}
			leaf, err := p.leaf(op, list)
			if err != nil {
				return nil, err
			}
			program = append(program, step{leaf: leaf})
			continue
		}
		if len(open) == 0 {
			return program, nil
		}
		top := &open[len(open)-1]
		if top.next < len(top.parts) {
			current, have = top.parts[top.next], true
			top.next++
			continue
		}
		program = append(program, step{group: top.op, count: len(top.parts)})
		open = open[:len(open)-1]
	}
	return nil, malformed(textsafe.Const("a filter has more parts than any real style"))
}

// leaf reads a filter that asks about one key.
func (p *parser) leaf(op string, list []any) (*node, error) {
	n := &node{op: op}
	if len(list) < 2 {
		return n, nil
	}
	if _, nested := list[1].([]any); nested {
		return nil, malformed(textsafe.Const("a filter is written as an expression"))
	}
	n.key, _ = list[1].(string)
	for _, operand := range list[2:] {
		resolved, err := p.constant(operand)
		if err != nil {
			return nil, err
		}
		if v, ok := scalar(resolved); ok {
			n.values = append(n.values, v)
		}
	}
	return n, nil
}

// stops reads a paint value as zoom stops: a plain value is one stop.
func (p *parser) stops(raw any) ([][2]any, error) {
	resolved, err := p.constant(raw)
	if err != nil {
		return nil, err
	}
	if _, isList := resolved.([]any); isList {
		return nil, malformed(textsafe.Const("a paint value is written as an expression"))
	}
	object, ok := resolved.(map[string]any)
	if !ok {
		return [][2]any{{json.Number("0"), resolved}}, nil
	}
	list, _ := object["stops"].([]any)
	var out [][2]any
	for _, entry := range list {
		pair, ok := entry.([]any)
		if !ok || len(pair) != 2 {
			return nil, malformed(textsafe.Const("a zoom stop is not a pair of a zoom and a value"))
		}
		v, err := p.constant(pair[1])
		if err != nil {
			return nil, err
		}
		out = append(out, [2]any{pair[0], v})
	}
	return out, nil
}

// red is upstream's colour for a layer with none, or one it cannot read (P-44).
func red() colour.RGB { return colour.RGB{R: 255} }

// hex reads #rgb or #rrggbb, the two forms upstream reads.
func hex(v any) colour.RGB {
	s, ok := v.(string)
	if !ok || (len(s) != 4 && len(s) != 7) || s[0] != '#' {
		return red()
	}
	digits := s[1:]
	if len(digits) == 3 {
		digits = string([]byte{digits[0], digits[0], digits[1], digits[1], digits[2], digits[2]})
	}
	n, err := strconv.ParseUint(digits, 16, 32)
	if err != nil {
		return red()
	}
	return colour.RGB{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n)}
}

// colours picks the layer's colour as upstream does - the line colour, else
// the fill colour, else the text colour, else red - with every stop kept.
func (p *parser) colours(paint map[string]any) ([]colourStop, error) {
	for _, key := range []string{"line-color", "fill-color", "text-color"} {
		raw, has := paint[key]
		if !has {
			continue
		}
		stops, err := p.stops(raw)
		if err != nil {
			return nil, err
		}
		var out []colourStop
		for _, s := range stops {
			out = append(out, colourStop{zoom: number(s[0]), colour: hex(s[1])})
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	return []colourStop{{colour: red()}}, nil
}

// widths reads the layer's line width: a number, or its stops.
func (p *parser) widths(paint map[string]any) ([]widthStop, error) {
	raw, has := paint["line-width"]
	if !has {
		return nil, nil
	}
	stops, err := p.stops(raw)
	if err != nil {
		return nil, err
	}
	var out []widthStop
	for _, s := range stops {
		if w := number(s[1]); w > 0 {
			out = append(out, widthStop{zoom: number(s[0]), width: w})
		}
	}
	return out, nil
}
