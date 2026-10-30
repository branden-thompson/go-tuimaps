package tuimaps_test

// m5_record_test.go — plan task P7.1 (M5, NFR-1, watchpost D-209, D-350): the
// time of M5's frame on the reference machine, measured as a release step
// and kept in a record the release check holds to this code.

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// recordM5 measures M5 on this machine and, on the reference machine, writes
// its record.
var recordM5 = flag.Bool("record-m5", false, "measure M5 on this machine and, on the reference machine, write "+m5RecordFile)

// m5RecordFile is M5's record: scripts/gate --cost writes it and
// scripts/gate --release reads it.
const m5RecordFile = "06_docs/02_features/propagation-fields/07-readiness/m5-record.md"

// referenceMachine is the machine M5's targets were set on (watchpost D-96,
// D-209).
const referenceMachine = "Apple M5 Pro"

// m5Runs is how many times each frame is measured; the record keeps the
// median.
const m5Runs = 5

// The record's rows: the code, the machine, the Go version and the moment,
// then each of m5Sizes' frames (m5Row) and the allocations of a frame at
// 400 by 110.
const (
	rowTree    = "Code tree"
	rowMachine = "Machine"
	rowGo      = "Go"
	rowWhen    = "Measured"
	rowAllocs  = "Allocations a frame, 400 × 110"
)

// m5Row is the record's row for a frame of m5Sizes' i-th size, in
// milliseconds.
func m5Row(i int) string {
	return fmt.Sprintf("%d × %d, ms a frame", m5Sizes[i].size.Cols, m5Sizes[i].size.Rows)
}

// middleOf is the middle of an odd number of figures.
func middleOf(xs []float64) float64 {
	s := slices.Clone(xs)
	slices.Sort(s)
	return s[len(s)/2]
}

// machineName is this machine's processor, as the kernel names it.
func machineName() string {
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	if b, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if name, ok := strings.CutPrefix(line, "model name"); ok {
				return strings.TrimSpace(strings.TrimLeft(name, " \t:")) + ", " + strconv.Itoa(runtime.NumCPU()) + " CPUs"
			}
		}
	}
	return "unknown (" + runtime.GOOS + "/" + runtime.GOARCH + "), " + strconv.Itoa(runtime.NumCPU()) + " CPUs"
}

// frameTime is one measure of globePan's frame: milliseconds and allocations
// a frame.
func frameTime(i int, overWater bool) (ms, allocs float64) {
	r := testing.Benchmark(func(b *testing.B) {
		step := globePan(b, m5Sizes[i].size, overWater)
		b.ReportAllocs()
		b.ResetTimer()
		for k := 0; b.Loop(); k++ {
			step(k)
		}
	})
	return float64(r.NsPerOp()) / 1e6, float64(r.AllocsPerOp())
}

// TestTheM5CostIsRecorded is P7.1's measure (watchpost D-350): each of
// m5Sizes' frames, the slower of over the sea and stopping at the shore,
// each the median of five runs, with the machine, the Go version and the code
// they were measured on. It runs only under scripts/gate --cost, which names
// the code (TUIMAPS_TREE), on a machine otherwise at rest. On the reference
// machine it writes the record and a target missed fails it; on any other it
// writes nothing and holds no target - the figures are in its log.
func TestTheM5CostIsRecorded(t *testing.T) {
	if !*recordM5 {
		t.Skip("a release step: scripts/gate --cost runs it")
	}
	if raceDetector {
		t.Fatal("M5 is measured without the race detector, which slows every frame")
	}
	tree := os.Getenv("TUIMAPS_TREE")
	if tree == "" {
		t.Fatal("TUIMAPS_TREE names the code measured; scripts/gate --cost sets it")
	}
	machine := machineName()
	rows := [][2]string{
		{rowTree, "`" + tree + "`"},
		{rowMachine, machine},
		{rowGo, runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH},
		{rowWhen, time.Now().UTC().Format(time.RFC3339)},
	}
	var allocs float64
	for i, s := range m5Sizes {
		slowest := 0.0
		for _, overWater := range []bool{true, false} {
			var runs []float64
			for range m5Runs {
				ms, a := frameTime(i, overWater)
				runs = append(runs, ms)
				if i == 0 {
					allocs = max(allocs, a)
				}
			}
			slowest = max(slowest, middleOf(runs))
		}
		v := strconv.FormatFloat(slowest, 'f', 3, 64)
		if s.target > 0 {
			v += " (target " + strconv.FormatFloat(s.target, 'f', -1, 64) + ")"
		}
		rows = append(rows, [2]string{m5Row(i), v})
	}
	rows = append(rows, [2]string{rowAllocs, strconv.FormatFloat(allocs, 'f', 0, 64)})
	var b strings.Builder
	b.WriteString("# M5's record: the globe frame on the reference machine\n\n")
	b.WriteString("Written by `scripts/gate --cost` (plan task P7.1, watchpost D-350) on the reference machine, an " + referenceMachine + ". A frame is\n")
	b.WriteString("one cell's pan over the whole world under a 2° field filled and drawn with labelled contours, the night and\n")
	b.WriteString("its terminator, and a reach's edge (`globePan`, `globe_bench_test.go`); each figure is the slower of the\n")
	b.WriteString("field over the sea and stopping at the shore, each the median of " + strconv.Itoa(m5Runs) + " runs. 800 × 220 is recorded, a known\n")
	b.WriteString("limit, not M5's (watchpost D-261). `scripts/gate --release` refuses a final v0.3 tag while this record is\n")
	b.WriteString("missing, names other code or another machine, or misses a target.\n\n")
	b.WriteString("| Field | Value |\n|---|---|\n")
	for _, r := range rows {
		b.WriteString("| " + r[0] + " | " + r[1] + " |\n")
	}
	t.Logf("%s", b.String())
	if !strings.Contains(machine, referenceMachine) {
		t.Logf("%s is not the reference machine, an %s: recorded in this log, not held to M5's targets; %s is left as it is", machine, referenceMachine, m5RecordFile)
		return
	}
	if err := os.WriteFile(m5RecordFile, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, problem := range m5RecordProblems(b.String(), tree) {
		t.Error(problem)
	}
}

// m5RecordProblems is every reason the release refuses a record for the code
// named tree: no record, another tree, another machine, a held frame's figure
// missing, unreadable or past its target.
func m5RecordProblems(record, tree string) []string {
	fields := map[string]string{}
	for _, line := range strings.Split(record, "\n") {
		cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
		if len(cells) == 2 {
			fields[strings.TrimSpace(cells[0])] = strings.TrimSpace(cells[1])
		}
	}
	var problems []string
	if got := strings.Trim(fields[rowTree], "`"); got != tree {
		problems = append(problems, fmt.Sprintf("the record names code %q; this is %q: run scripts/gate --cost on the release's code", got, tree))
	}
	if !strings.Contains(fields[rowMachine], referenceMachine) {
		problems = append(problems, fmt.Sprintf("the record was measured on %q, not the reference machine, an %s", fields[rowMachine], referenceMachine))
	}
	for i, s := range m5Sizes {
		if s.target == 0 {
			continue
		}
		figure, _, _ := strings.Cut(fields[m5Row(i)], " ")
		v, err := strconv.ParseFloat(figure, 64)
		if err != nil || math.IsNaN(v) {
			problems = append(problems, fmt.Sprintf("%s: %q is no figure", m5Row(i), fields[m5Row(i)]))
			continue
		}
		if v > s.target {
			problems = append(problems, fmt.Sprintf("%s is %.3f, past its target %v (M5)", m5Row(i), v, s.target))
		}
	}
	return problems
}

// TestTheM5RecordHoldsForThisCode is P7.1's release check (watchpost D-350):
// the record names the code released (TUIMAPS_TREE, set by scripts/gate
// --release), the reference machine, and every target met.
func TestTheM5RecordHoldsForThisCode(t *testing.T) {
	if os.Getenv("TUIMAPS_RELEASE") != "1" {
		t.Skip("a release step: scripts/gate --release runs it")
	}
	body, err := os.ReadFile(m5RecordFile)
	if err != nil {
		t.Fatalf("%v: run scripts/gate --cost on the reference machine and commit the record", err)
	}
	for _, problem := range m5RecordProblems(string(body), os.Getenv("TUIMAPS_TREE")) {
		t.Error(problem)
	}
}

// TestAnM5RecordIsRefused (P7.1): the release check refuses a record of other
// code, another machine, a missed target or a figure that is none, holds no
// target on 800 by 220, and passes a record that holds.
func TestAnM5RecordIsRefused(t *testing.T) {
	good := "| Field | Value |\n|---|---|\n| " + rowTree + " | `abc123` |\n| " + rowMachine + " | Apple M5 Pro |\n| " +
		m5Row(0) + " | 7.500 (target 16) |\n| " + m5Row(1) + " | 2.100 (target 8) |\n| " + m5Row(2) + " | 99.000 |\n"
	if p := m5RecordProblems(good, "abc123"); len(p) != 0 {
		t.Fatalf("a record that holds is refused: %v", p)
	}
	for _, c := range []struct{ name, record, tree, want string }{
		{"no record", "", "abc123", "names code"},
		{"other code", good, "def456", "names code"},
		{"another machine", strings.Replace(good, "Apple M5 Pro", "Apple M1", 1), "abc123", "reference machine"},
		{"a slow frame at 400 × 110", strings.Replace(good, "7.500", "16.100", 1), "abc123", "past its target"},
		{"a slow frame at 200 × 56", strings.Replace(good, "2.100", "8.100", 1), "abc123", "past its target"},
		{"no figure", strings.Replace(good, "2.100", "NaN", 1), "abc123", "no figure"},
		{"no row", strings.Replace(good, m5Row(0), "Something else", 1), "abc123", "no figure"},
	} {
		p := m5RecordProblems(c.record, c.tree)
		if !strings.Contains(strings.Join(p, "; "), c.want) {
			t.Errorf("%s: refused for %v; want a reason saying %q", c.name, p, c.want)
		}
	}
}
