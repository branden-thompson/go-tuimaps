package tuimaps_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The two documents this suite is built from. The matrix is the authority
// for what a row means and which release it is in; the mapping says who
// owns each row and which test proves it.
const (
	matrixPath  = "06_docs/02_features/go-tuimaps/02-analysis/parity-matrix.md"
	mappingPath = "06_docs/02_features/go-tuimaps/04-development/parity-mapping.md"
	// M3's denominator for this release, as the matrix states it.
	rowsInRelease = 62
)

// parityRow is one row of the mapping.
type parityRow struct {
	id, behaviour, disposition, owner, test string
}

// mappingRows reads the mapping.
func mappingRows(t *testing.T) []parityRow {
	t.Helper()
	var rows []parityRow
	for _, line := range lines(t, mappingPath) {
		if !strings.HasPrefix(line, "| P-") {
			continue
		}
		parts := cells(line)
		if len(parts) != 5 {
			t.Fatalf("this row of the mapping has %d columns: %s", len(parts), line)
		}
		rows = append(rows, parityRow{id: parts[0], behaviour: parts[1],
			disposition: plain(parts[2]), owner: parts[3], test: strings.Trim(parts[4], "`")})
	}
	return rows
}

// releaseRows is every row of the matrix that this release counts, with the
// disposition the matrix gives it.
func releaseRows(t *testing.T) map[string]string {
	t.Helper()
	found := map[string]string{}
	for _, line := range lines(t, matrixPath) {
		if !strings.HasPrefix(line, "| P-") {
			continue
		}
		parts := cells(line)
		if len(parts) != 7 || parts[5] != "v0.1.0" {
			continue
		}
		// The identifier carries the half of the release it belongs to, as
		// "P-68a [APP]"; the row is known by the identifier alone.
		id, _, _ := strings.Cut(parts[0], " ")
		found[id] = plain(parts[4])
	}
	return found
}

// cells is one table row split into its columns, trimmed. The empty cell a
// row with no note ends in is a column like any other, so the ends are
// dropped rather than trimmed away: trimming would lose it and make the row
// look one column short.
func cells(line string) []string {
	parts := strings.Split(line, "|")
	if len(parts) < 2 {
		return nil
	}
	parts = parts[1 : len(parts)-1]
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// plain is a disposition without the emphasis the documents give it.
func plain(text string) string { return strings.Trim(text, "* ") }

// lines is a document, read.
func lines(t *testing.T, path string) []string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(string(body), "\n")
}

// TestParityMappingComplete is plan task 14.1, and M3's guard: every row
// this release counts is in the mapping, with the disposition the matrix
// gives it, and **every test the mapping names exists**. A row with no test
// cannot be counted, and a test renamed without the mapping being changed
// would leave a row counted by a name nothing answers to.
func TestParityMappingComplete(t *testing.T) {
	rows := mappingRows(t)
	if len(rows) != rowsInRelease {
		t.Errorf("the mapping has %d rows; this release counts %d", len(rows), rowsInRelease)
	}
	wanted := releaseRows(t)
	if len(wanted) != rowsInRelease {
		t.Errorf("the matrix marks %d rows for this release; the count is %d", len(wanted), rowsInRelease)
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if seen[row.id] {
			t.Errorf("%s is in the mapping twice", row.id)
		}
		seen[row.id] = true
		disposition, counted := wanted[row.id]
		if !counted {
			t.Errorf("%s is in the mapping and is not a row this release counts", row.id)
			continue
		}
		if disposition != row.disposition {
			t.Errorf("%s is %q in the mapping and %q in the matrix", row.id, row.disposition, disposition)
		}
		if row.test == "" {
			t.Errorf("%s names no test", row.id)
		}
	}
	for id := range wanted {
		if !seen[id] {
			t.Errorf("%s is counted by the matrix and is in no row of the mapping", id)
		}
	}
	// Every named test is a test that exists.
	written := testsInTree(t)
	for _, row := range rows {
		if !written[row.test] {
			t.Errorf("%s names %s, and no such test is written", row.id, row.test)
		}
	}
}

// testsInTree is every test written anywhere in the repository, by name. It
// reads the source rather than asking the toolchain, so that a test in
// another module - the app's four rows - counts the same as one here.
func testsInTree(t *testing.T) map[string]bool {
	t.Helper()
	found := map[string]bool{}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name := entry.Name(); name != "." && strings.HasPrefix(name, ".") || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(body), "\n") {
			name, ok := strings.CutPrefix(line, "func Test")
			if !ok {
				continue
			}
			if name, _, ok = strings.Cut(name, "("); ok {
				found["Test"+name] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

// TestParitySuite is plan task 14.2: **M3 is reported by running the tests
// the mapping names, not by counting rows by hand.** Every module that owns
// a row is run with exactly those tests selected; a row whose test does not
// run is not a row this release can claim.
func TestParitySuite(t *testing.T) {
	if os.Getenv("TUIMAPS_PARITY_CHILD") != "" {
		t.Skip("this is the run the suite started; it runs the named tests and not the suite itself")
	}
	rows := mappingRows(t)
	byModule := map[string][]string{}
	for _, row := range rows {
		byModule[moduleOf(row.owner)] = append(byModule[moduleOf(row.owner)], row.test)
	}
	work := workspaceFor(t)
	passed := 0
	for module, tests := range byModule {
		ran, err := runNamed(t, module, tests, work)
		if err != nil {
			t.Errorf("%s: %v", module, err)
		}
		passed += ran
	}
	if passed != rowsInRelease {
		t.Errorf("M3 for this release: %d of %d rows proved by a passing test", passed, rowsInRelease)
	}
	t.Logf("M3, this release: %d of %d parity rows proved by a passing test", passed, rowsInRelease)
}

// moduleOf is the module directory a row's owner keeps its tests in. Every
// work package but the app builds part of the library.
func moduleOf(owner string) string {
	if strings.Contains(owner, "app") {
		return filepath.Join("cmd", "tuimaps")
	}
	return "."
}

// workspaceFor writes a throw-away workspace file outside the repository,
// so that a module which requires the library at a version nobody has
// published can be run. The tree itself stays as it is: a tracked workspace
// file is something the gate refuses.
func workspaceFor(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "go.work")
	body := "go 1.25.0\n\nuse (\n\t" + root + "\n\t" + filepath.Join(root, "cmd", "tuimaps") + "\n)\n\n" +
		"replace github.com/branden-thompson/go-tuimaps v0.0.0 => " + root + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// runNamed runs exactly the named tests in one module and answers how many
// of them passed. A test that does not run at all is not counted, which is
// the whole point: M3 counts rows proved, never rows listed.
func runNamed(t *testing.T, module string, tests []string, work string) (int, error) {
	t.Helper()
	selected := "^(" + strings.Join(tests, "|") + ")$"
	cmd := exec.Command("go", "test", "-count=1", "-run", selected, "-v", "./...")
	cmd.Dir = module
	cmd.Env = append(os.Environ(),
		"TUIMAPS_PARITY_CHILD=1", "GOWORK="+work, "GOFLAGS=-mod=readonly", "GOPROXY=off")
	out, err := cmd.CombinedOutput()
	passed := 0
	for _, line := range strings.Split(string(out), "\n") {
		name, ok := strings.CutPrefix(strings.TrimSpace(line), "--- PASS: Test")
		if !ok {
			continue
		}
		name, _, _ = strings.Cut(strings.TrimSpace(name), " ")
		if strings.Contains(name, "/") {
			continue // a sub-test; the row is the test itself
		}
		passed++
	}
	if err != nil {
		return passed, failure(out, err)
	}
	return passed, nil
}

// failure is what went wrong, with the output that shows it, cut to what a
// reader needs.
func failure(out []byte, err error) error {
	text := string(out)
	if len(text) > 4000 {
		text = text[:4000] + "..."
	}
	return &suiteFailure{text: text, err: err}
}

// suiteFailure carries a failed run's own output.
type suiteFailure struct {
	text string
	err  error
}

func (f *suiteFailure) Error() string { return f.err.Error() + "\n" + f.text }
