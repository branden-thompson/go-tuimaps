package tuimaps_test

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
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
		out = append(out, declared(set, file)...)
	}
	sort.Strings(out)
	return out
}

// declared are the exported declarations of one file, written as the host
// reads them (v0.2.0 L1.1, L-4.3): a function or method with its signature, a
// type with its definition — a struct's exported fields, an interface's methods
// — and a constant or variable with its type. Parameter names are left out:
// they are not part of the contract.
func declared(set *token.FileSet, file *ast.File) []string {
	var out []string
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if !d.Name.IsExported() {
				continue
			}
			sig := signature(set, d.Type)
			if d.Recv == nil {
				out = append(out, "func "+d.Name.Name+sig)
				continue
			}
			out = append(out, "method "+receiver(d.Recv)+"."+d.Name.Name+sig)
		case *ast.GenDecl:
			out = append(out, fromGroup(set, d)...)
		}
	}
	return out
}

// fromGroup are the exported declarations of one const, var or type group.
func fromGroup(set *token.FileSet, d *ast.GenDecl) []string {
	var out []string
	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name.IsExported() {
				out = append(out, typeLine(set, s))
			}
		case *ast.ValueSpec:
			typ := ""
			if s.Type != nil {
				typ = " " + render(set, s.Type)
			}
			for _, name := range s.Names {
				if name.IsExported() {
					out = append(out, kindWord(d.Tok.String())+" "+name.Name+typ)
				}
			}
		}
	}
	return out
}

// typeLine is one exported type, with what a host can see of its definition.
func typeLine(set *token.FileSet, s *ast.TypeSpec) string {
	head := "type " + s.Name.Name + " "
	if s.Assign.IsValid() {
		head += "= "
	}
	switch t := s.Type.(type) {
	case *ast.StructType:
		var fields []string
		for _, f := range t.Fields.List {
			typ := render(set, f.Type)
			if len(f.Names) == 0 {
				fields = append(fields, typ) // embedded
				continue
			}
			for _, n := range f.Names {
				if n.IsExported() {
					fields = append(fields, n.Name+" "+typ)
				}
			}
		}
		return head + "struct{ " + strings.Join(fields, "; ") + " }"
	case *ast.InterfaceType:
		var methods []string
		for _, m := range t.Methods.List {
			if ft, ok := m.Type.(*ast.FuncType); ok && len(m.Names) > 0 {
				methods = append(methods, m.Names[0].Name+signature(set, ft))
				continue
			}
			methods = append(methods, render(set, m.Type))
		}
		return head + "interface{ " + strings.Join(methods, "; ") + " }"
	}
	return head + render(set, s.Type)
}

// signature is a function type's parameters and results, types only.
func signature(set *token.FileSet, ft *ast.FuncType) string {
	sig := "(" + typesOnly(set, ft.Params) + ")"
	if ft.Results == nil || len(ft.Results.List) == 0 {
		return sig
	}
	results := typesOnly(set, ft.Results)
	if len(ft.Results.List) == 1 && len(ft.Results.List[0].Names) <= 1 {
		return sig + " " + results
	}
	return sig + " (" + results + ")"
}

// typesOnly lists a field list's types, one for each name it declares.
func typesOnly(set *token.FileSet, fl *ast.FieldList) string {
	if fl == nil {
		return ""
	}
	var types []string
	for _, f := range fl.List {
		typ := render(set, f.Type)
		n := max(len(f.Names), 1)
		for range n {
			types = append(types, typ)
		}
	}
	return strings.Join(types, ", ")
}

// render prints one type expression on one line.
func render(set *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, set, node); err != nil {
		return "?"
	}
	return strings.Join(strings.Fields(buf.String()), " ")
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

// TestTheSurfaceRecordsSignatures is v0.2.0 L1.1 (L-4.3): a change to an
// exported signature or struct field must change the snapshot, not only a
// change of name.
func TestTheSurfaceRecordsSignatures(t *testing.T) {
	pairs := [][2]string{
		{"func Render(s Size) Frame", "func Render(s Size, now int) Frame"},
		{"type Frame struct{ Lines []string }", "type Frame struct{ Lines []string; Changed uint64 }"},
		{"func (m *Map) Purge() error", "func (m *Map) Purge() (int, error)"},
		{"const MaxFrames int = 36", "const MaxFrames uint = 36"},
	}
	for _, p := range pairs {
		a, b := surfaceOf(t, p[0]), surfaceOf(t, p[1])
		if a == b {
			t.Errorf("the surface cannot tell %q from %q: both read %q", p[0], p[1], a)
		}
	}
	if surfaceOf(t, "func F(a int) int") != surfaceOf(t, "func F(renamed int) int") {
		t.Error("a parameter's name is not part of the contract, and must not move the snapshot")
	}
}

// surfaceOf renders one source fragment's exported surface.
func surfaceOf(t *testing.T, src string) string {
	t.Helper()
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "x.go", "package tuimaps\n"+src, 0)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(declared(set, file), "\n")
}
