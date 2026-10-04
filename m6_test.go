package tuimaps_test

// m6_test.go — v0.2.0 plan task L10.10 (M6): five consecutive clean full runs
// of the gate before SHIP, counted from 06_docs/gate-runs.md. Only full runs
// count: the docs lane, the fuzz mode and the release check are not the gate,
// and a run with an override of the gate's settings (a hosted CI leg handing
// its second architecture to its sibling) is not a whole run of it.

import (
	"os"
	"strings"
	"testing"
)

// m6Streak is the number of consecutive green full runs at the end of the
// log, and the full runs it read.
func m6Streak(log string) (streak, full int) {
	for _, line := range strings.Split(log, "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 8 || strings.TrimSpace(cells[4]) != "full" || strings.TrimSpace(cells[7]) != "-" {
			continue
		}
		full++
		if strings.TrimSpace(cells[5]) == "green" {
			streak++
		} else {
			streak = 0
		}
	}
	return streak, full
}

// TestM6CountsOnlyFullRunsAndResetsOnAFailure: the counter itself, on a
// planted log.
func TestM6CountsOnlyFullRunsAndResetsOnAFailure(t *testing.T) {
	log := strings.Join([]string{
		"| t | a | x | full | green | 1 | - |",
		"| t | a | x | full | FAILED | 1 | - |",
		"| t | b | x | full | green | 1 | - |",
		"| t | b | x | docs | FAILED | 1 | - |",
		"| t | b | x | release | FAILED | 1 | - |",
		"| t | c | x | full | green | 1 | - |",
		"| t | c | x | full | INTERRUPTED | 1 | - |",
		"| t | d | x | full | green | 1 | - |",
		"| t | e | x | full | FAILED | 1 | GATE_ARCH_LEG=sibling |",
	}, "\n")
	if streak, full := m6Streak(log); streak != 1 || full != 6 {
		t.Errorf("streak %d over %d full runs; want 1 over 6 (an interrupted run is not clean)", streak, full)
	}
}

// TestM6ReportsTheCurrentRunOfGreen reports M6 as the log stands. It fails
// only when the log cannot be read as one: the count is evidence for the
// release checklist's row, not a gate of its own until SHIP.
func TestM6ReportsTheCurrentRunOfGreen(t *testing.T) {
	body, err := os.ReadFile("06_docs/gate-runs.md")
	if err != nil {
		t.Fatal(err)
	}
	streak, full := m6Streak(string(body))
	if full == 0 {
		t.Fatal("no full run read from the gate log: its table no longer reads as one")
	}
	t.Logf("M6: %d consecutive clean full run(s) of the gate; five are needed before SHIP", streak)
}
