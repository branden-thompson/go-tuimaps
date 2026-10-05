package tuimaps_test

// exposure_test.go — the release checklist's row 2.5: no address or name
// that should not leave the tree. The repository is public; a home
// directory or a person's email address in a tracked file is published.

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// publicSenders are addresses a tracked file may carry because they are
// published by their owners: a public agency's sender in a recorded alert.
var publicSenders = map[string]bool{"w-nws.webmaster@noaa.gov": true}

// TestNoPathOrAddressLeavesTheTree: no tracked text file names a home
// directory, and every email address in one is at a domain reserved for
// examples or a listed public sender.
func TestNoPathOrAddressLeavesTheTree(t *testing.T) {
	out, err := exec.Command("git", "ls-files", "-z").Output()
	if err != nil {
		t.Fatalf("the tracked files could not be listed (%v); this check needs the repository", err)
	}
	home := regexp.MustCompile(`/(Users|home)/[a-z][a-z0-9_-]+`)
	// An address ends in a name of letters: a module version (tool@v1.8.0)
	// or a place written Name@lon,lat is not one.
	email := regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9][A-Za-z0-9-]*(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}\b`)
	reserved := regexp.MustCompile(`(?i)@([a-z0-9-]+\.)*(example\.(com|org|net)|example|test|invalid|localhost)$`)
	files := 0
	for _, name := range strings.Split(strings.TrimRight(string(out), "\x00"), "\x00") {
		body, err := os.ReadFile(name)
		if err != nil || bytes.IndexByte(body, 0) >= 0 {
			continue // gone from the working tree, or not text
		}
		files++
		if m := home.Find(body); m != nil {
			t.Errorf("%s names a home directory: %s", name, m)
		}
		for _, m := range email.FindAll(body, -1) {
			address := strings.TrimRight(string(m), ".")
			if !reserved.MatchString(address) && !publicSenders[address] && !strings.HasPrefix(address, "noreply@") {
				t.Errorf("%s carries an email address: %s", name, address)
			}
		}
	}
	if files < 100 {
		t.Fatalf("only %d tracked text files were read; the check has stopped seeing the tree", files)
	}
}
