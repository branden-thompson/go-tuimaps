package tuimaps_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestEveryAliasNamesItsFields: go doc does not show an alias's fields, so
// the doc comment of each alias of an internal struct names them by hand -
// and a field added to the struct and left out of the comment is a field no
// library user can find. Every exported field of the aliased struct is named
// in its alias's comment.
func TestEveryAliasNamesItsFields(t *testing.T) {
	docs := map[string]string{}
	fset := token.NewFileSet()
	for _, file := range []string{"describe.go", "overlays.go"} {
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				if ts.Assign.IsValid() && gen.Doc != nil {
					docs[ts.Name.Name] = gen.Doc.Text()
				}
			}
		}
	}
	for name, typ := range map[string]reflect.Type{
		"Answer": reflect.TypeOf(tuimaps.Answer{}),
		"Grid":   reflect.TypeOf(tuimaps.Grid{}),
		"Type":   reflect.TypeOf(tuimaps.Type{}),
		"Image":  reflect.TypeOf(tuimaps.Image{}),
	} {
		doc, ok := docs[name]
		if !ok {
			t.Errorf("no doc comment found for the alias %s", name)
			continue
		}
		for i := range typ.NumField() {
			if f := typ.Field(i); f.IsExported() && !strings.Contains(doc, f.Name) {
				t.Errorf("%s's comment does not name its field %s: go doc shows no alias's fields, so the comment is the only list", name, f.Name)
			}
		}
	}
}
