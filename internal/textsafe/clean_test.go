package textsafe

import (
	"os"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// TestInvalidUTF8BecomesReplacement is plan task 02.1.
func TestInvalidUTF8BecomesReplacement(t *testing.T) {
	cases := map[string]string{
		"Fort Wayne":   "Fort Wayne",
		"a\xffb":       "a\uFFFDb",
		"\xc3(":        "\uFFFD(",
		"ok\xe2\x82":   "ok\uFFFD\uFFFD",
		"Z\u00FCrich":  "Z\u00FCrich",
		"\u6771\u4EAC": "\u6771\u4EAC",
		"":             "",
	}
	for in, want := range cases {
		got := Clean(in).String()
		if got != want {
			t.Errorf("Clean(%q) = %q, want %q", in, got, want)
		}
		if !utf8.ValidString(got) {
			t.Errorf("Clean(%q) is not valid UTF-8", in)
		}
	}
}

// TestControlsDropped is plan task 02.2.
func TestControlsDropped(t *testing.T) {
	cases := map[string]string{
		"a\x1b[31mred\x1b[0m":       "a[31mred[0m", // the escape byte goes; what is left is inert text
		"bell\x07 null\x00 del\x7f": "bell null del",
		"tab\there\nnext\r":         "tabherenext",
		"c1\u0080\u009B\u009F!":     "c1!",
		"line\u2028para\u2029end":   "lineparaend",
	}
	for in, want := range cases {
		if got := Clean(in).String(); got != want {
			t.Errorf("Clean(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestBidiDropped is plan task 02.3: a name must not be able to reorder the
// text around it.
func TestBidiDropped(t *testing.T) {
	for _, r := range []rune{0x061C, 0x200E, 0x200F, 0x202A, 0x202B, 0x202C, 0x202D, 0x202E, 0x2066, 0x2067, 0x2068, 0x2069} {
		in := "ab" + string(r) + "cd"
		if got := Clean(in).String(); got != "abcd" {
			t.Errorf("U+%04X: Clean(%q) = %q, want %q", r, in, got, "abcd")
		}
	}
}

// TestZeroWidthOutsideClusterDropped is plan task 02.4, first half.
func TestZeroWidthOutsideClusterDropped(t *testing.T) {
	cases := map[string]string{
		"a\u200Bb":         "ab",     // zero-width space
		"a\u2060b":         "ab",     // word joiner
		"\xef\xbb\xbfname": "name",   // byte-order mark
		"\u200Dstart":      "start",  // a joiner with nothing to join
		"\u0301start":      "start",  // a combining accent with no base
		"a\u00ADb":         "ab",     // soft hyphen
		"tag\U000E0041end": "tagend", // a tag character on its own
	}
	for in, want := range cases {
		if got := Clean(in).String(); got != want {
			t.Errorf("Clean(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestZeroWidthInsideClusterKept is plan task 02.4, second half.
func TestZeroWidthInsideClusterKept(t *testing.T) {
	for _, in := range []string{
		"e\u0301",              // a letter with a combining accent
		"\U0001F1FA\U0001F1F8", // a flag: two regional indicators
		"\U0001F468\u200D\U0001F469\u200D\U0001F467", // a family joined by zero-width joiners
		"\u2764\uFE0F",             // a heart with its variation selector
		"\u0915\u094D\u0937\u093F", // a Devanagari conjunct
	} {
		if got := Clean(in).String(); got != in {
			t.Errorf("Clean(%q) = %q; a cluster's own marks, joiners and selectors stay with it", in, got)
		}
	}
}

// TestTagCharactersStayOnlyInAFlag: tag characters are invisible, and
// outside the black-flag sequences they were made for they can carry text
// nobody sees.
func TestTagCharactersStayOnlyInAFlag(t *testing.T) {
	scotland := "\U0001F3F4\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F"
	if got := Clean(scotland).String(); got != scotland {
		t.Errorf("a flag's own tag sequence was damaged: %q", got)
	}
	smuggled := "Paris\U000E0069\U000E0067\U000E006E\U000E006F\U000E0072\U000E0065"
	if got := Clean(smuggled).String(); got != "Paris" {
		t.Errorf("Clean(%q) = %q; tag characters after a letter must go", smuggled, got)
	}
}

func TestCleanIsIdempotent(t *testing.T) {
	for _, in := range []string{"a\x1b]0;title\x07b", "e\u0301\u200B\u202Ex", strings.Repeat("\xff", 9)} {
		once := Clean(in).String()
		if twice := Clean(once).String(); twice != once {
			t.Errorf("Clean(Clean(%q)) = %q, but Clean gave %q", in, twice, once)
		}
	}
}

// TestNoHeadlessMark: a cleaned text never begins with a combining mark. One
// would attach itself to whatever came before it in a row, which would make
// the row one cell narrower than its characters say (NFR-8). Found by the
// renderer's own fuzz target, which drew a credit line of one such mark.
func TestNoHeadlessMark(t *testing.T) {
	for _, s := range []string{"\u0301", "\ua9c0", "\u0301abc", "\u20e3", "\u0e31 x"} {
		got := Clean(s)
		if got.String() == "" {
			continue
		}
		first, _ := utf8.DecodeRuneInString(got.String())
		if unicode.In(first, unicode.Mn, unicode.Mc, unicode.Me) {
			t.Errorf("Clean(%q) begins with the mark %U", s, first)
		}
		// Measured after anything else, it still takes the cells it says.
		before := Width(Clean("x"))
		if joined := Width(Clean("x" + got.String())); joined != before+Width(got) {
			t.Errorf("Clean(%q) is %d cells alone and %d after another character", s, Width(got), joined-before)
		}
	}
	// A mark after its own base is kept: it is part of that cluster.
	if got := Clean("e\u0301"); got.String() != "e\u0301" {
		t.Errorf("Clean of a letter with its accent gave %q", got.String())
	}
}
