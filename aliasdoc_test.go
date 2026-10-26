package tuimaps_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"strings"
	"testing"
)

// TestEveryAliasNamesItsFields: go doc does not show an alias's fields, so
// the doc comment of each alias of an internal struct names them by hand -
// and a field added to the struct and left out of the comment is a field no
// library user can find. The package is type-checked, so every exported
// alias of a struct is found, not a list kept by hand; every exported field
// of the aliased struct is named in its alias's comment.
func TestEveryAliasNamesItsFields(t *testing.T) {
	fset := token.NewFileSet()
	parsed, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	docs := map[string]string{}
	for _, f := range parsed["tuimaps"].Files {
		files = append(files, f)
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				docs[ts.Name.Name] = gen.Doc.Text() + ts.Doc.Text()
			}
		}
	}
	conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil)}
	pkg, err := conf.Check("github.com/branden-thompson/go-tuimaps", fset, files, nil)
	if err != nil {
		t.Fatal(err)
	}
	aliases := 0
	for _, name := range pkg.Scope().Names() {
		tn, ok := pkg.Scope().Lookup(name).(*types.TypeName)
		if !ok || !tn.Exported() || !tn.IsAlias() {
			continue
		}
		st, ok := types.Unalias(tn.Type()).Underlying().(*types.Struct)
		if !ok {
			continue
		}
		aliases++
		for i := range st.NumFields() {
			if f := st.Field(i); f.Exported() && !strings.Contains(docs[name], f.Name()) {
				t.Errorf("%s's comment does not name its field %s: go doc shows no alias's fields, so the comment is the only list", name, f.Name())
			}
		}
	}
	if aliases < 4 {
		t.Errorf("%d aliases of a struct found; the package has more, so the search is broken", aliases)
	}
}
