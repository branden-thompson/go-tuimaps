package tuimaps_test

// contract_breaks_test.go — v0.1.0 NFR-22 (D-60), v0.2.0 D-137: the public
// contract breaks nothing of the last release that contract section 12 does
// not list.

import (
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fetch"
)

// lastRelease is the tag whose public surface this release is held to: the
// newest final release before the one the library says it is.
const lastRelease = "v0.2.2"

// releaseNumber is a final release tag's version as one comparable number,
// and false for any other tag - a candidate, a nested module's.
func releaseNumber(tag string) (int, bool) {
	m := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`).FindStringSubmatch(tag)
	if m == nil {
		return 0, false
	}
	n := 0
	for _, part := range m[1:] {
		v, _ := strconv.Atoi(part)
		n = n*1000 + v
	}
	return n, true
}

// TestTheContractIsHeldToTheLastRelease (watchpost D-344's SHIP close-out):
// the release the contract check reads is the newest final tag before this
// one, so a release cannot be checked against a surface two releases old.
func TestTheContractIsHeldToTheLastRelease(t *testing.T) {
	out, err := exec.Command("git", "tag", "--list", "v*").Output()
	if err != nil {
		t.Fatalf("the tags could not be read from git (%v); a clone without its tags cannot run this check", err)
	}
	this, ok := releaseNumber(fetch.Version)
	if !ok {
		t.Fatalf("the library says it is %q, no release version", fetch.Version)
	}
	newest, newestTag := -1, ""
	for _, tag := range strings.Fields(string(out)) {
		if n, final := releaseNumber(tag); final && n < this && n > newest {
			newest, newestTag = n, tag
		}
	}
	if newestTag != lastRelease {
		t.Errorf("the contract check is held to %s; the last release before %s is %s", lastRelease, fetch.Version, newestTag)
	}
}

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
