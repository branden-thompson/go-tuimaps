package tuimaps_test

// archleg_test.go — L-6.1, D-105 (plan task L10.5): a hosted CI job hands
// the second architecture to its sibling with GATE_ARCH_LEG=sibling. Nothing
// else is accepted, and the run's log line says the leg was handed on.

import (
	"os"
	"os/exec"
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

// runLog is the planted run log's header, seven columns.
const runLog = "# Gate runs\n\n| When (UTC) | Commit | Uncommitted | Mode | Result | Seconds | Overrides |\n|---|---|---|---|---|---|---|\n"

// lastRun is the last line of a planted tree's run log.
func lastRun(t *testing.T, root string) string {
	t.Helper()
	log, err := os.ReadFile(filepath.Join(root, "06_docs", "gate-runs.md"))
	if err != nil {
		t.Fatal(err)
	}
	last := strings.TrimSpace(string(log))
	return last[strings.LastIndex(last, "\n")+1:]
}

// TestTheFuzzBudgetScalesToTheRunner (D-129): GATE_FUZZ_SCALE divides every
// fuzz count, the leg says so, and the run's line carries the setting, so M6
// does not count a scaled run. Anything but a whole number above zero is
// refused.
func TestTheFuzzBudgetScalesToTheRunner(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	for _, bad := range []string{"0", "half", "-2", "1.5"} {
		if out, err := runGateWith(t, plantTree(t, true), []string{"GATE_FUZZ_SCALE=" + bad}, "--fuzz"); err == nil || !strings.Contains(out, "GATE_FUZZ_SCALE") {
			t.Errorf("GATE_FUZZ_SCALE=%s was accepted:\n%s", bad, out)
		}
	}
	root := plantTree(t, true)
	writeFile(t, root, "quick_test.go", "package lib\n\nimport \"testing\"\n\nfunc FuzzQuick(f *testing.F) {\n\tf.Add(1)\n\tf.Fuzz(func(t *testing.T, n int) {})\n}\n")
	writeFile(t, root, "06_docs/gate-runs.md", runLog)
	out, err := runGateWith(t, root, []string{"GATE_FUZZ_SCALE=5000"}, "--fuzz")
	if err != nil {
		t.Fatalf("the scaled fuzz run failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "fuzz FuzzQuick, 1000x") || !strings.Contains(out, "GATE_FUZZ_SCALE") {
		t.Errorf("the leg does not run the default 5000000x divided by 5000, or does not say it was scaled:\n%s", out)
	}
	if last := lastRun(t, root); !strings.Contains(last, "GATE_FUZZ_SCALE=5000") {
		t.Errorf("the run's line does not carry the scale: %q", last)
	}
}

// TestAFailedRunNamesItsLegs: the FAILED line in the run log names the legs
// that failed, and the module of each, so a failure states its cause.
func TestAFailedRunNamesItsLegs(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, false)
	writeFile(t, root, "06_docs/gate-runs.md", runLog)
	if out, err := runGate(t, root); err == nil {
		t.Fatalf("the gate passed a failing module:\n%s", out)
	}
	last := lastRun(t, root)
	if !strings.Contains(last, "FAILED") || !strings.Contains(last, "tests, race detector on") || !strings.Contains(last, "tools/t") {
		t.Errorf("the FAILED line does not name the failing leg and its module: %q", last)
	}
	if strings.Count(last, "|") != 8 {
		t.Errorf("the FAILED line is not seven cells: %q", last)
	}
}

// TestTheP10Leg (D-136): where the harness is set up (a2dh on PATH and the
// tree's .a2dh.yml), `a2dh p10 check` runs and a live finding fails the
// gate; elsewhere the leg is named among NOT RUN.
func TestTheP10Leg(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	stub := func(rc string) []string {
		bin := t.TempDir()
		writeFile(t, bin, "a2dh", "#!/bin/sh\n[ \"$1 $2\" = \"p10 check\" ] || exit 9\necho 'planted: 1 P10 finding(s) in changed code'\nexit "+rc+"\n")
		if err := os.Chmod(filepath.Join(bin, "a2dh"), 0o755); err != nil {
			t.Fatal(err)
		}
		return []string{"PATH=" + bin + ":" + os.Getenv("PATH")}
	}
	root := plantTree(t, true)
	writeFile(t, root, ".a2dh.yml", "planted: true\n")
	if out, err := runGateWith(t, root, stub("1"), "--quick"); err == nil || !strings.Contains(out, "P10") || !strings.Contains(out, "planted: 1 P10 finding(s)") {
		t.Errorf("a live P10 finding did not fail the gate and show the report:\n%s", out)
	}
	if out, err := runGateWith(t, root, stub("0"), "--quick"); err != nil || !strings.Contains(out, "ok      P10") {
		t.Errorf("a clean P10 check did not pass as a leg: %v\n%s", err, out)
	}
	bare := plantTree(t, true)
	out, err := runGateWith(t, bare, stub("1"), "--quick")
	if err != nil {
		t.Fatalf("a tree without the harness failed: %v\n%s", err, out)
	}
	if notRun := out[strings.LastIndex(out, "NOT RUN"):]; !strings.Contains(notRun, "P10") {
		t.Errorf("the P10 check is not named among NOT RUN where the harness is not set up:\n%s", out)
	}
}

// TestALicenceIsNeededForEveryModuleBuiltIn: a module whose code a module of
// the tree compiles needs a licence file, and the check reads the modules
// from the packages' own dependencies, each with its directory.
func TestALicenceIsNeededForEveryModuleBuiltIn(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	dep := t.TempDir()
	writeFile(t, dep, "go.mod", "module example.com/dep\n\ngo "+goFloor+"\n")
	writeFile(t, dep, "dep.go", "// Package dep is planted.\npackage dep\n\n// Six is planted.\nfunc Six() int { return 6 }\n")
	writeFile(t, root, "tools/t/go.mod", "module example.com/lib/tools/t\n\ngo "+goFloor+"\n\nrequire (\n\texample.com/lib v0.0.0\n\texample.com/dep v0.0.0\n)\n\nreplace example.com/dep => "+dep+"\n")
	writeFile(t, root, "tools/t/t.go", "// Package t is planted.\npackage t\n\nimport \"example.com/dep\"\n\n// Seven is planted.\nfunc Seven() int { return dep.Six() + 1 }\n")
	out, err := runGate(t, root)
	if err == nil || !strings.Contains(out, "no licence file in "+dep) {
		t.Fatalf("a compiled-in module with no licence file passed:\n%s", out)
	}
	writeFile(t, dep, "LICENSE", "Planted.\n")
	if out, err := runGate(t, root); err != nil {
		t.Errorf("a compiled-in module with a licence file failed: %v\n%s", err, out)
	}
}

// TestNotRunNamesTheOtherArchitecture: the closing NOT RUN line names the
// real hardware this machine is not, never its own.
func TestNotRunNamesTheOtherArchitecture(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	out, err := runGate(t, plantTree(t, true))
	if err != nil {
		t.Fatalf("the gate failed: %v\n%s", err, out)
	}
	host, err := exec.Command("uname", "-m").Output()
	if err != nil {
		t.Fatal(err)
	}
	own, other := "arm64", "amd64"
	if h := strings.TrimSpace(string(host)); h == "x86_64" || h == "amd64" {
		own, other = "amd64", "arm64"
	}
	notRun := out[strings.LastIndex(out, "NOT RUN on this machine"):]
	if strings.Contains(notRun, "real "+own+" hardware") || !strings.Contains(notRun, "real "+other+" hardware") {
		t.Errorf("on %s the NOT RUN line must name real %s hardware, not its own:\n%s", own, other, notRun)
	}
}

// workflowStep is the text of the hosted workflow's step with the given
// name, up to the next step or job.
func workflowStep(t *testing.T, workflow, name string) string {
	t.Helper()
	at := strings.Index(workflow, "- name: "+name+"\n")
	if at < 0 {
		t.Fatalf("the workflow has no step %q", name)
	}
	step := workflow[at+len("- name: "+name+"\n"):]
	for _, end := range []string{"\n      - ", "\n  both-architectures:"} {
		if i := strings.Index(step, end); i >= 0 {
			step = step[:i]
		}
	}
	return step
}

// TestTheHostedFullGateFitsTheRunner (D-129): the hosted full gate scales
// its fuzz budget and hands the second architecture to its sibling; the
// quick gate needs neither, and no step re-exports PATH; the workflow says
// its schedule fires only from the default branch, and keeps dispatch.
func TestTheHostedFullGateFitsTheRunner(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(".github", "workflows", "gate.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(body)
	full := workflowStep(t, workflow, "full gate")
	for _, want := range []string{"GATE_FUZZ_SCALE:", "GATE_ARCH_LEG: sibling", "run: scripts/gate\n"} {
		if !strings.Contains(full+"\n", want) {
			t.Errorf("the full gate step does not carry %q:\n%s", want, full)
		}
	}
	quick := workflowStep(t, workflow, "quick gate")
	if strings.Contains(quick, "GATE_") {
		t.Errorf("the quick gate step sets a gate setting it does not use:\n%s", quick)
	}
	if strings.Contains(workflow, "PATH=") {
		t.Error("a step re-exports PATH; the scanner's directory is added once, through GITHUB_PATH")
	}
	if !strings.Contains(workflow, "GITHUB_PATH") {
		t.Error("no step puts the installed scanner's directory on PATH")
	}
	if !strings.Contains(workflow, "default branch") || !strings.Contains(workflow, "workflow_dispatch:") {
		t.Error("the workflow must say its schedule fires only from the default branch, and keep workflow_dispatch")
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
