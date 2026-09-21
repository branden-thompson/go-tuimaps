package rules

import (
	"strings"
	"testing"
)

// TestInvisibleCharactersInSourceAreFound: a source file may name any
// character by its escape, and may hold none of the invisible ones
// literally - they hide what the code says from the person reading it.
func TestInvisibleCharactersInSourceAreFound(t *testing.T) {
	const head = "// Package tiles is planted.\npackage tiles\n\n"
	invisible := map[string]string{
		"a zero-width space":        "\u200B",
		"a zero-width joiner":       "\u200D",
		"a right-to-left override":  "\u202E",
		"a left-to-right isolate":   "\u2066",
		"a soft hyphen":             "\u00AD",
		"a word joiner":             "\u2060",
		"a tag character":           "\U000E0041",
		"an escape byte":            "\x1b",
		"a line separator":          "\u2028",
		"an interlinear annotation": "\uFFF9",
	}
	for name, ch := range invisible {
		for _, file := range []string{"internal/tiles/a.go", "internal/tiles/a_test.go", "internal/testkit/a.go"} {
			src := head + "var name = \"x" + ch + "y\"\n"
			if strings.HasSuffix(file, "_test.go") {
				src = strings.Replace(hook, "package p", "package tiles", 1) + "\nvar name = \"x" + ch + "y\"\n"
			}
			if strings.Contains(file, "testkit") {
				src = strings.Replace(src, "tiles", "testkit", 2)
			}
			files := map[string]string{file: src}
			if strings.HasSuffix(file, "_test.go") {
				files["internal/tiles/doc.go"] = head
			}
			got, err := Check(writeTree(t, files), mod)
			if err != nil {
				t.Fatalf("%s in %s: %v", name, file, err)
			}
			if len(got) != 1 || got[0].Rule != RuleInvisible || got[0].File != file {
				t.Errorf("%s in %s: got %v; want one %q finding", name, file, got, RuleInvisible)
			}
		}
	}
}

func TestEscapesAndVisibleTextArePermitted(t *testing.T) {
	src := "// Package tiles is planted. Z\u00fcrich, \u6771\u4eac and a dash \u2014 are visible.\n" +
		"package tiles\n\nvar names = []string{\"\\u200B\", \"\\u202E\", \"e\u0301\", \"\t\"}\n"
	got, err := Check(writeTree(t, map[string]string{"internal/tiles/a.go": src}), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("unexpected %v; an escape names a character without hiding it, and accents, ideographs and tabs are visible", got)
	}
}
