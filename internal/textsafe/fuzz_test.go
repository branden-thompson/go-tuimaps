package textsafe

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// unsafeRune reports what may never leave the library: an escape byte or any
// other control, a line or paragraph separator, a bidirectional control.
func unsafeRune(r rune) bool {
	return isControl(r) || isBidiControl(r)
}

// checkSafe fails the test if s could harm a terminal or a speech engine.
func checkSafe(t *testing.T, what, in, s string) {
	t.Helper()
	if !utf8.ValidString(s) {
		t.Fatalf("%s(%q) is not valid UTF-8: %q", what, in, s)
	}
	if i := strings.IndexFunc(s, unsafeRune); i >= 0 {
		t.Fatalf("%s(%q) = %q holds an unsafe character at byte %d", what, in, s, i)
	}
	if strings.ContainsRune(s, 0x1b) {
		t.Fatalf("%s(%q) holds an escape byte", what, in)
	}
}

// FuzzClean is plan task 02.9, on both of FR-34's paths: text for a frame
// and text handed back as data.
func FuzzClean(f *testing.F) {
	for _, seed := range []string{
		"", "Fort Wayne", "a\x1b[31mred\x1b[0m", "\x1b]0;title\x07", "\x9b31m", "a\xffb\xfe", "\xe2\x80",
		"e\\u0301", "\xe2\x80\x8b", "\xe2\x80\xae", "\xf3\xa0\x81\x81", "\xef\xbb\xbf", "\xf0\x9f\x87\xba\xf0\x9f\x87\xb8",
		strings.Repeat("\xcc\x81", 300), strings.Repeat("x", 300),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, in string) {
		cleaned := Clean(in).String()
		checkSafe(t, "Clean", in, cleaned)
		if again := Clean(cleaned).String(); again != cleaned {
			t.Fatalf("Clean is not idempotent on %q: %q then %q", in, cleaned, again)
		}
		if len(cleaned) > len(in)*3 {
			t.Fatalf("Clean(%q) grew from %d to %d bytes", in, len(in), len(cleaned))
		}
		quoted := Quote(in).String()
		checkSafe(t, "Quote", in, quoted)
		if w := Width(Fit(Clean(in), 7)); w > 7 {
			t.Fatalf("Fit(%q, 7) is %d cells wide", in, w)
		}
		if id, err := ID(in); err == nil {
			checkSafe(t, "ID", in, id.String())
			if id.String() != in {
				t.Fatalf("ID(%q) returned %q; an id is never altered", in, id)
			}
		}
	})
}
