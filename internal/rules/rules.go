// Package rules holds the library's static rules: what its source may not
// contain, checked by reading the source. The library starts no goroutine
// and owns no clock, writes nothing to standard output, opens connections
// in one package only, lets third-party code into one package only, and
// keeps the import edges its architecture depends on.
//
// Like the test kit, this package is imported by test files only.
package rules

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// The rules a finding can name.
const (
	RuleGoroutine = "goroutine" // a go statement in a non-test file (D-73)
	RuleTimer     = "timer"     // a timer, ticker, after or sleep of the library's own (D-73, FR-25)
	RuleOutput    = "output"    // a write to standard output or standard error (NFR-20)
	RuleImport    = "import"    // an import the architecture forbids
	RuleTestOnly  = "test-only" // the test kit or these rules imported by a non-test file
	RuleDialHook  = "dial-hook" // a test package that does not run under the loopback-only guard (NFR-11)
)

// maxFiles bounds a check, so a wrong root cannot make it run away.
const maxFiles = 10000

// Finding is one broken rule, at one place.
type Finding struct {
	File   string // slash-separated, relative to the root
	Line   int
	Rule   string
	Detail string
}

func (f Finding) String() string {
	return fmt.Sprintf("%s:%d: [%s] %s", f.File, f.Line, f.Rule, f.Detail)
}

// Check reads every non-test Go file of the library under root — the files
// in root itself and everything under root/internal — and returns what
// breaks a rule. Separate modules and test data are not the library.
func Check(root, module string) ([]Finding, error) {
	if root == "" {
		return nil, errors.New("rules: empty root")
	}
	if module == "" {
		return nil, errors.New("rules: empty module path")
	}
	c := &checker{root: root, module: module, fset: token.NewFileSet(), hooked: map[string]bool{}, tested: map[string]int{}}
	err := filepath.WalkDir(root, c.visit)
	if err != nil {
		return nil, err
	}
	for dir := range c.tested {
		if !c.hooked[dir] {
			c.found = append(c.found, Finding{File: dir, Line: c.tested[dir], Rule: RuleDialHook, Detail: "this test package has no TestMain that calls testkit.Main, so its tests are not held to loopback"})
		}
	}
	sort.Slice(c.found, func(i, j int) bool {
		if c.found[i].File != c.found[j].File {
			return c.found[i].File < c.found[j].File
		}
		return c.found[i].Line < c.found[j].Line
	})
	return c.found, nil
}

type checker struct {
	root, module string
	fset         *token.FileSet
	found        []Finding
	files        int
	hooked       map[string]bool // package directory → a test file calls testkit.Main
	tested       map[string]int  // package directory → 1, when it has test files
}

// visit is the walk's callback: it keeps the walk inside the library and
// hands each Go file to checkFile.
func (c *checker) visit(full string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(c.root, full)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)
	if d.IsDir() {
		if rel == "." || rel == "internal" || strings.HasPrefix(rel, "internal/") && path.Base(rel) != "testdata" {
			return nil
		}
		return filepath.SkipDir
	}
	if !strings.HasSuffix(rel, ".go") {
		return nil
	}
	if c.files >= maxFiles {
		return fmt.Errorf("rules: more than %d Go files under %s", maxFiles, c.root)
	}
	c.files++
	return c.checkFile(full, rel)
}

// checkFile parses one file and applies the rules that fit it.
func (c *checker) checkFile(full, rel string) error {
	if full == "" || rel == "" {
		return errors.New("rules: empty file name")
	}
	file, err := parser.ParseFile(c.fset, full, nil, parser.SkipObjectResolution)
	if err != nil {
		return fmt.Errorf("rules: %s does not parse: %w", rel, err)
	}
	dir := path.Dir(rel)
	names, err := importNames(file)
	if err != nil {
		return fmt.Errorf("rules: %s: %w", rel, err)
	}
	if strings.HasSuffix(rel, "_test.go") {
		c.tested[dir] = 1
		hooked, err := callsHook(file, names, c.module, dir)
		if err != nil {
			return err
		}
		if hooked {
			c.hooked[dir] = true
		}
		return nil
	}
	// The test kit and these rules are test-only tooling: they may start
	// goroutines, sleep and print.
	for _, tool := range []string{"internal/testkit", "internal/rules"} {
		if dir == tool || strings.HasPrefix(dir, tool+"/") {
			return nil
		}
	}
	if err := c.checkImports(file, rel, dir); err != nil {
		return err
	}
	return c.checkBody(file, rel, names)
}

// importNames maps each import's local name to its path.
func importNames(file *ast.File) (map[string]string, error) {
	if file == nil {
		return nil, errors.New("nil file")
	}
	names := make(map[string]string, len(file.Imports))
	for _, spec := range file.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("import path %s: %w", spec.Path.Value, err)
		}
		name := path.Base(p)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		names[name] = p
	}
	return names, nil
}

// callsHook reports whether a test file calls the test kit's Main.
func callsHook(file *ast.File, names map[string]string, module, dir string) (bool, error) {
	if file == nil || names == nil {
		return false, errors.New("rules: callsHook needs a parsed file and its imports")
	}
	if module == "" {
		return false, errors.New("rules: empty module path")
	}
	kit := module + "/internal/testkit"
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			if x, ok := fun.X.(*ast.Ident); ok && fun.Sel.Name == "Main" && names[x.Name] == kit {
				found = true
			}
		case *ast.Ident:
			if fun.Name == "Main" && dir == "internal/testkit" {
				found = true
			}
		}
		return !found
	})
	return found, nil
}

// checkImports applies the import rules to one non-test file.
func (c *checker) checkImports(file *ast.File, rel, dir string) error {
	if file == nil {
		return errors.New("rules: checkImports needs a parsed file")
	}
	if rel == "" || dir == "" {
		return errors.New("rules: checkImports needs the file's name and directory")
	}
	for _, spec := range file.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return fmt.Errorf("rules: %s: import path %s: %w", rel, spec.Path.Value, err)
		}
		rule, detail, err := judgeImport(p, dir, c.module)
		if err != nil {
			return err
		}
		if rule != "" {
			c.found = append(c.found, Finding{rel, c.fset.Position(spec.Pos()).Line, rule, detail})
		}
	}
	return nil
}

// judgeImport decides whether the package in dir may import p. It returns
// the rule broken and what to say, or an empty rule.
func judgeImport(p, dir, module string) (rule, detail string, err error) {
	if p == "" || dir == "" {
		return "", "", errors.New("rules: judgeImport needs an import path and a directory")
	}
	if module == "" {
		return "", "", errors.New("rules: empty module path")
	}
	if name, inside := strings.CutPrefix(p, module+"/internal/"); inside {
		return judgeInternalImport(name, path.Base(dir))
	}
	return judgeOutsideImport(p, dir, module)
}

// judgeInternalImport decides whether package pkg may import the library's
// own internal package called name.
func judgeInternalImport(name, pkg string) (rule, detail string, err error) {
	if name == "" || pkg == "" {
		return "", "", errors.New("rules: judgeInternalImport needs both package names")
	}
	// The internal packages a package may never import: render reads only
	// what is already on hand, and work knows jobs only through scene.
	var banned []string
	switch pkg {
	case "render":
		banned = []string{"tiles", "fetch", "overlay", "mvt", "archive"}
	case "work":
		banned = []string{"tiles", "overlay", "describe", "render", "fetch", "mvt", "archive"}
	}
	switch {
	case name == "testkit", name == "rules":
		return RuleTestOnly, "internal/" + name + " is for test files only", nil
	case pkg == "scene":
		return RuleImport, "scene imports internal/" + name + "; it may import the standard library only", nil
	case slices.Contains(banned, name):
		return RuleImport, pkg + " may never import " + name, nil
	}
	return "", "", nil
}

// judgeOutsideImport decides whether the package in dir may import p, a
// package from outside the library: only fetch opens a connection, and only
// textsafe lets third-party code in.
func judgeOutsideImport(p, dir, module string) (rule, detail string, err error) {
	if p == "" || dir == "" {
		return "", "", errors.New("rules: judgeOutsideImport needs an import path and a directory")
	}
	if module == "" {
		return "", "", errors.New("rules: empty module path")
	}
	first, _, _ := strings.Cut(p, "/")
	ours := p == module || strings.HasPrefix(p, module+"/")
	switch {
	case slices.Contains([]string{"net", "net/http"}, p), strings.HasPrefix(p, "net/http/"):
		if dir != "internal/fetch" {
			return RuleImport, p + " is imported outside internal/fetch; only fetch opens a connection", nil
		}
	case strings.Contains(first, ".") && !ours:
		if dir != "internal/textsafe" {
			return RuleImport, p + " is third-party code outside internal/textsafe", nil
		}
	}
	return "", "", nil
}

// checkBody applies the goroutine, timer and output rules to one non-test file.
func (c *checker) checkBody(file *ast.File, rel string, names map[string]string) error {
	if file == nil || names == nil {
		return errors.New("rules: checkBody needs a parsed file and its imports")
	}
	if rel == "" {
		return errors.New("rules: checkBody needs the file's name")
	}
	var failed error
	ast.Inspect(file, func(n ast.Node) bool {
		rule, detail := "", ""
		switch node := n.(type) {
		case *ast.GoStmt:
			rule, detail = RuleGoroutine, "a go statement; the library starts no goroutine"
		case *ast.CallExpr:
			if id, ok := node.Fun.(*ast.Ident); ok && (id.Name == "print" || id.Name == "println") {
				rule, detail = RuleOutput, "the "+id.Name+" builtin writes to standard error"
			}
		case *ast.SelectorExpr:
			if x, ok := node.X.(*ast.Ident); ok && names[x.Name] != "" {
				rule, detail, failed = judgeSelector(names[x.Name], node.Sel.Name)
			}
		}
		if rule != "" {
			c.found = append(c.found, Finding{rel, c.fset.Position(n.Pos()).Line, rule, detail})
		}
		return failed == nil
	})
	return failed
}

// judgeSelector decides whether a non-test file may name pkg.sel: a clock
// of the library's own, or a way to standard output or standard error.
func judgeSelector(pkg, sel string) (rule, detail string, err error) {
	if pkg == "" || sel == "" {
		return "", "", errors.New("rules: judgeSelector needs a package path and a name")
	}
	switch pkg {
	case "time":
		if slices.Contains([]string{"NewTimer", "NewTicker", "After", "AfterFunc", "Tick", "Sleep"}, sel) {
			return RuleTimer, "time." + sel + "; the library owns no clock", nil
		}
	case "fmt":
		if slices.Contains([]string{"Print", "Printf", "Println"}, sel) {
			return RuleOutput, "fmt." + sel + " writes to standard output", nil
		}
	case "os":
		if sel == "Stdout" || sel == "Stderr" {
			return RuleOutput, "os." + sel, nil
		}
	}
	return "", "", nil
}
