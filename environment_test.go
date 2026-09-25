package tuimaps_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// environmentReads are the calls that read the process's environment: the
// os and syscall readers, and net/http's proxy function, which reads the
// proxy variables.
var environmentReads = map[string]map[string]bool{
	"os":       {"Getenv": true, "LookupEnv": true, "Environ": true, "ExpandEnv": true},
	"syscall":  {"Getenv": true, "Environ": true},
	"net/http": {"ProxyFromEnvironment": true},
}

// TestEveryEnvironmentReadIsListed is v0.2.0 L10.1 (L-13.8): the library's
// environment reads are exactly these, and the contract names every variable
// they read, so a new one cannot arrive unnoticed. The command, the examples
// and the tools are modules of their own and are not the library.
func TestEveryEnvironmentReadIsListed(t *testing.T) {
	want := []string{
		"internal/fetch/fetch.go: http.ProxyFromEnvironment", // the library's own transport only
		"look.go: os.Getenv", // NO_COLOR, through colour.ChooseDepth
	}
	var got []string
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != "." && (entry.Name() == "testdata" || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(path, "go.mod")); path != "." && err == nil {
				return filepath.SkipDir // another module
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		names := map[string]string{} // the name a file gives each reader package
		for _, spec := range file.Imports {
			imported, _ := strconv.Unquote(spec.Path.Value)
			if environmentReads[imported] == nil {
				continue
			}
			name := imported[strings.LastIndex(imported, "/")+1:]
			if spec.Name != nil {
				name = spec.Name.Name
			}
			names[name] = imported
		}
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if imported, ok := names[pkg.Name]; ok && environmentReads[imported][sel.Sel.Name] {
				got = append(got, filepath.ToSlash(path)+": "+pkg.Name+"."+sel.Sel.Name)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("the library's environment reads are\n  %s\nthe list says\n  %s\na new read is named here and in the contract, or removed",
			strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
	contract := readContract(t)
	for _, variable := range []string{"NO_COLOR", "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
		if !strings.Contains(contract, "`"+variable+"`") {
			t.Errorf("the contract does not name %s, which the library reads", variable)
		}
	}
}

// TestNoDocCommentNamesAnotherDeclaration is v0.2.0 L10.2 (L-13.4): a doc
// comment that opens with the name of another declaration in its package has
// drifted onto the wrong one, as the disk cache's did when calls were renamed.
func TestNoDocCommentNamesAnotherDeclaration(t *testing.T) {
	packages := map[string][]*ast.File{}
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != "." && (entry.Name() == "testdata" || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(path, "go.mod")); path != "." && err == nil {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		packages[filepath.Dir(path)] = append(packages[filepath.Dir(path)], file)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for dir, files := range packages {
		declared := map[string]bool{}
		for _, file := range files {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					declared[d.Name.Name] = true
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						if ts, ok := spec.(*ast.TypeSpec); ok {
							declared[ts.Name.Name] = true
						}
					}
				}
			}
		}
		for _, file := range files {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Doc == nil {
					continue
				}
				checked++
				first, _, _ := strings.Cut(strings.TrimSpace(fn.Doc.Text()), " ")
				if first != fn.Name.Name && declared[first] {
					t.Errorf("%s: the doc comment of %s opens with %s, another declaration", dir, fn.Name.Name, first)
				}
			}
		}
	}
	if checked < 500 {
		t.Fatalf("checked %d doc comments; the library has far more", checked)
	}
}
