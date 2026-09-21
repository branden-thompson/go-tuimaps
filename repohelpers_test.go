package tuimaps_test

import (
	"bytes"
	"strings"
)

// allowedModules is the whole of the library's permitted module graph: the
// library itself and the two text modules of rulings D-75 and D-81.
var allowedModules = map[string]bool{
	"github.com/branden-thompson/go-tuimaps": true,
	"github.com/mattn/go-runewidth":          true,
	"github.com/clipperhouse/uax29/v2":       true,
}

// disallowedModules reads the output of "go list -m all" and returns the
// path of every module that is not allowed, in the order listed.
func disallowedModules(list string) []string {
	var strangers []string
	for line := range strings.SplitSeq(list, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if !allowedModules[fields[0]] {
			strangers = append(strangers, fields[0])
		}
	}
	return strangers
}

// replaceLines returns every line of a module file that opens a replace
// directive, single-line or block. Comment lines are ignored.
func replaceLines(mod []byte) []string {
	var found []string
	for _, raw := range strings.Split(string(mod), "\n") {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "//") {
			continue
		}
		if line == "replace" || strings.HasPrefix(line, "replace ") || strings.HasPrefix(line, "replace(") {
			found = append(found, line)
		}
	}
	return found
}

// hasLine reports whether data holds want as a whole line.
func hasLine(data []byte, want string) bool {
	for _, l := range bytes.Split(data, []byte("\n")) {
		if string(bytes.TrimSpace(l)) == want {
			return true
		}
	}
	return false
}
