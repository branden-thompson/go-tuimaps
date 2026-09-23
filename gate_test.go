package tuimaps_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// plantTree writes a small repository: a root module and one nested module,
// each with one test. nestedPasses decides the nested module's test.
func plantTree(t *testing.T, nestedPasses bool) string {
	t.Helper()
	root := t.TempDir()
	// The nested module requires the root at a version that was never
	// published, and imports it: what every separate module of this
	// repository does before a remote exists.
	body := "if lib.Answer() != 42 {\n\t\tt.Fatal(\"the nested module did not reach the root module\")\n\t}"
	if !nestedPasses {
		body = "t.Fatal(\"planted failure\", lib.Answer())"
	}
	files := map[string]string{
		"go.mod":               "module example.com/lib\n\ngo 1.25.0\n",
		"LICENSE":              "Planted licence.\n",
		"lib.go":               "// Package lib is planted.\npackage lib\n\n// Answer is planted.\nfunc Answer() int { return 42 }\n",
		"lib_test.go":          "package lib\n\nimport \"testing\"\n\nfunc TestRoot(t *testing.T) { t.Log(\"ok\") }\n",
		"tools/t/go.mod":       "module example.com/lib/tools/t\n\ngo 1.25.0\n\nrequire example.com/lib v0.0.0\n",
		"tools/t/t.go":         "// Package t is planted.\npackage t\n",
		"tools/t/t_test.go":    "package t\n\nimport (\n\t\"testing\"\n\n\t\"example.com/lib\"\n)\n\nfunc TestNested(t *testing.T) {\n\t" + body + "\n}\n",
		"cmd/later/go.mod":     "module example.com/lib/cmd/later\n\ngo 1.25.0\n",
		"testdata/x/go.mod":    "module example.com/ignored\n\ngo 1.25.0\n",
		"testdata/x/x_test.go": "package x\n\nimport \"testing\"\n\nfunc TestIgnored(t *testing.T) { t.Fatal(\"test data is not a module of the repository\") }\n",
	}
	for rel, src := range files {
		writeFile(t, root, rel, src)
	}
	return root
}

// runGate runs the gate's quick legs over a planted tree.
func runGate(t *testing.T, root string) (string, error) {
	t.Helper()
	return runGateWith(t, root, nil, "--quick")
}

// TestGateIsGreenOnAPassingTree is the "green run" of plan task 00.7.
func TestGateIsGreenOnAPassingTree(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	out, err := runGate(t, root)
	if err != nil {
		t.Fatalf("the gate failed on a passing tree: %v\n%s", err, out)
	}
	for _, m := range []string{"example.com/lib", "example.com/lib/tools/t"} {
		if !strings.Contains(out, m) {
			t.Errorf("the gate's report does not name module %s:\n%s", m, out)
		}
	}
	if !strings.Contains(out, "example.com/lib/cmd/later") || !strings.Contains(out, "no Go packages yet") {
		t.Errorf("a module with no packages yet must be named and said to be empty, not failed and not hidden:\n%s", out)
	}
	if strings.Contains(out, "example.com/ignored") {
		t.Errorf("the gate ran a module under testdata:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(root, "go.work")); err == nil {
		t.Error("the gate left its workspace file behind")
	}
}

// TestGateFailsWhenOneModuleFails is plan task 00.12, first half.
func TestGateFailsWhenOneModuleFails(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	out, err := runGate(t, plantTree(t, false))
	if err == nil {
		t.Fatalf("the gate passed although a nested module's test fails:\n%s", out)
	}
	if !strings.Contains(out, "example.com/lib/tools/t") || !strings.Contains(out, "FAILED") {
		t.Errorf("the gate's report does not say which module failed:\n%s", out)
	}
}

// TestGateFailsWhenTheWorkspaceFileIsTracked is plan task 00.12, second half.
func TestGateFailsWhenTheWorkspaceFileIsTracked(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	if err := os.WriteFile(filepath.Join(root, "go.work"), []byte("go 1.25.0\n\nuse .\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "go.work"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	out, err := runGate(t, root)
	if err == nil {
		t.Fatalf("the gate passed although go.work is tracked:\n%s", out)
	}
	if !strings.Contains(out, "go.work") {
		t.Errorf("the gate's report does not name go.work:\n%s", out)
	}
}

// runDocsLane runs the gate's docs lane over a planted tree (v0.2.0 D-15).
func runDocsLane(t *testing.T, root string) (string, error) {
	t.Helper()
	return runGateWith(t, root, nil, "--docs")
}

// plantCommitted plants a tree whose root module has a test that reads a
// Markdown file, as this repository's own tests do, and commits it: the docs
// lane decides from what has changed since the last commit.
func plantCommitted(t *testing.T) string {
	t.Helper()
	root := plantTree(t, true)
	files := map[string]string{
		"NOTES.md":      "# Notes\n\nFine.\n",
		"notes_test.go": "package lib\n\nimport (\n\t\"os\"\n\t\"strings\"\n\t\"testing\"\n)\n\nfunc TestNotes(t *testing.T) {\n\tb, err := os.ReadFile(\"NOTES.md\")\n\tif err != nil || strings.Contains(string(b), \"BROKEN\") {\n\t\tt.Fatal(\"the notes are broken\", err)\n\t}\n}\n",
	}
	for rel, src := range files {
		writeFile(t, root, rel, src)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "user.name=gate", "-c", "user.email=gate@example.com", "commit", "-q", "-m", "planted"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return root
}

// git runs one git command in a planted tree and fails the test if it fails.
func git(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=gate", "-c", "user.email=gate@example.com"}, args...)...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, root, rel, src string) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestDocsLaneRunsTheTestsThatReadDocuments: a change of Markdown alone is
// checked by every module's tests, which is where documents are read - so a
// document a test rejects still turns the lane red.
func TestDocsLaneRunsTheTestsThatReadDocuments(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantCommitted(t)
	writeFile(t, root, "NOTES.md", "# Notes\n\nStill fine, and longer.\n")
	writeFile(t, root, "docs/new.md", "# A new page\n")
	git(t, root, "add", "-A")
	out, err := runDocsLane(t, root)
	if err != nil {
		t.Fatalf("the docs lane failed on a Markdown-only change the tests accept: %v\n%s", err, out)
	}
	if !strings.Contains(out, "docs lane") || !strings.Contains(out, "example.com/lib/tools/t") {
		t.Errorf("the docs lane must say it is the docs lane and name every module it ran:\n%s", out)
	}
	writeFile(t, root, "NOTES.md", "# Notes\n\nBROKEN\n")
	git(t, root, "add", "-A")
	out, err = runDocsLane(t, root)
	if err == nil {
		t.Fatalf("the docs lane passed a Markdown change a test rejects:\n%s", out)
	}
}

// TestDocsLaneRefusesAnyOtherFile: one file that is not Markdown - code, a
// test, a script, a specimen - anywhere in the change means the full gate.
func TestDocsLaneRefusesAnyOtherFile(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	for name, other := range map[string]string{
		"a changed source file":  "lib.go",
		"a new staged file":      "specimens/frame.txt",
		"a deleted tracked file": "",
	} {
		t.Run(name, func(t *testing.T) {
			root := plantCommitted(t)
			writeFile(t, root, "NOTES.md", "# Notes\n\nFine again.\n")
			want := other
			switch other {
			case "":
				want = "lib_test.go"
				if err := os.Remove(filepath.Join(root, want)); err != nil {
					t.Fatal(err)
				}
			case "lib.go":
				writeFile(t, root, other, "// Package lib is planted.\npackage lib\n\n// Answer is planted.\nfunc Answer() int { return 43 }\n")
			default:
				writeFile(t, root, other, "a frame\n")
			}
			git(t, root, "add", "-A")
			out, err := runDocsLane(t, root)
			if err == nil {
				t.Fatalf("the docs lane accepted a change with %s in it:\n%s", want, out)
			}
			if !strings.Contains(out, want) || !strings.Contains(out, "scripts/gate") {
				t.Errorf("the refusal must name %s and send the change to the full gate:\n%s", want, out)
			}
		})
	}
}

// runGateWith runs the gate with arguments and extra environment over a
// planted tree.
func runGateWith(t *testing.T, root string, env []string, args ...string) (string, error) {
	t.Helper()
	gate, err := filepath.Abs(filepath.Join("scripts", "gate"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(gate, args...)
	// The gate must not need the network to find the root module.
	cmd.Env = append(append(os.Environ(), "GATE_ROOT="+root, "GOPROXY=off"), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// TestDocsLaneWithNothingToCheckFails: a lane run after the commit, or on a
// clean tree, checked nothing, and must not read as a pass (v0.2.0 D-31).
func TestDocsLaneWithNothingToCheckFails(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	out, err := runDocsLane(t, plantCommitted(t))
	if err == nil {
		t.Fatalf("the docs lane passed with nothing to check:\n%s", out)
	}
	if !strings.Contains(out, "nothing to check") {
		t.Errorf("the refusal must say there is nothing to check:\n%s", out)
	}
}

// TestGateFailsWhenAModuleCannotBeLoaded: a module whose packages cannot be
// listed is broken, not empty (v0.2.0 D-31).
func TestGateFailsWhenAModuleCannotBeLoaded(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	writeFile(t, root, "tools/t/broken.go", "this is not Go\n")
	out, err := runGate(t, root)
	if err == nil {
		t.Fatalf("the gate passed a module that cannot be loaded:\n%s", out)
	}
	section := out[strings.Index(out, "module example.com/lib/tools/t"):]
	if !strings.Contains(section, "cannot be listed") || strings.Contains(section, "no Go packages yet") {
		t.Errorf("a module that cannot be loaded must be named as failed, not called empty:\n%s", out)
	}
}

// TestAStalledFuzzLegFails: a fuzz leg that stops making progress fails when
// its wall-clock limit runs out, instead of holding the gate forever
// (v0.2.0 D-31).
func TestAStalledFuzzLegFails(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	writeFile(t, root, "stall_test.go", "package lib\n\nimport (\n\t\"testing\"\n\t\"time\"\n)\n\nfunc FuzzStall(f *testing.F) {\n\tf.Add(1)\n\tf.Fuzz(func(t *testing.T, n int) { time.Sleep(time.Hour) })\n}\n")
	started := time.Now()
	out, err := runGateWith(t, root, []string{"GATE_FUZZ_LIMIT=5"}, "--fuzz")
	if err == nil {
		t.Fatalf("the gate passed a fuzz leg that never finished:\n%s", out)
	}
	if !strings.Contains(out, "TIME LIMIT") || !strings.Contains(out, "FuzzStall") {
		t.Errorf("the failure must name the leg and say it stalled:\n%s", out)
	}
	if took := time.Since(started); took > 90*time.Second {
		t.Errorf("the limit did not hold: the gate took %v", took)
	}
}

// TestEveryRunIsLogged: a run leaves a line - commit, mode, result - in the
// tracked run log, which M6 counts from (v0.2.0 D-31).
func TestEveryRunIsLogged(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantCommitted(t)
	writeFile(t, root, "06_docs/gate-runs.md", "# Gate runs\n\n| When (UTC) | Commit | Uncommitted | Mode | Result | Seconds |\n|---|---|---|---|---|---|\n")
	writeFile(t, root, "NOTES.md", "# Notes\n\nFine, and logged.\n")
	git(t, root, "add", "NOTES.md")
	tree := git(t, root, "write-tree")[:12]
	out, err := runGateWith(t, root, []string{"GATE_FUZZ_LIMIT=77"}, "--docs")
	if err != nil {
		t.Fatalf("the docs lane failed: %v\n%s", err, out)
	}
	log, err := os.ReadFile(filepath.Join(root, "06_docs", "gate-runs.md"))
	if err != nil {
		t.Fatal(err)
	}
	head := exec.Command("git", "rev-parse", "--short", "HEAD")
	head.Dir = root
	sha, err := head.Output()
	if err != nil {
		t.Fatal(err)
	}
	last := strings.TrimSpace(string(log))
	last = last[strings.LastIndex(last, "\n")+1:]
	for _, want := range []string{strings.TrimSpace(string(sha)), tree, "docs", "green", "GATE_FUZZ_LIMIT=77"} {
		if !strings.Contains(last, want) {
			t.Errorf("the run's line does not carry %q: %q", want, last)
		}
	}
	// A run that fails is logged as failed: M6 counts the Result column.
	writeFile(t, root, "NOTES.md", "# Notes\n\nBROKEN\n")
	git(t, root, "add", "NOTES.md")
	if out, err := runDocsLane(t, root); err == nil {
		t.Fatalf("the docs lane passed a note a test rejects:\n%s", out)
	}
	log, err = os.ReadFile(filepath.Join(root, "06_docs", "gate-runs.md"))
	if err != nil {
		t.Fatal(err)
	}
	last = strings.TrimSpace(string(log))
	last = last[strings.LastIndex(last, "\n")+1:]
	if !strings.Contains(last, "FAILED") || strings.Contains(last, "green") {
		t.Errorf("a failed run must be logged as FAILED: %q", last)
	}
}

// TestFuzzCountTableMatchesTheTargets: every row of the gate's per-target
// count table names a fuzz target that exists, and every target has a row, so
// a rename cannot leave a stale row and an uncalibrated target (v0.2.0 D-38).
func TestFuzzCountTableMatchesTheTargets(t *testing.T) {
	script, err := os.ReadFile(filepath.Join("scripts", "gate"))
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^\s+(Fuzz\w+)\)\s+echo "\d+x"`).FindAllStringSubmatch(string(script), -1) {
		rows[m[1]] = true
	}
	if len(rows) < 10 {
		t.Fatalf("found %d rows in the count table; the table's shape has changed and this test has lost its subject", len(rows))
	}
	targets := map[string]bool{}
	fuzzFunc := regexp.MustCompile(`(?m)^func (Fuzz\w+)\(`)
	err = filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != "." && (d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".") || d.Name() == "_a2dh") {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range fuzzFunc.FindAllStringSubmatch(string(body), -1) {
			targets[m[1]] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for name := range rows {
		if !targets[name] {
			t.Errorf("the count table has a row for %s, and no such fuzz target exists", name)
		}
	}
	for name := range targets {
		if !rows[name] {
			t.Errorf("fuzz target %s has no row in the count table", name)
		}
	}
}

// TestNotRunNamesTheLastTag: the gate's closing NOT RUN line names the last
// tag rather than claiming none exists (v0.2.0 D-38).
func TestNotRunNamesTheLastTag(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	// Releases are squash-merged (D-1), so the last release's tag is never an
	// ancestor of a work branch: tag a commit on a branch of its own.
	root := plantCommitted(t)
	work := git(t, root, "rev-parse", "--abbrev-ref", "HEAD")
	git(t, root, "checkout", "-q", "-b", "release")
	writeFile(t, root, "RELEASE.md", "# Released\n")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "squashed release")
	git(t, root, "tag", "v9.9.9")
	git(t, root, "checkout", "-q", work)
	out, err := runGate(t, root)
	if err != nil {
		t.Fatalf("the gate failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "v9.9.9") || strings.Contains(out, "none exists yet") {
		t.Errorf("the NOT RUN line must name the last tag:\n%s", out)
	}
}

// TestDocsLaneJudgesOnlyWhatIsStaged: the lane checks what the commit will
// record; an unstaged code change and an untracked file are not part of it
// (v0.2.0 D-41).
func TestDocsLaneJudgesOnlyWhatIsStaged(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantCommitted(t)
	writeFile(t, root, "NOTES.md", "# Notes\n\nStaged.\n")
	git(t, root, "add", "NOTES.md")
	writeFile(t, root, "lib.go", "// Package lib is planted, and edited without being staged.\npackage lib\n\n// Answer is planted.\nfunc Answer() int { return 42 }\n")
	writeFile(t, root, "scratch.txt", "not part of the commit\n")
	out, err := runDocsLane(t, root)
	if err != nil {
		t.Fatalf("the lane refused a staged Markdown change because of files that are not staged: %v\n%s", err, out)
	}
	if !strings.Contains(out, "not staged") {
		t.Errorf("the lane must say how many files it left out as not staged:\n%s", out)
	}
}

// TestDocsLaneFailsWhenGitCannotRead: a change the lane cannot read is not a
// change it may pass (v0.2.0 D-31, D-40).
func TestDocsLaneFailsWhenGitCannotRead(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantCommitted(t)
	if err := os.WriteFile(filepath.Join(root, ".git", "index"), []byte("not an index"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runDocsLane(t, root)
	if err == nil {
		t.Fatalf("the lane passed a change git could not read:\n%s", out)
	}
	if !strings.Contains(out, "could not read") {
		t.Errorf("the refusal must say the change could not be read:\n%s", out)
	}
}

// TestGateRefusesTwoModes: two mode flags together skipped every leg and
// printed green (v0.2.0 D-40).
func TestGateRefusesTwoModes(t *testing.T) {
	for _, args := range [][]string{{"--quick", "--fuzz"}, {"--docs", "--fuzz"}, {"--docs", "--quick"}} {
		out, err := runGateWith(t, plantTree(t, true), nil, args...)
		if err == nil || !strings.Contains(out, "one mode") {
			t.Errorf("%v must be refused as more than one mode:\n%s", args, out)
		}
	}
}

// TestFuzzModeFailsWithNothingToRun: a fuzz run that ran no fuzz leg checked
// nothing, and must not read as green (v0.2.0 D-40).
func TestFuzzModeFailsWithNothingToRun(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	out, err := runGateWith(t, plantTree(t, true), nil, "--fuzz")
	if err == nil || !strings.Contains(out, "no fuzz leg") {
		t.Fatalf("a fuzz run with no fuzz target must fail and say so:\n%s", out)
	}
}

// TestFuzzModeFailsWhenTestsCannotBeListed: a package whose tests do not
// compile has fuzz targets nobody can list, and that is a failure, not an
// absence (v0.2.0 D-40).
func TestFuzzModeFailsWhenTestsCannotBeListed(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	writeFile(t, root, "bad_test.go", "package lib\n\nimport \"testing\"\n\nfunc FuzzBad(f *testing.F) { undefinedHere() }\n")
	out, err := runGateWith(t, root, nil, "--fuzz")
	if err == nil || !strings.Contains(out, "cannot be listed") {
		t.Fatalf("a package whose tests do not compile must fail the fuzz run:\n%s", out)
	}
}

// TestLicenceCheckFailsWhenTheGraphCannotBeListed: a module graph the licence
// check cannot list was a loop over nothing, and passed (v0.2.0 D-40).
func TestLicenceCheckFailsWhenTheGraphCannotBeListed(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	writeFile(t, root, "tools/t/go.mod", "module example.com/lib/tools/t\n\ngo 1.25.0\n\nrequire (\n\texample.com/lib v0.0.0\n\texample.com/nowhere v1.2.3\n)\n")
	out, err := runGate(t, root)
	if err == nil || !strings.Contains(out, "licence") {
		t.Fatalf("a module graph that cannot be listed must fail the licence check:\n%s", out)
	}
}

// TestALegThatIgnoresTermIsStillStopped: a process that ignores TERM is
// killed when the limit runs out, so the limit is a limit (v0.2.0 D-40).
func TestALegThatIgnoresTermIsStillStopped(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	writeFile(t, root, "stubborn_test.go", "package lib\n\nimport (\n\t\"os/exec\"\n\t\"testing\"\n)\n\nfunc FuzzStubborn(f *testing.F) {\n\tf.Add(1)\n\tf.Fuzz(func(t *testing.T, n int) { _ = exec.Command(\"sh\", \"-c\", \"trap '' TERM; exec sleep 3607\").Run() })\n}\n")
	started := time.Now()
	out, err := runGateWith(t, root, []string{"GATE_FUZZ_LIMIT=3"}, "--fuzz")
	if err == nil || !strings.Contains(out, "TIME LIMIT") {
		t.Fatalf("a leg that ignores TERM must still fail at its limit:\n%s", out)
	}
	if took := time.Since(started); took > 60*time.Second {
		t.Errorf("the limit did not hold against a process that ignores TERM: %v", took)
	}
	// Nothing of the leg may outlive it: a survivor would keep running, and
	// would hold the leg's output open if it wrote to it.
	time.Sleep(time.Second)
	if left, _ := exec.Command("pgrep", "-f", "sleep 3607").Output(); len(strings.TrimSpace(string(left))) > 0 {
		_ = exec.Command("pkill", "-9", "-f", "sleep 3607").Run()
		t.Errorf("a process that ignored TERM outlived the leg: pid %s", strings.TrimSpace(string(left)))
	}
}

// TestGateRefusesABadFuzzLimit: the limit is a number of seconds, or the gate
// does not start (v0.2.0 D-40).
func TestGateRefusesABadFuzzLimit(t *testing.T) {
	out, err := runGateWith(t, plantTree(t, true), []string{"GATE_FUZZ_LIMIT=soon"}, "--fuzz")
	if err == nil || !strings.Contains(out, "GATE_FUZZ_LIMIT") {
		t.Fatalf("a limit that is not a whole number of seconds must be refused:\n%s", out)
	}
}

// TestGateFailsOnAnUnformattedFile: a Go file gofmt would change fails the
// gate; one sat unformatted through several green runs before this leg
// existed (v0.2.0 D-46).
func TestGateFailsOnAnUnformattedFile(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantTree(t, true)
	writeFile(t, root, "lib.go", "// Package lib is planted.\npackage lib\n\n// Answer is planted.\nfunc Answer() int {    return 42 }\n")
	out, err := runGate(t, root)
	if err == nil || !strings.Contains(out, "lib.go") || !strings.Contains(out, "gofmt") {
		t.Fatalf("an unformatted file must fail the gate and be named:\n%s", out)
	}
}

// TestDocsLaneAcceptsTheGeneratedAtlas: a diagram edit regenerates the atlas
// page, which is not Markdown; the lane takes that one generated file, whose
// match with the documents the atlas's own test checks, and refuses every
// other non-Markdown file (v0.2.0 D-53).
func TestDocsLaneAcceptsTheGeneratedAtlas(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the gate; skipped with -short")
	}
	root := plantCommitted(t)
	writeFile(t, root, "NOTES.md", "# Notes\n\nA diagram was edited.\n")
	writeFile(t, root, "06_docs/architecture-atlas.html", "<html>generated</html>\n")
	git(t, root, "add", "-A")
	if out, err := runDocsLane(t, root); err != nil {
		t.Fatalf("the lane refused the generated atlas page: %v\n%s", err, out)
	}
	writeFile(t, root, "06_docs/other.html", "<html>hand-made</html>\n")
	git(t, root, "add", "-A")
	out, err := runDocsLane(t, root)
	if err == nil || !strings.Contains(out, "06_docs/other.html") {
		t.Fatalf("the lane must still refuse any other non-Markdown file:\n%s", out)
	}
}
