package testkit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFixturePinned is plan task 00.9: every file of the pinned fixture
// matches the committed list, and nothing unlisted has crept in.
func TestFixturePinned(t *testing.T) {
	root, err := FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyFixture(root); err != nil {
		t.Fatal(err)
	}
}

// writeFixture builds a small fixture in a temporary directory and returns
// its root. The list is correct for the files as written.
func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	var list strings.Builder
	list.WriteString("# SHA-256  path  bytes\n")
	for rel, body := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256([]byte(body))
		fmt.Fprintf(&list, "%s  %s  %d\n", hex.EncodeToString(sum[:]), rel, len(body))
	}
	if err := os.WriteFile(filepath.Join(root, "HASHES"), []byte(list.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestVerifyFixtureAcceptsACorrectList(t *testing.T) {
	root := writeFixture(t, map[string]string{"a/one.bin": "one", "two.bin": "two"})
	if err := VerifyFixture(root); err != nil {
		t.Fatalf("a correct fixture was refused: %v", err)
	}
}

func TestVerifyFixtureCatchesEveryKindOfDrift(t *testing.T) {
	files := map[string]string{"a/one.bin": "one", "two.bin": "two"}
	cases := []struct {
		name   string
		damage func(root string) error
		want   string
	}{
		{"a changed byte", func(r string) error { return os.WriteFile(filepath.Join(r, "two.bin"), []byte("twO"), 0o644) }, "two.bin"},
		{"a changed length", func(r string) error { return os.WriteFile(filepath.Join(r, "two.bin"), []byte("twoo"), 0o644) }, "two.bin"},
		{"a missing file", func(r string) error { return os.Remove(filepath.Join(r, "a", "one.bin")) }, "a/one.bin"},
		{"an unlisted file", func(r string) error { return os.WriteFile(filepath.Join(r, "a", "extra.bin"), []byte("x"), 0o644) }, "a/extra.bin"},
		{"a malformed line", func(r string) error {
			return os.WriteFile(filepath.Join(r, "HASHES"), []byte("not-a-hash  two.bin\n"), 0o644)
		}, "HASHES"},
		{"a path that escapes", func(r string) error {
			return os.WriteFile(filepath.Join(r, "HASHES"), []byte(strings.Repeat("0", 64)+"  ../two.bin  3\n"), 0o644)
		}, ".."},
	}
	for _, c := range cases {
		root := writeFixture(t, files)
		if err := c.damage(root); err != nil {
			t.Fatal(err)
		}
		err := VerifyFixture(root)
		if err == nil {
			t.Errorf("%s: not caught", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %q does not name %q", c.name, err, c.want)
		}
	}
}

func TestLoadFixtureReadsAListedFileOnly(t *testing.T) {
	root := writeFixture(t, map[string]string{"a/one.bin": "one"})
	got, err := LoadFixture(root, "a/one.bin")
	if err != nil || string(got) != "one" {
		t.Fatalf("got %q, %v", got, err)
	}
	for _, bad := range []string{"", "../HASHES", "/etc/hosts", "a/../../x", "a/missing.bin"} {
		if _, err := LoadFixture(root, bad); err == nil {
			t.Errorf("LoadFixture(%q) succeeded; it must be refused", bad)
		}
	}
}
