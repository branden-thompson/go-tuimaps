package rules

import (
	"strings"
	"testing"
)

// TestUndocumentedExportedNamesAreFound is plan task 00.8.
func TestUndocumentedExportedNamesAreFound(t *testing.T) {
	const head = "// Package tiles is planted.\npackage tiles\n\n"
	cases := []struct {
		name string
		src  string
		want string // the name the finding must mention; empty means no finding
	}{
		{"an exported function", head + "func Open() {}\n", "Open"},
		{"a documented function", head + "// Open opens.\nfunc Open() {}\n", ""},
		{"an unexported function", head + "func open() {}\n", ""},
		{"an exported type", head + "type Cache struct{}\n", "Cache"},
		{"a method on an exported type", head + "// Cache is one.\ntype Cache struct{}\n\nfunc (Cache) Get() {}\n", "Get"},
		{"a method on a pointer to an exported type", head + "// Cache is one.\ntype Cache struct{}\n\nfunc (c *Cache) Put() {}\n", "Put"},
		{"a method on an unexported type", head + "type cache struct{}\n\nfunc (cache) Get() {}\n", ""},
		{"an exported constant", head + "const Limit = 1\n", "Limit"},
		{"a constant group with one comment for the group", head + "// The limits.\nconst (\n\tLimit = 1\n\tOther = 2\n)\n", ""},
		{"a group where one name has no comment of its own", head + "const (\n\t// Limit is one.\n\tLimit = 1\n\tOther = 2\n)\n", "Other"},
		{"a trailing comment counts", head + "const (\n\tLimit = 1 // the limit\n)\n", ""},
		{"an exported variable", head + "var ErrGone = 1\n", "ErrGone"},
		{"no package comment", "package tiles\n", "package comment"},
	}
	for _, c := range cases {
		root := writeTree(t, map[string]string{"internal/tiles/a.go": c.src})
		got, err := Check(root, mod)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if c.want == "" {
			if len(got) != 0 {
				t.Errorf("%s: unexpected %v", c.name, got)
			}
			continue
		}
		if len(got) != 1 || got[0].Rule != RuleUndocumented || !strings.Contains(got[0].Detail, c.want) {
			t.Errorf("%s: got %v; want one %q finding naming %q", c.name, got, RuleUndocumented, c.want)
		}
	}
}

func TestOnePackageCommentServesTheWholePackage(t *testing.T) {
	root := writeTree(t, map[string]string{
		"internal/tiles/doc.go": "// Package tiles is planted.\npackage tiles\n",
		"internal/tiles/a.go":   "package tiles\n",
	})
	got, err := Check(root, mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("unexpected %v", got)
	}
}

func TestTheTestKitIsHeldToTheDocumentationRuleToo(t *testing.T) {
	root := writeTree(t, map[string]string{
		"internal/testkit/a.go": "// Package testkit is planted.\npackage testkit\n\nfunc Helper() { go Helper() }\n",
	})
	got, err := Check(root, mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Rule != RuleUndocumented {
		t.Errorf("got %v; the test kit may start goroutines, but its exported names are documented like any other", got)
	}
}
