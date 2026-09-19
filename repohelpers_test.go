package tuimaps_test

import (
	"bytes"
	"strings"
)

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
