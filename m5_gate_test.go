package tuimaps_test

// m5_gate_test.go — plan task P7.1 (watchpost D-350): the gate's --cost mode
// fails when it measures nothing, and the release check holds a final v0.3
// tag to M5's record.

import (
	"strings"
	"testing"
)

// m5Tree is a planted, committed v0.3.0 tree whose checklist is ticked, with
// m5Test, when given, as its root module's record check.
func m5Tree(t *testing.T, m5Test string) (root string, env []string) {
	t.Helper()
	root, env = releaseTree(t, doneRow, "v1.8.0", "No vulnerabilities found.", "0")
	writeFile(t, root, "internal/fetch/fetch.go", "package fetch\n\nconst (\n\tVersion = \"0.3.0\"\n)\n")
	writeFile(t, root, "06_docs/02_features/propagation-fields/07-readiness/release-checklist.md", "# Release checklist\n\n| # | Check | Done |\n|---|---|---|\n"+doneRow)
	if m5Test != "" {
		writeFile(t, root, "m5_test.go", "package lib\n\nimport (\n\t\"os\"\n\t\"testing\"\n)\n\nfunc TestTheM5RecordHoldsForThisCode(t *testing.T) {\n\t"+m5Test+"\n}\n")
	}
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "v0.3.0")
	return root, env
}

// TestAFinalV03TagWaitsForM5sRecord: a final v0.3 tag is refused when M5's
// record check is not there to run or refuses the record, and passes when it
// holds, given the code's tree; a release candidate is not held to it.
func TestAFinalV03TagWaitsForM5sRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root, env := m5Tree(t, "")
	if out, err := runGateWith(t, root, env, "--release", "v0.3.0"); err == nil || !strings.Contains(out, "M5's record was not checked") {
		t.Errorf("a final v0.3 tag passed with no record check run:\n%s", out)
	}
	refusing := "t.Fatal(\"planted refusal\", os.Getenv(\"TUIMAPS_TREE\"))"
	root, env = m5Tree(t, refusing)
	if out, err := runGateWith(t, root, env, "--release", "v0.3.0"); err == nil || !strings.Contains(out, "M5's record does not hold") || !strings.Contains(out, "planted refusal") {
		t.Errorf("a final v0.3 tag passed a record the check refuses:\n%s", out)
	}
	if out, err := runGateWith(t, root, env, "--release", "v0.3.0-rc.1"); err != nil {
		t.Errorf("a release candidate was held to M5's record:\n%s", out)
	}
	holding := "if os.Getenv(\"TUIMAPS_RELEASE\") != \"1\" || len(os.Getenv(\"TUIMAPS_TREE\")) != 12 {\n\t\tt.Fatal(\"the release check named no code\")\n\t}"
	root, env = m5Tree(t, holding)
	if out, err := runGateWith(t, root, env, "--release", "v0.3.0"); err != nil || !strings.Contains(out, "M5's record holds") {
		t.Errorf("a final v0.3 tag was refused with a record that holds:\n%s", out)
	}
}

// TestACostRunThatMeasuresNothingFails: --cost over a tree without M5's
// measure fails and says so, never green with nothing timed.
func TestACostRunThatMeasuresNothingFails(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root, env := m5Tree(t, "")
	if out, err := runGateWith(t, root, env, "--cost"); err == nil || !strings.Contains(out, "cost FAILED") {
		t.Errorf("a cost run that measured nothing did not fail:\n%s", out)
	}
	if out, err := runGateWith(t, root, env, "--cost", "--quick"); err == nil || !strings.Contains(out, "one mode at most") {
		t.Errorf("--cost with another mode was not refused:\n%s", out)
	}
}
