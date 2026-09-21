package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

const mod = "example.com/lib"

// hook is the test file every planted package needs, so that a case plants
// exactly one violation.
const hook = "package p\nimport (\n\t\"os\"\n\t\"testing\"\n\t\"example.com/lib/internal/testkit\"\n)\nfunc TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }\n"

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, src := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// withDocs gives every planted non-test file a package comment, so that a
// case about some other rule plants exactly one violation.
func withDocs(files map[string]string) map[string]string {
	out := make(map[string]string, len(files))
	for rel, src := range files {
		if !strings.HasSuffix(rel, "_test.go") && strings.HasPrefix(src, "package ") {
			src = "// Package planted is planted.\n" + src
		}
		out[rel] = src
	}
	return out
}

func TestACleanTreePasses(t *testing.T) {
	root := writeTree(t, withDocs(map[string]string{
		"doc.go":                      "// Package lib.\npackage lib\n",
		"internal/tiles/tiles.go":     "package tiles\nimport \"example.com/lib/internal/scene\"\nvar _ scene.T\n",
		"internal/tiles/hook_test.go": strings.Replace(hook, "package p", "package tiles", 1),
		"internal/scene/scene.go":     "package scene\n\n// T is planted.\ntype T int\n",
		"internal/fetch/fetch.go":     "package fetch\nimport \"net/http\"\nvar _ http.Client\n",
		"internal/textsafe/w.go":      "package textsafe\nimport \"github.com/mattn/go-runewidth\"\nvar _ = runewidth.StringWidth\n",
		"cmd/app/main.go":             "package main\nimport \"fmt\"\nfunc main() { go fmt.Println(\"a separate module is not the library\") }\n",
		"testdata/x/bad.go":           "package x\nfunc f() { go f() }\n",
	}))
	got, err := Check(root, mod)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		t.Errorf("unexpected finding: %s", f)
	}
}

func TestEveryPlantedViolationIsFound(t *testing.T) {
	cases := []struct {
		name string
		file string
		src  string
		rule string
	}{
		{"a go statement", "internal/tiles/a.go", "package tiles\nfunc f() { go f() }\n", RuleGoroutine},
		{"a timer", "internal/tiles/a.go", "package tiles\nimport \"time\"\nfunc f() { time.NewTimer(time.Second) }\n", RuleTimer},
		{"a ticker", "internal/work/a.go", "package work\nimport \"time\"\nvar _ = time.NewTicker\n", RuleTimer},
		{"time.After", "internal/work/a.go", "package work\nimport \"time\"\nfunc f() { <-time.After(1) }\n", RuleTimer},
		{"a sleep", "internal/work/a.go", "package work\nimport \"time\"\nfunc f() { time.Sleep(1) }\n", RuleTimer},
		{"a renamed time import", "internal/work/a.go", "package work\nimport clock \"time\"\nfunc f() { clock.AfterFunc(1, f) }\n", RuleTimer},
		{"a print", "internal/render/a.go", "package render\nimport \"fmt\"\nfunc f() { fmt.Println(1) }\n", RuleOutput},
		{"standard error", "map.go", "package lib\nimport \"os\"\nvar w = os.Stderr\n", RuleOutput},
		{"the println builtin", "internal/render/a.go", "package render\nfunc f() { println(1) }\n", RuleOutput},
		{"render importing tiles", "internal/render/a.go", "package render\nimport _ \"example.com/lib/internal/tiles\"\n", RuleImport},
		{"render importing overlay", "internal/render/a.go", "package render\nimport _ \"example.com/lib/internal/overlay\"\n", RuleImport},
		{"work importing a supplier of jobs", "internal/work/a.go", "package work\nimport _ \"example.com/lib/internal/describe\"\n", RuleImport},
		{"scene importing anything here", "internal/scene/a.go", "package scene\nimport _ \"example.com/lib/internal/project\"\n", RuleImport},
		{"a connection outside fetch", "internal/tiles/a.go", "package tiles\nimport _ \"net/http\"\n", RuleImport},
		{"a dial outside fetch", "internal/tiles/a.go", "package tiles\nimport _ \"net\"\n", RuleImport},
		{"third-party code outside textsafe", "internal/render/a.go", "package render\nimport _ \"github.com/mattn/go-runewidth\"\n", RuleImport},
		{"the test kit in a non-test file", "internal/tiles/a.go", "package tiles\nimport _ \"example.com/lib/internal/testkit\"\n", RuleTestOnly},
		{"the rules in a non-test file", "internal/tiles/a.go", "package tiles\nimport _ \"example.com/lib/internal/rules\"\n", RuleTestOnly},
	}
	for _, c := range cases {
		dir := filepath.ToSlash(filepath.Dir(c.file))
		pkg := filepath.Base(dir)
		if dir == "." {
			pkg = "lib"
		}
		root := writeTree(t, withDocs(map[string]string{
			c.file:                c.src,
			dir + "/hook_test.go": strings.Replace(hook, "package p", "package "+pkg, 1),
		}))
		got, err := Check(root, mod)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if len(got) != 1 || got[0].Rule != c.rule {
			t.Errorf("%s: got %v; want exactly one finding of rule %q", c.name, got, c.rule)
			continue
		}
		if got[0].File != c.file || got[0].Line == 0 {
			t.Errorf("%s: finding %v does not name %s and a line", c.name, got[0], c.file)
		}
	}
}

func TestATestPackageWithoutTheDialHookIsFound(t *testing.T) {
	root := writeTree(t, withDocs(map[string]string{
		"internal/tiles/tiles.go":      "package tiles\n",
		"internal/tiles/tiles_test.go": "package tiles\nimport \"testing\"\nfunc TestX(t *testing.T) {}\n",
	}))
	got, err := Check(root, mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Rule != RuleDialHook {
		t.Errorf("got %v; want one finding of rule %q", got, RuleDialHook)
	}
}

func TestTestFilesAndTheTestKitMayStartGoroutines(t *testing.T) {
	body := "import \"time\"\nfunc f() { go f(); time.Sleep(1) }\n"
	root := writeTree(t, withDocs(map[string]string{
		"internal/tiles/tiles.go":      "package tiles\n",
		"internal/tiles/tiles_test.go": "package tiles\n" + body,
		"internal/tiles/hook_test.go":  strings.Replace(hook, "package p", "package tiles", 1),
		"internal/testkit/leak.go":     "package testkit\n" + body,
		"internal/rules/rules.go":      "package rules\nimport \"fmt\"\nfunc g() { fmt.Println() }\n",
	}))
	got, err := Check(root, mod)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		t.Errorf("unexpected finding: %s", f)
	}
}

func TestCheckRefusesBadArguments(t *testing.T) {
	if _, err := Check("", mod); err == nil {
		t.Error("an empty root must be an error")
	}
	if _, err := Check(t.TempDir(), ""); err == nil {
		t.Error("an empty module path must be an error")
	}
	root := writeTree(t, map[string]string{"internal/tiles/a.go": "package tiles\nfunc {"})
	if _, err := Check(root, mod); err == nil {
		t.Error("a file that does not parse must be an error, not a pass")
	}
}
