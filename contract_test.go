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

// readContract is the contract without its changelog, whose names are
// removed or not yet built.
func readContract(t *testing.T) string {
	t.Helper()
	before, _, _ := strings.Cut(readWholeContract(t), changelogHeading)
	return before
}

func readWholeContract(t *testing.T) string {
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

// changelogHeading opens the contract's list of v0.2.0's breaks (L1.5). The
// names it cites are removed or not yet built, so the two name checks above
// read the contract without it.
const changelogHeading = "## 11 · Changelog: what v0.2.0 breaks (D-58)"

// TestTheChangelogListsEveryBreak is v0.2.0 L1.5 (D-58): the contract has a
// changelog with one row per break the plan names, each saying what changes,
// why, what a host does instead, and its state; and a row marked landed is
// true: the names it says were removed are gone from the package.
func TestTheChangelogListsEveryBreak(t *testing.T) {
	_, log, ok := strings.Cut(readWholeContract(t), changelogHeading)
	if !ok {
		t.Fatalf("contract.md has no %q section", changelogHeading)
	}
	rows := map[string][]string{}
	for _, line := range strings.Split(log, "\n") {
		if strings.HasPrefix(line, "## ") {
			break
		}
		if !strings.HasPrefix(line, "| ") || strings.HasPrefix(line, "| Break ") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "| "), " | ")
		if len(cells) != 5 {
			t.Errorf("a changelog row has %d cells, want 5: %s", len(cells), line)
			continue
		}
		for i, c := range cells {
			if strings.TrimSpace(c) == "" {
				t.Errorf("a changelog row leaves cell %d empty: %s", i+1, line)
			}
		}
		rows[cells[0]] = cells
	}
	for _, ruling := range []string{"D-62", "D-66", "D-67", "D-70", "D-74"} {
		if !strings.Contains(log, ruling) {
			t.Errorf("the changelog does not cite %s", ruling)
		}
	}
	for _, name := range []string{"The borrow check", "`Changed()`", "Loop playback", "`Describe`", "`Purge`", "`Fetcher`", "New names' shapes"} {
		if rows[name] == nil {
			t.Errorf("the changelog has no row for %s", name)
		}
	}
	known := knownNames(t)
	for _, cells := range rows {
		state := cells[4]
		if !strings.HasPrefix(state, "landed") && !strings.HasPrefix(state, "lands in") && !strings.HasPrefix(state, "lands with") {
			t.Errorf("row %s: state %q is neither landed nor lands in/with", cells[0], state)
			continue
		}
		if !strings.HasPrefix(state, "landed") {
			continue
		}
		_, removed, _ := strings.Cut(cells[1], "Removed")
		for _, span := range regexp.MustCompile("`([^`]*)`").FindAllStringSubmatch(removed, -1) {
			if m := apiSpan.FindStringSubmatch(span[1]); m != nil && known[m[1]] {
				t.Errorf("row %s is marked landed, but %s is still in the package", cells[0], m[1])
			}
		}
	}
}
