package tuimaps_test

// plan_map_test.go — v0.2.0 plan task L10.12 (D-72): the integration map and
// the plan are kept in step. The plan names no work package the map does not,
// and records the map commit it was reconciled against.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

const (
	planPath = "06_docs/02_features/radar-loops/04-development/implementation-plan.md"
	mapPath  = "06_docs/02_features/radar-loops/03-architecture-design/integration-map.md"
)

// workPackages are the WP- names a document cites.
func workPackages(t *testing.T, rel string) map[string]bool {
	t.Helper()
	body, err := os.ReadFile(filepath.FromSlash(rel))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]bool{}
	for _, wp := range regexp.MustCompile(`WP-[A-Z]+[0-9]+`).FindAllString(string(body), -1) {
		out[wp] = true
	}
	return out
}

// TestThePlanNamesOnlyTheMapsWorkPackages: a work package the plan cites and
// the map does not is a plan out of step with the one page both plans follow.
func TestThePlanNamesOnlyTheMapsWorkPackages(t *testing.T) {
	onMap := workPackages(t, mapPath)
	plan := workPackages(t, planPath)
	if len(plan) == 0 {
		t.Fatal("the plan names no work package: the pattern no longer reads it")
	}
	for _, wp := range notOnMap(plan, onMap) {
		t.Errorf("the plan names %s, which the integration map does not", wp)
	}
}

// notOnMap is every work package the plan names and the map does not.
func notOnMap(plan, onMap map[string]bool) []string {
	var out []string
	for wp := range plan {
		if !onMap[wp] {
			out = append(out, wp)
		}
	}
	return out
}

// TestAPackageMissingFromTheMapIsFound: the check's own control - a plan
// naming a package the map lacks is caught.
func TestAPackageMissingFromTheMapIsFound(t *testing.T) {
	got := notOnMap(map[string]bool{"WP-L1": true, "WP-L99": true}, map[string]bool{"WP-L1": true})
	if len(got) != 1 || got[0] != "WP-L99" {
		t.Errorf("a plan naming WP-L99 against a map without it: %v", got)
	}
}

// TestThePlanRecordsTheMapCommitItFollows: the plan says which commit of the
// map it was reconciled against, and that commit holds the map.
func TestThePlanRecordsTheMapCommitItFollows(t *testing.T) {
	body, err := os.ReadFile(filepath.FromSlash(planPath))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile("Reconciled with the integration map at `([0-9a-f]{7,40})`").FindStringSubmatch(string(body))
	if m == nil {
		t.Fatal("the plan does not record the map commit it was reconciled against (\"Reconciled with the integration map at `<commit>`\")")
	}
	if out, err := exec.Command("git", "cat-file", "-e", m[1]+":"+mapPath).CombinedOutput(); err != nil {
		t.Errorf("the recorded commit %s does not hold the map: %v %s", m[1], err, out)
	}
}
