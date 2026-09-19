package style

import (
	"encoding/json"
	"strings"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// value is a JSON scalar as a filter compares it. Upstream compares JSON
// values, so an integer and a float are distinct, and a string is never a
// number (P-42).
type value struct {
	text    string
	number  float64
	isText  bool
	isWhole bool // a number written without a fraction or an exponent
}

// node is one legacy filter that asks about a key: an operator, the key, and
// the values it compares with.
type node struct {
	op     string
	key    string
	values []value
}

// step is one step of a compiled filter, in the order the filter is worked
// out: a leaf pushes its answer; a group - all, any, none - takes the
// answers of its count of parts and pushes one. A filter nests, and nothing
// here calls itself to follow it.
type step struct {
	leaf  *node
	group string
	count int
}

func scalar(v any) (value, bool) {
	switch x := v.(type) {
	case string:
		return value{text: x, isText: true}, true
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			return value{}, false
		}
		return value{number: f, isWhole: !strings.ContainsAny(x.String(), ".eE")}, true
	case bool:
		if x {
			return value{number: 1, isWhole: true}, true
		}
		return value{isWhole: true}, true
	}
	return value{}, false
}

// attr is a feature's value for a key, and whether it has one.
func attr(a Attrs, key string) (value, bool) {
	whole := func(n int64, present bool) (value, bool) {
		return value{number: float64(n), isWhole: true}, present
	}
	switch key {
	case "$type":
		return value{text: typeName(a.Kind), isText: true}, true
	case "class":
		return value{text: a.Class, isText: true}, a.Class != ""
	case "name":
		return value{text: a.Name, isText: true}, a.Name != ""
	case "rank":
		return whole(int64(a.Rank), a.Rank != 0)
	case "admin_level":
		return whole(int64(a.AdminLevel), a.AdminLevel != 0)
	case "maritime":
		if a.Maritime {
			return whole(1, true)
		}
		return whole(0, a.AdminLevel != 0)
	}
	return value{}, false
}

func typeName(k scene.GeomKind) string {
	switch k {
	case scene.GeomPoint:
		return "Point"
	case scene.GeomLine:
		return "LineString"
	case scene.GeomPolygon:
		return "Polygon"
	}
	return ""
}

func (v value) equals(o value) bool {
	if v.isText != o.isText {
		return false
	}
	if v.isText {
		return v.text == o.text
	}
	return v.isWhole == o.isWhole && v.number == o.number
}

// run works a compiled filter out for a feature. An empty filter takes
// everything.
func run(program []step, a Attrs) bool {
	if len(program) == 0 {
		return true
	}
	var room [32]bool
	answers := room[:0]
	for _, s := range program {
		if s.leaf != nil {
			answers = append(answers, s.leaf.eval(a))
			continue
		}
		if s.count > len(answers) {
			return false // a program this package did not compile
		}
		parts := answers[len(answers)-s.count:]
		answers = append(answers[:len(answers)-s.count], group(s.group, parts))
	}
	return len(answers) == 1 && answers[0]
}

// group is all, any or none of its parts.
func group(op string, parts []bool) bool {
	if op == "" {
		return false
	}
	trues := 0
	for _, p := range parts {
		if p {
			trues++
		}
	}
	switch op {
	case "all":
		return trues == len(parts)
	case "any":
		return trues > 0
	}
	return trues == 0
}

// eval is upstream's evaluation of one filter (P-42): an operator it does
// not know is true, and == on a key the feature lacks is false.
func (n *node) eval(a Attrs) bool {
	if n == nil {
		return true
	}
	switch n.op {
	case "==", "!=", "in", "!in", "has", "!has":
		return n.evalEquality(a)
	case ">", ">=", "<", "<=":
		return n.evalOrder(a)
	}
	return true
}

func (n *node) evalEquality(a Attrs) bool {
	got, present := attr(a, n.key)
	if n.op == "has" || n.op == "!has" {
		return present == (n.op == "has")
	}
	found := false
	for _, v := range n.values {
		if present && got.equals(v) {
			found = true
		}
	}
	return found == (n.op == "==" || n.op == "in")
}

func (n *node) evalOrder(a Attrs) bool {
	got, present := attr(a, n.key)
	if !present || got.isText || len(n.values) == 0 || n.values[0].isText {
		return false
	}
	want := n.values[0].number
	switch n.op {
	case ">":
		return got.number > want
	case ">=":
		return got.number >= want
	case "<":
		return got.number < want
	}
	return got.number <= want
}
