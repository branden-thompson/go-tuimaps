package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckRefusesWhatWouldWriteABrokenPage: each guard must refuse its case,
// and a sound set must pass, so a guard that stopped matching is seen.
func TestCheckRefusesWhatWouldWriteABrokenPage(t *testing.T) {
	good := []diagram{{feature: "radar-loops", file: "06_docs/a.md", title: "A", src: "graph TD\n  a-->b\n"}}
	whole := "<!--NAV--><!--BODY--><!--COUNT-->"
	if err := check(good, whole); err != nil {
		t.Fatalf("a sound page was refused: %v", err)
	}
	cases := []struct {
		name  string
		found []diagram
		page  string
		want  string
	}{
		{"no diagrams", nil, whole, "no mermaid blocks"},
		{"a template without the body", good, "<!--NAV--><!--COUNT-->", "<!--BODY-->"},
		{"a template without the nav", good, "<!--BODY--><!--COUNT-->", "<!--NAV-->"},
		{"a template without the count", good, "<!--NAV--><!--BODY-->", "<!--COUNT-->"},
		{"a diagram with no source", []diagram{{feature: "docs", file: "06_docs/b.md", src: " \n"}}, whole, "06_docs/b.md"},
		{"a feature that is not safe as an id", []diagram{{feature: `x" onload="y`, file: "06_docs/c.md", src: "graph TD\n"}}, whole, "not safe"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := check(c.found, c.page)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want an error naming %q", err, c.want)
			}
		})
	}
}

// TestThePageMatchesTheDocuments: the committed page is what the documents
// build, so a diagram edited in Markdown cannot leave the page stale
// (v0.2.0 D-38). Rebuild it with: cd tools/atlas && go run . -root ../..
func TestThePageMatchesTheDocuments(t *testing.T) {
	root := filepath.Join("..", "..")
	want, _, err := build(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, dest))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s is stale: rebuild it with `cd tools/atlas && go run . -root ../..`", dest)
	}
}
