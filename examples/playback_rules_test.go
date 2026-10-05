package examples_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestThePlaybackHostKeepsNoState is v0.2.0 L4.11's rules test (L-1.13): the
// example host wires every control and Settings row with no playback state of
// its own. Its type holds the map and nothing else.
func TestThePlaybackHostKeepsNoState(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "example_playback_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok || ts.Name.Name != "playbackHost" {
			return true
		}
		found = true
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			t.Fatalf("playbackHost is not a struct")
		}
		var fields []string
		for _, f := range st.Fields.List {
			for _, name := range f.Names {
				fields = append(fields, name.Name)
			}
			if len(f.Names) == 0 {
				fields = append(fields, "(embedded)")
			}
		}
		if len(fields) != 1 || fields[0] != "m" {
			t.Errorf("playbackHost has fields %v; a host needs only the map", fields)
		}
		if star, ok := st.Fields.List[0].Type.(*ast.StarExpr); !ok || !isMap(star.X) {
			t.Error("playbackHost's one field is not the map")
		}
		return false
	})
	if !found {
		t.Fatal("the example has no playbackHost")
	}
}

func isMap(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	return ok && sel.Sel.Name == "Map"
}
