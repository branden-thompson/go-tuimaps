package tuimaps_test

// contract_breaks_test.go — v0.1.0 NFR-22 (D-60), v0.2.0 D-137: the public
// contract breaks nothing of the last release that contract section 12 does
// not list.

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// lastRelease is the tag whose public surface this release is held to.
const lastRelease = "v0.1.0"

// surfaceNames is each line of a surface snapshot by the name it declares,
// "kind Name" - the snapshot of v0.1.0 names, today's adds the signature.
func surfaceNames(snapshot string) map[string]string {
	names := map[string]string{}
	for _, line := range strings.Split(snapshot, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name, _, _ := strings.Cut(fields[1], "(")
		names[fields[0]+" "+name] = line
	}
	return names
}

// TestNothingOfTheLastReleaseBreaksUnlisted (D-137): every name the last
// release exported is still exported, or is named in contract section 12's
// ruled breaks. Where the last release's snapshot carries a signature, the
// signature is held too; v0.1.0's carries names alone, so from v0.1.0 the
// names are what is held.
func TestNothingOfTheLastReleaseBreaksUnlisted(t *testing.T) {
	old, err := exec.Command("git", "show", lastRelease+":"+surfaceFile).Output()
	if err != nil {
		t.Fatalf("the last release's surface could not be read from git (%v); a clone without its tags cannot run this check", err)
	}
	now, err := os.ReadFile(surfaceFile)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := os.ReadFile(contractFile)
	if err != nil {
		t.Fatal(err)
	}
	section := string(contract)
	at := strings.Index(section, "## 12 ")
	if at < 0 {
		t.Fatal("the contract has no section 12, the ruled breaks")
	}
	section = section[at:]
	if next := regexp.MustCompile(`(?m)^## 1[3-9] `).FindStringIndex(section); next != nil {
		section = section[:next[0]]
	}
	was, is := surfaceNames(string(old)), surfaceNames(string(now))
	if len(was) < 100 {
		t.Fatalf("the last release's surface read as %d names; the snapshot format has changed under the reader", len(was))
	}
	for name, line := range was {
		bare := name[strings.Index(name, " ")+1:]
		listed := strings.Contains(section, "`"+bare+"`") || strings.Contains(section, "`"+bare[strings.LastIndex(bare, ".")+1:]+"`")
		current, kept := is[name]
		changed := kept && strings.Contains(line, "(") && current != line
		if (!kept || changed) && !listed {
			t.Errorf("%s: %s, and contract section 12 does not name it", map[bool]string{true: "changed", false: "removed"}[kept], line)
		}
	}
}
