package tuimaps_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// contractFile is the library's one contract (v0.2.0 L1.2, L-4.1).
const contractFile = "06_docs/02_features/go-tuimaps/03-architecture-design/contract.md"

// apiSpan is a code span that names the API: an exported identifier,
// optionally one selector and a call. Error-kind strings, lowercase words and
// flags are prose, not API.
var apiSpan = regexp.MustCompile(`^([A-Z][A-Za-z0-9]*(?:\.[A-Z][A-Za-z0-9]*)?)(?:\(.*\))?$`)

// TestTheContractNamesOnlyWhatExists is v0.2.0 L1.2, one way: every name the
// contract puts in a code span exists in the package, directly or as a field
// or method reached through one of its types.
func TestTheContractNamesOnlyWhatExists(t *testing.T) {
	known := knownNames(t)
	var missing []string
	for _, span := range regexp.MustCompile("`([^`]*)`").FindAllStringSubmatch(readContract(t), -1) {
		m := apiSpan.FindStringSubmatch(strings.TrimSpace(span[1]))
		if m == nil {
			continue
		}
		if !known[m[1]] {
			missing = append(missing, m[1])
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("contract.md names what the package does not have: %s", strings.Join(dedupe(missing), ", "))
	}
}

// TestTheContractNamesEverythingExported is v0.2.0 L1.2, the other way: every
// exported name and method of the package appears in the contract (a method as
// Type.Method).
func TestTheContractNamesEverythingExported(t *testing.T) {
	text := readContract(t)
	var absent []string
	for _, name := range append(topLevelNames(t), methodNames(t)...) {
		if !regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`).MatchString(text) {
			absent = append(absent, name)
		}
	}
	if len(absent) > 0 {
		t.Errorf("exported but not in contract.md (%d): %s", len(absent), strings.Join(absent, ", "))
	}
}

func readContract(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(contractFile)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// topLevelNames are the package's exported top-level names, sorted.
func topLevelNames(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, line := range surface(t) {
		f := strings.Fields(line)
		if len(f) < 2 || f[0] == "method" {
			continue
		}
		out = append(out, strings.SplitN(f[1], "(", 2)[0])
	}
	sort.Strings(out)
	return dedupe(out)
}

// methodNames are the package's exported methods, as "Type.Method".
func methodNames(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, line := range surface(t) {
		if rest, ok := strings.CutPrefix(line, "method "); ok {
			out = append(out, strings.SplitN(rest, "(", 2)[0])
		}
	}
	sort.Strings(out)
	return out
}

// knownNames are the top-level names, "Type.Member" for every field and method
// of the package's own types, and the same for aliases, followed into the
// internal package that declares them.
func knownNames(t *testing.T) map[string]bool {
	t.Helper()
	known := map[string]bool{}
	for _, n := range topLevelNames(t) {
		known[n] = true
	}
	aliases := map[string][2]string{} // exported name → (internal package, type)
	members(t, ".", func(typ, member string) {
		known[typ+"."+member] = true
		known[member] = true // a method or field cited without its type
	})
	for _, name := range mustGlob(t, "*.go") {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			g, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range g.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Assign.IsValid() || !ts.Name.IsExported() {
					continue
				}
				if sel, ok := ts.Type.(*ast.SelectorExpr); ok {
					if pkg, ok := sel.X.(*ast.Ident); ok {
						aliases[ts.Name.Name] = [2]string{pkg.Name, sel.Sel.Name}
					}
				}
			}
		}
	}
	for exported, target := range aliases {
		dir := filepath.Join("internal", target[0])
		members(t, dir, func(typ, member string) {
			if typ == target[1] {
				known[exported+"."+member] = true
				known[member] = true
			}
		})
	}
	return known
}

// members calls found for every exported field and method in one directory.
func members(t *testing.T, dir string, found func(typ, member string)) {
	t.Helper()
	for _, name := range mustGlob(t, filepath.Join(dir, "*.go")) {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv != nil && d.Name.IsExported() {
					found(receiver(d.Recv), d.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if st, ok := ts.Type.(*ast.StructType); ok {
						for _, f := range st.Fields.List {
							for _, n := range f.Names {
								if n.IsExported() {
									found(ts.Name.Name, n.Name)
								}
							}
						}
					}
				}
			}
		}
	}
}

func mustGlob(t *testing.T, pattern string) []string {
	t.Helper()
	names, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	return names
}

func dedupe(sorted []string) []string {
	var out []string
	for i, s := range sorted {
		if i == 0 || s != sorted[i-1] {
			out = append(out, s)
		}
	}
	return out
}
