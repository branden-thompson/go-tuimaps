package tuimaps_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestReadmeQuickStartBuilds is plan task 12.15: the quick start in the
// README is taken out of the file and parsed, so it cannot rot unnoticed.
// Its calls are the ones the example test runs, which is what proves it
// works; this holds it to being real Go that names the real calls.
func TestReadmeQuickStartBuilds(t *testing.T) {
	body, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	blocks := goBlocks(string(body))
	if len(blocks) == 0 {
		t.Fatal("the README has no Go in it at all")
	}
	for i, block := range blocks {
		if _, err := parser.ParseFile(token.NewFileSet(), "readme.go", block, parser.AllErrors); err != nil {
			t.Errorf("block %d of the README is not Go: %v", i, err)
		}
	}
	quick := blocks[0]
	for _, call := range []string{"tuimaps.New", "tuimaps.WithSize", "tuimaps.Embed", "assets.Tile", "m.Settle", "m.Render", "m.Close"} {
		if !strings.Contains(quick, call) {
			t.Errorf("the quick start does not use %s", call)
		}
	}
	// The three calls are three: create, settle, render.
	if n := strings.Count(quick, "m.Settle") + strings.Count(quick, "m.Render") + strings.Count(quick, "tuimaps.New"); n != 3 {
		t.Errorf("the quick start makes %d of the three calls", n)
	}
}

// goBlocks are the fenced Go blocks of a markdown file.
func goBlocks(body string) []string {
	var out []string
	parts := strings.Split(body, "```")
	for i := 1; i < len(parts); i += 2 {
		block := parts[i]
		if head, rest, ok := strings.Cut(block, "\n"); ok && strings.TrimSpace(head) == "go" {
			out = append(out, rest)
		}
	}
	return out
}

// TestReadmeSaysWhatIsOwed: the README tells a host the things it cannot
// find out for itself - the braille font, what is sent and stored, the
// credit the data asks for, and what is not built yet.
func TestReadmeSaysWhatIsOwed(t *testing.T) {
	body, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, owed := range []string{"braille", "NO_COLOR", "OpenStreetMap", "About", "CacheRoot", "Not built yet", "MIT"} {
		if !strings.Contains(text, owed) {
			t.Errorf("the README does not mention %q", owed)
		}
	}
}
