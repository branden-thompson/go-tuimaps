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

// releaseTree is a planted, committed tree with a checklist of the given
// rows, and a directory holding a stub govulncheck that answers -version
// with version and a scan with out, exiting rc.
func releaseTree(t *testing.T, rows, version, out, rc string) (root string, env []string) {
	t.Helper()
	root, env, _ = releaseTreeWith(t, "# Release checklist\n\n| # | Check | Done |\n|---|---|---|\n"+rows, version, out, rc)
	return root, env
}

// releaseTreeWith is releaseTree with the whole checklist given. calls is
// the file the stub scanner appends one line to for every scan: the
// directory it ran in and the GOTOOLCHAIN it was given.
func releaseTreeWith(t *testing.T, checklist, version, out, rc string) (root string, env []string, calls string) {
	t.Helper()
	root = plantTree(t, true)
	writeFile(t, root, "06_docs/02_features/radar-loops/07-readiness/release-checklist.md", checklist)
	writeFile(t, root, "internal/fetch/fetch.go", "package fetch\n\nconst (\n\tVersion = \"0.2.0\"\n)\n")
	git(t, root, "init", "-q")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "planted")
	bin := t.TempDir()
	calls = filepath.Join(bin, "calls")
	writeFile(t, bin, "govulncheck", "#!/bin/sh\nif [ \"$1\" = \"-version\" ]; then echo 'Go: go1.27.1'; echo 'Scanner: govulncheck@"+version+"'; exit 0; fi\n"+
		"echo \"$PWD ${GOTOOLCHAIN:-unset}\" >> '"+calls+"'\n"+
		"printf '%s\\n' '"+out+"'\nexit "+rc+"\n")
	if err := os.Chmod(filepath.Join(bin, "govulncheck"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root, []string{"PATH=" + bin + ":" + os.Getenv("PATH")}, calls
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
	git(t, root, "commit", "-q", "-am", "dev version")
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

// TestABrokenChecklistRefusesAFinalTag: the release check counts what it
// read. A checklist with no Done table, a Done table with no rows, or a
// numbered row outside every Done table (a blank line splitting a table
// leaves the rows after it unread) refuses a final tag.
func TestABrokenChecklistRefusesAFinalTag(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	head := "# Release checklist\n\n"
	for name, c := range map[string]struct{ checklist, want string }{
		"no Done table":            {head + "| # | Check | Evidence |\n|---|---|---|\n| 0.1 | ticked | [x] |\n", "no Done table"},
		"a Done table with no row": {head + "| # | Check | Done |\n|---|---|---|\n", "no row"},
		"a blank line splits one":  {head + "| # | Check | Done |\n|---|---|---|\n| 0.1 | ticked | [x] |\n\n| 0.2 | open, after the blank line | [ ] |\n", "0.2"},
		"no table at all":          {head + "Nothing here yet.\n", "no Done table"},
	} {
		t.Run(name, func(t *testing.T) {
			root, env, _ := releaseTreeWith(t, c.checklist, "v1.8.0", "No vulnerabilities found.", "0")
			out, err := runGateWith(t, root, env, "--release", "v0.2.0")
			if err == nil || !strings.Contains(out, c.want) {
				t.Errorf("a final tag passed, or was refused without naming %q:\n%s", c.want, out)
			}
		})
	}
	whole := head + "| # | Check | Done |\n|---|---|---|\n| 0.1 | ticked | [x] |\n| 0.2 | ticked | [x] |\n\nBetween the tables.\n\n| # | Check | Evidence | Done |\n|---|---|---|---|\n| 1.1 | ticked | here | [x] |\n"
	root, env, _ := releaseTreeWith(t, whole, "v1.8.0", "No vulnerabilities found.", "0")
	out, err := runGateWith(t, root, env, "--release", "v0.2.0")
	if err != nil || !strings.Contains(out, "2 Done table(s), 3 row(s)") {
		t.Errorf("a whole checklist of two tables was refused, or not counted: %v\n%s", err, out)
	}
}

// TestTheReleaseCheckJudgesTheCommittedTree: a tree with uncommitted
// changes is refused; an existing tag must be HEAD; the run's line names the
// tag it checked.
func TestTheReleaseCheckJudgesTheCommittedTree(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root, env := releaseTree(t, doneRow, "v1.8.0", "No vulnerabilities found.", "0")
	writeFile(t, root, "06_docs/gate-runs.md", "# Gate runs\n\n| When (UTC) | Commit | Uncommitted | Mode | Result | Seconds | Overrides |\n|---|---|---|---|---|---|---|\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "the run log")
	writeFile(t, root, "lib.go", "// Package lib is planted, and changed.\npackage lib\n\n// Answer is planted.\nfunc Answer() int { return 42 }\n")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err == nil || !strings.Contains(out, "uncommitted") || !strings.Contains(out, "lib.go") {
		t.Errorf("a tree with an uncommitted change was not refused by name:\n%s", out)
	}
	git(t, root, "checkout", "-q", "lib.go")
	git(t, root, "tag", "v0.2.0")
	writeFile(t, root, "NOTES.md", "# After the tag\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "after the tag")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err == nil || !strings.Contains(out, "is not HEAD") {
		t.Errorf("a tag on another commit than HEAD was not refused:\n%s", out)
	}
	git(t, root, "tag", "-f", "v0.2.0")
	out, err := runGateWith(t, root, env, "--release", "v0.2.0")
	if err != nil || !strings.Contains(out, "release check green") {
		t.Fatalf("a clean tree at its tag was refused (the run log the gate writes is not a change): %v\n%s", err, out)
	}
	log, err := os.ReadFile(filepath.Join(root, "06_docs", "gate-runs.md"))
	if err != nil {
		t.Fatal(err)
	}
	last := strings.TrimSpace(string(log))
	if last = last[strings.LastIndex(last, "\n")+1:]; !strings.Contains(last, "release v0.2.0") {
		t.Errorf("the run's line does not name the tag it checked: %q", last)
	}
}

// TestTheReleaseScansEveryModuleAtBothToolchains (D-133): every module with
// packages is scanned, nested ones included, on this machine's toolchain
// and on the floor read from go.mod.
func TestTheReleaseScansEveryModuleAtBothToolchains(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root, env, calls := releaseTreeWith(t, "# Release checklist\n\n| # | Check | Done |\n|---|---|---|\n"+doneRow, "v1.8.0", "No vulnerabilities found.", "0")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err != nil {
		t.Fatalf("the release check failed: %v\n%s", err, out)
	}
	body, err := os.ReadFile(calls)
	if err != nil {
		t.Fatal(err)
	}
	real, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	scans := map[string]bool{}
	for _, l := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		f := strings.Fields(l)
		if len(f) != 2 {
			t.Fatalf("a scan line %q is not a directory and a toolchain", l)
		}
		dir, err := filepath.EvalSymlinks(f[0])
		if err != nil {
			t.Fatal(err)
		}
		scans[dir+" "+f[1]] = true
	}
	nested := filepath.Join(real, "tools", "t")
	for _, want := range []string{real + " local", real + " go1.25.0", nested + " local", nested + " go1.25.0"} {
		if !scans[want] {
			t.Errorf("no scan %q; the scans were:\n%s", want, body)
		}
	}
	for s := range scans {
		if strings.Contains(s, filepath.Join("cmd", "later")) {
			t.Errorf("a module with no packages was scanned: %s", s)
		}
	}
}

// TestTheTagsAreDoneAfterTheCheck: a section whose heading says it is done
// after the release check - the tags themselves, and the host pinning them -
// is listed, not held: those rows cannot be done before the tag they
// follow. Every row before it is held, and an open one still refuses.
func TestTheTagsAreDoneAfterTheCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	before := "# Release checklist\n\n## 0 · Close-out\n\n| # | Check | Done |\n|---|---|---|\n| 0.1 | ticked | [x] |\n\n"
	after := "## 3 · The tags, in order (after the release check)\n\n| # | Check | Done |\n|---|---|---|\n| 3.1 | the tag | [ ] |\n| 3.2 | the host pins it | [ ] |\n"
	root, env, _ := releaseTreeWith(t, before+after, "v1.8.0", "No vulnerabilities found.", "0")
	out, err := runGateWith(t, root, env, "--release", "v0.2.0")
	if err != nil || !strings.Contains(out, "2 row(s) after the release check") {
		t.Errorf("open rows after the release check refused the tag, or were not listed: %v\n%s", err, out)
	}
	open := strings.Replace(before, "| 0.1 | ticked | [x] |", "| 0.1 | open | [ ] |", 1)
	root, env, _ = releaseTreeWith(t, open+after, "v1.8.0", "No vulnerabilities found.", "0")
	if out, err := runGateWith(t, root, env, "--release", "v0.2.0"); err == nil || !strings.Contains(out, "0.1") {
		t.Errorf("an open row before the release check passed a final tag:\n%s", out)
	}
}
