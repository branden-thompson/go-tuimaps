package rules

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

// RuleInvisible names a character written literally in a source file that
// the person reading the file cannot see: a control, a format character, a
// bidirectional control. Such a character can make code read as one thing
// and compile as another. An escape names the same character in plain sight.
const RuleInvisible = "invisible"

// checkInvisible reports every invisible character in one source file, test
// files and test-only tooling included, and returns how many it found.
func (c *checker) checkInvisible(src []byte, rel string) (int, error) {
	if rel == "" {
		return 0, errors.New("rules: checkInvisible needs the file's name")
	}
	if len(src) > maxSource {
		return 0, fmt.Errorf("rules: %s is larger than %d bytes", rel, maxSource)
	}
	found, line := 0, 1
	for i := 0; i < len(src); {
		r, size := utf8.DecodeRune(src[i:])
		i += size
		switch {
		case r == '\n':
			line++
		case r == utf8.RuneError && size == 1:
			found++
			c.found = append(c.found, Finding{rel, line, RuleInvisible, "a byte that is not valid UTF-8"})
		case isInvisible(r):
			found++
			c.found = append(c.found, Finding{rel, line, RuleInvisible, fmt.Sprintf("U+%04X is written literally; write it as an escape", r)})
		}
	}
	return found, nil
}

// maxSource bounds one source file.
const maxSource = 4 << 20

// isInvisible reports the characters no reader can see: controls other than
// the tab and the line feed, the line and paragraph separators, and every
// format character - the zero-width and bidirectional controls among them.
func isInvisible(r rune) bool {
	if r == '\t' || r == '\n' {
		return false
	}
	if r == 0x2028 || r == 0x2029 {
		return true
	}
	return unicode.IsControl(r) || unicode.Is(unicode.Cf, r)
}
