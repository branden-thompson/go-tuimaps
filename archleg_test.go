package tuimaps_test

// archleg_test.go — L-6.1, D-105 (plan task L10.5): a hosted CI job hands
// the second architecture to its sibling with GATE_ARCH_LEG=sibling. Nothing
// else is accepted, and the run's log line says the leg was handed on.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTheArchitectureLegIsHandedOnOnlyByName: only "sibling" is accepted,
// and a run that used it says so on its line.
func TestTheArchitectureLegIsHandedOnOnlyByName(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	if out, err := runGateWith(t, root, []string{"GATE_ARCH_LEG=skip"}, "--quick"); err == nil || !strings.Contains(out, "may only be 'sibling'") {
		t.Errorf("GATE_ARCH_LEG=skip was accepted:\n%s", out)
	}
	writeFile(t, root, "06_docs/gate-runs.md", "# Gate runs\n\n| When (UTC) | Commit | Uncommitted | Mode | Result | Seconds | Overrides |\n|---|---|---|---|---|---|---|\n")
	if out, err := runGateWith(t, root, []string{"GATE_ARCH_LEG=sibling"}, "--quick"); err != nil {
		t.Fatalf("the quick gate with the leg handed on failed: %v\n%s", err, out)
	}
	log, err := os.ReadFile(filepath.Join(root, "06_docs", "gate-runs.md"))
	if err != nil {
		t.Fatal(err)
	}
	if last := strings.TrimSpace(string(log)); !strings.Contains(last[strings.LastIndex(last, "\n")+1:], "GATE_ARCH_LEG=sibling") {
		t.Errorf("the run's line does not say the leg was handed on:\n%s", log)
	}
}

// TestTheSoakIsAModeOfItsOwn: the hour is never folded into another mode.
func TestTheSoakIsAModeOfItsOwn(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	if out, err := runGateWith(t, root, nil, "--soak", "--quick"); err == nil || !strings.Contains(out, "one mode at most") {
		t.Errorf("--soak with --quick was accepted:\n%s", out)
	}
}
