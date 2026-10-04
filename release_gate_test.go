package tuimaps_test

// release_gate_test.go — v0.2.0 L-5.3 and L-5.5 (plan tasks L10.3, L10.4):
// the release check refuses a final tag while its checklist is unfinished,
// holds no release candidate to it, and requires the pinned scanner to find
// nothing reachable but what the checklist carries as a ruled exception.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// releaseTree is a planted tree with a checklist of the given rows, and a
// directory holding a stub govulncheck that answers -version with version
// and a scan with out, exiting rc.
func releaseTree(t *testing.T, rows, version, out, rc string) (root string, env []string) {
	t.Helper()
	root = plantTree(t, true)
	writeFile(t, root, "06_docs/02_features/radar-loops/07-readiness/release-checklist.md",
		"# Release checklist\n\n| # | Check | Done |\n|---|---|---|\n"+rows)
	writeFile(t, root, "internal/fetch/fetch.go", "package fetch\n\nconst (\n\tVersion = \"0.2.0\"\n)\n")
	bin := t.TempDir()
	writeFile(t, bin, "govulncheck", "#!/bin/sh\nif [ \"$1\" = \"-version\" ]; then echo 'Go: go1.27.1'; echo 'Scanner: govulncheck@"+version+"'; exit 0; fi\n"+
		"printf '%s\\n' '"+out+"'\nexit "+rc+"\n")
	if err := os.Chmod(filepath.Join(bin, "govulncheck"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root, []string{"PATH=" + bin + ":" + os.Getenv("PATH")}
}

const (
	doneRow = "| 1 | The gate is green on a fresh clone | [x] |\n"
	openRow = "| 2 | M1 is scored | [ ] |\n"
)

// TestAFinalTagWaitsForItsChecklist: an unticked row refuses the final tag
// and names the row; every row ticked passes.
func TestAFinalTagWaitsForItsChecklist(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root, env := releaseTree(t, doneRow+openRow, "v1.8.0", "No vulnerabilities found.", "0")
	out, err := runGateWith(t, root, env, "--release", "v0.2.0")
	if err == nil || !strings.Contains(out, "M1 is scored") {
		t.Errorf("a final tag with a row open was not refused by name:\n%s", out)
	}
	root, env = releaseTree(t, doneRow, "v1.8.0", "No vulnerabilities found.", "0")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err != nil || !strings.Contains(out, "release check green") {
		t.Errorf("a final tag with every row ticked was refused: %v\n%s", err, out)
	}
}

// TestAReleaseCandidateIsNotHeldToTheChecklist: rows open, an rc tag passes,
// and says how many are open.
func TestAReleaseCandidateIsNotHeldToTheChecklist(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root, env := releaseTree(t, doneRow+openRow, "v1.8.0", "No vulnerabilities found.", "0")
	out, err := runGateWith(t, root, env, "--release", "v0.2.0-rc.35")
	if err != nil || !strings.Contains(out, "1 row(s) open") {
		t.Errorf("a release candidate was refused or did not count the open rows: %v\n%s", err, out)
	}
}

// TestAReachableFindingNeedsARuledException: a finding the scanner calls
// reachable refuses the tag, unless a checklist row carries its id with a
// ruling; a scanner that is not the pinned one refuses it outright.
func TestAReachableFindingNeedsARuledException(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	found := "Vulnerability #1: GO-2099-0001"
	root, env := releaseTree(t, doneRow, "v1.8.0", found, "3")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err == nil || !strings.Contains(out, "GO-2099-0001 is reachable") {
		t.Errorf("an injected reachable finding did not fail the release check:\n%s", out)
	}
	root, env = releaseTree(t, doneRow+"| GO-2099-0001 | not called with untrusted input (D-999) | [x] |\n", "v1.8.0", found, "3")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err != nil {
		t.Errorf("a finding carried as a ruled exception was refused: %v\n%s", err, out)
	}
	root, env = releaseTree(t, doneRow, "v1.9.9", "No vulnerabilities found.", "0")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err == nil || !strings.Contains(out, "not the pinned") {
		t.Errorf("a scanner that is not the pinned one was accepted:\n%s", out)
	}
}

// TestTheReleaseCheckNeedsAVersionTag: no tag, or a tag that is not a
// version, is a usage error.
func TestTheReleaseCheckNeedsAVersionTag(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root, env := releaseTree(t, doneRow, "v1.8.0", "", "0")
	for _, args := range [][]string{{"--release"}, {"--release", "latest"}, {"--release", "v0.2.0x"}} {
		if out, err := runGateWith(t, root, env, args...); err == nil {
			t.Errorf("%v passed:\n%s", args, out)
		}
	}
}

// TestTheTagIsTheReleaseTheLibrarySays (L-13.1): the library's User-Agent
// version must be the tag's; a release candidate leads to the same release.
func TestTheTagIsTheReleaseTheLibrarySays(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root, env := releaseTree(t, doneRow, "v1.8.0", "No vulnerabilities found.", "0")
	if out, err := runGateWith(t, root, env, "--release", "v0.3.0"); err == nil || !strings.Contains(out, "the library says it is 0.2.0") {
		t.Errorf("a tag the library does not say was accepted:\n%s", out)
	}
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0-rc.40"); err != nil {
		t.Errorf("a release candidate of the release the library says was refused: %v\n%s", err, out)
	}
	writeFile(t, root, "internal/fetch/fetch.go", "package fetch\n\nconst (\n\tVersion = \"0.2.0-dev\"\n)\n")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err == nil {
		t.Errorf("a library calling itself 0.2.0-dev was released as v0.2.0:\n%s", out)
	}
}

// TestARowIsDoneOnlyWhenItSaysSo: the release check fails closed - a row
// whose last cell is anything but "[x]" is open, a final tag waits for it.
func TestARowIsDoneOnlyWhenItSaysSo(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	for _, cell := range []string{"[~]", "", "[ ] owed", "[]", "done"} {
		root, env := releaseTree(t, doneRow+"| 3 | half done | "+cell+" |\n", "v1.8.0", "No vulnerabilities found.", "0")
		if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err == nil {
			t.Errorf("a row ending %q passed the final tag:\n%s", cell, out)
		}
	}
}
