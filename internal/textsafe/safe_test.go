package textsafe

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestSafeOnlyBuiltHere is plan task 02.11, first half: there is one
// enforced way out. Text cannot be made from a plain string anywhere but
// here, and here only by the functions that clean or validate.
func TestSafeOnlyBuiltHere(t *testing.T) {
	typ := reflect.TypeOf(Text{})
	for i := range typ.NumField() {
		if typ.Field(i).IsExported() {
			t.Errorf("Text has an exported field %s; another package could fill it with anything", typ.Field(i).Name)
		}
	}
	if typ.NumField() == 0 {
		t.Error("Text has no field: a plain conversion would build one")
	}

	// Const takes a type no other package can name, so a caller elsewhere can
	// pass an untyped string constant and nothing else.
	param := reflect.TypeOf(Const).In(0)
	if param.Kind() != reflect.String || param.PkgPath() == "" || ast.IsExported(param.Name()) {
		t.Errorf("Const takes %v; it must take an unexported string type of this package", param)
	}
	if got := Const("drawn from embedded tiles").String(); got != "drawn from embedded tiles" {
		t.Errorf("Const changed its constant: %q", got)
	}

	// Every exported function that turns a plain string into Text must be one
	// that cleans or validates.
	allowed := map[string]bool{"Clean": true, "Quote": true, "ID": true}
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range sources {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !fn.Name.IsExported() || fn.Recv != nil || fn.Type.Results == nil {
				continue
			}
			if takesPlainString(fn) && returnsText(fn) && !allowed[fn.Name.Name] {
				t.Errorf("%s turns a plain string into Text without being on the list of functions that clean or validate", fn.Name.Name)
			}
		}
	}
}

func takesPlainString(fn *ast.FuncDecl) bool {
	for _, p := range fn.Type.Params.List {
		if id, ok := p.Type.(*ast.Ident); ok && id.Name == "string" {
			return true
		}
	}
	return false
}

func returnsText(fn *ast.FuncDecl) bool {
	for _, r := range fn.Type.Results.List {
		if id, ok := r.Type.(*ast.Ident); ok && id.Name == "Text" {
			return true
		}
	}
	return false
}
