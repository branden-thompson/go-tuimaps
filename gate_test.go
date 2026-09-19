package tuimaps_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// plantTree writes a small repository: a root module and one nested module,
// each with one test. nestedPasses decides the nested module's test.
func plantTree(t *testing.T, nestedPasses bool) string {
	t.Helper()
	root := t.TempDir()
	body := "t.Log(\"ok\")"
	if !nestedPasses {
		body = "t.Fatal(\"planted failure\")"
	}
	files := map[string]string{
		"go.mod":               "module example.com/lib\n\ngo 1.25.0\n",
		"lib.go":               "// Package lib is planted.\npackage lib\n",
		"lib_test.go":          "package lib\n\nimport \"testing\"\n\nfunc TestRoot(t *testing.T) { t.Log(\"ok\") }\n",
		"tools/t/go.mod":       "module example.com/lib/tools/t\n\ngo 1.25.0\n",
		"tools/t/t.go":         "// Package t is planted.\npackage t\n",
		"tools/t/t_test.go":    "package t\n\nimport \"testing\"\n\nfunc TestNested(t *testing.T) { " + body + " }\n",
		"cmd/later/go.mod":     "module example.com/lib/cmd/later\n\ngo 1.25.0\n",
		"testdata/x/go.mod":    "module example.com/ignored\n\ngo 1.25.0\n",
		"testdata/x/x_test.go": "package x\n\nimport \"testing\"\n\nfunc TestIgnored(t *testing.T) { t.Fatal(\"test data is not a module of the repository\") }\n",
	}
	for rel, src := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// runGate runs the gate's quick legs over a planted tree.
func runGate(t *testing.T, root string) (string, error) {
	t.Helper()
	gate, err := filepath.Abs(filepath.Join("scripts", "gate"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(gate, "--quick")
	cmd.Env = append(os.Environ(), "GATE_ROOT="+root)
	out, err := cmd.CombinedOutput()
	return string(out), err
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
