package tuimaps_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// surfaceFile is where the public surface is written down, so that a change
// to it is a change to a committed file and never a surprise (NFR-22).
const surfaceFile = "06_docs/02_features/go-tuimaps/07-readiness/public-surface.txt"

// TestPublicSurfaceSnapshot is plan task 12.16: the exported names of the
// package, written down and compared. A name added, removed or renamed
// fails this test until the snapshot is updated in the same commit - which
// is the point: the surface cannot move quietly.
func TestPublicSurfaceSnapshot(t *testing.T) {
	got := strings.Join(surface(t), "\n") + "\n"
	want, err := os.ReadFile(filepath.Join("..", "..", surfaceFile))
	if os.IsNotExist(err) {
		want, err = os.ReadFile(surfaceFile)
	}
	if err != nil {
		t.Fatalf("the snapshot is missing: %v\nwrite it with the lines below:\n%s", err, got)
	}
	if got != string(want) {
		t.Errorf("the public surface has changed. If that was meant, put this in %s:\n%s", surfaceFile, got)
	}
}

// surface is every exported declaration of the package, one a line, sorted.
// The files are read one by one rather than as a directory, which reads the
// package's own files and nothing a build tag would have taken out.
func surface(t *testing.T) []string {
	t.Helper()
	names, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	set := token.NewFileSet()
	var out []string
	for _, name := range names {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(set, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		if file.Name.Name != "tuimaps" {
			continue
		}
		out = append(out, declared(file)...)
	}
	sort.Strings(out)
	return out
}

// declared are the exported declarations of one file, written as the host
// reads them: a function with its signature, a type with its kind, a
// constant or variable by name.
func declared(file *ast.File) []string {
	var out []string
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if !d.Name.IsExported() {
				continue
			}
			if d.Recv == nil {
				out = append(out, "func "+d.Name.Name)
				continue
			}
			out = append(out, "method "+receiver(d.Recv)+"."+d.Name.Name)
		case *ast.GenDecl:
			out = append(out, fromGroup(d)...)
		}
	}
	return out
}

// fromGroup are the exported names of one const, var or type group.
func fromGroup(d *ast.GenDecl) []string {
	var out []string
	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name.IsExported() {
				out = append(out, "type "+s.Name.Name)
			}
		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name.IsExported() {
					out = append(out, kindWord(d.Tok.String())+" "+name.Name)
				}
			}
		}
	}
	return out
}

func kindWord(tok string) string {
	if tok == "const" {
		return "const"
	}
	return "var"
}

// receiver is the type a method hangs off, without its pointer.
func receiver(list *ast.FieldList) string {
	if len(list.List) == 0 {
		return "?"
	}
	switch at := list.List[0].Type.(type) {
	case *ast.StarExpr:
		if name, ok := at.X.(*ast.Ident); ok {
			return name.Name
		}
	case *ast.Ident:
		return at.Name
	}
	return "?"
}
