// Package testkit holds helpers for the library's tests: the pinned
// fixture's loaders, and — as the scaffold grows — the leak check, the
// loopback-only dial hook and the blocking transport. It is imported by
// test files only; a static check says so.
package testkit

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// hashList is the name of the fixture's list of hashes.
	hashList = "HASHES"
	// readme is the fixture's description; it is the one file the list
	// does not pin, because it describes the list.
	readme = "README.md"
	// maxEntries bounds the list, so a damaged file cannot make the check
	// run away.
	maxEntries = 4096
	// maxClimb bounds the walk from the working directory to the module root.
	maxClimb = 16
)

// entry is one line of the hash list.
type entry struct {
	sum  string
	rel  string
	size int64
}

// FixtureRoot returns the pinned fixture's directory, found by climbing from
// the working directory to the module root.
func FixtureRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("testkit: working directory: %w", err)
	}
	for range maxClimb {
		root := filepath.Join(dir, "testdata", "fixture")
		if _, err := os.Stat(filepath.Join(root, hashList)); err == nil {
			return root, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.New("testkit: testdata/fixture/HASHES not found above the working directory")
}

// cleanRel validates a slash-separated path relative to the fixture's root
// and refuses anything that could leave it.
func cleanRel(rel string) (string, error) {
	if rel == "" {
		return "", errors.New("testkit: empty fixture path")
	}
	if strings.HasPrefix(rel, "/") || strings.Contains(rel, "\\") {
		return "", fmt.Errorf("testkit: fixture path %q is not relative and slash-separated", rel)
	}
	for part := range strings.SplitSeq(rel, "/") {
		if part == ".." || part == "." || part == "" {
			return "", fmt.Errorf("testkit: fixture path %q contains %q", rel, part)
		}
	}
	return rel, nil
}

// parseEntry reads one line of the list: a SHA-256 in hex, a relative path
// and a size in bytes, separated by runs of spaces.
func parseEntry(line string) (entry, error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return entry{}, fmt.Errorf("testkit: %s: line %q has %d fields, want 3", hashList, line, len(fields))
	}
	if len(fields[0]) != hex.EncodedLen(sha256.Size) {
		return entry{}, fmt.Errorf("testkit: %s: %q is not a SHA-256", hashList, fields[0])
	}
	rel, err := cleanRel(fields[1])
	if err != nil {
		return entry{}, err
	}
	size, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || size < 0 {
		return entry{}, fmt.Errorf("testkit: %s: size %q of %s is not a byte count", hashList, fields[2], rel)
	}
	return entry{sum: strings.ToLower(fields[0]), rel: rel, size: size}, nil
}

// readList parses the hash list under root.
func readList(root string) ([]entry, error) {
	if root == "" {
		return nil, errors.New("testkit: empty fixture root")
	}
	data, err := os.ReadFile(filepath.Join(root, hashList))
	if err != nil {
		return nil, fmt.Errorf("testkit: %w", err)
	}
	var entries []entry
	for raw := range strings.SplitSeq(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if len(entries) >= maxEntries {
			return nil, fmt.Errorf("testkit: %s lists more than %d files", hashList, maxEntries)
		}
		e, err := parseEntry(line)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("testkit: %s lists no files", hashList)
	}
	return entries, nil
}

// checkEntry compares one listed file with what is on disk.
func checkEntry(root string, e entry) error {
	if root == "" || e.rel == "" {
		return errors.New("testkit: empty root or path")
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(e.rel)))
	if err != nil {
		return fmt.Errorf("testkit: fixture file %s: %w", e.rel, err)
	}
	if int64(len(data)) != e.size {
		return fmt.Errorf("testkit: fixture file %s is %d bytes, the list says %d", e.rel, len(data), e.size)
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != e.sum {
		return fmt.Errorf("testkit: fixture file %s does not match its pinned SHA-256", e.rel)
	}
	return nil
}

// VerifyFixture checks that every file the list names matches its hash and
// size, and that no file is present that the list does not name.
func VerifyFixture(root string) error {
	entries, err := readList(root)
	if err != nil {
		return err
	}
	listed := make(map[string]bool, len(entries))
	for _, e := range entries {
		if listed[e.rel] {
			return fmt.Errorf("testkit: %s names %s twice", hashList, e.rel)
		}
		listed[e.rel] = true
		if err := checkEntry(root, e); err != nil {
			return err
		}
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == hashList || rel == readme || listed[rel] {
			return nil
		}
		return fmt.Errorf("testkit: fixture file %s is not in %s", rel, hashList)
	})
}

// LoadFixture returns one fixture file's bytes. The file must be in the
// list, and it is checked against the list before it is returned.
func LoadFixture(root, rel string) ([]byte, error) {
	rel, err := cleanRel(rel)
	if err != nil {
		return nil, err
	}
	entries, err := readList(root)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.rel != rel {
			continue
		}
		if err := checkEntry(root, e); err != nil {
			return nil, err
		}
		return os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	}
	return nil, fmt.Errorf("testkit: %s is not in %s", rel, hashList)
}
