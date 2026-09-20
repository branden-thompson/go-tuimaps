package textsafe

import (
	"os"
	"strings"
	"testing"
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

// TestNoHeadlessMark: a cleaned text never begins with a character that
// attaches itself to the one before it. Such a character would take a cell of
// its own here and then dissolve into whatever came before it in a row, which
// would make the row one cell narrower than its characters say (NFR-8). Found
// by the renderer's own fuzz target, which drew a credit line of one such
// mark - and found a second time, by the same target on the floor toolchain,
// in U+113C2, a vowel sign the pinned segmenter joins and that release's
// character categories had never heard of.
func TestNoHeadlessMark(t *testing.T) {
	for _, s := range []string{"\u0301", "\ua9c0", "\u0301abc", "\u20e3", "\u0e31 x", "\U000113c2"} {
		got := Clean(s)
		if got.String() == "" {
			continue
		}
		// **Asked of the pinned segmenter, not of the character categories.**
		// The categories move with the Go release; the segmenter is the table
		// the library pins, and it is the one that decides what joins what.
		if headlessMark(got.String()) {
			first, _ := utf8.DecodeRuneInString(got.String())
			t.Errorf("Clean(%q) begins with %U, which joins the character before it", s, first)
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

// TestCleanKeepsCleanTextWithoutCopyingIt is D-115's half in this package:
// **text that is already clean is handed back, never copied.** A copy costs
// bytes in proportion to the text, so the rule is put as the thing a copy
// could not satisfy: **the price does not move when the text grows.** The
// same words a thousand times over clean for what one of them cost.
//
// The price itself may be one small fixed probe and never a second - the
// question the leading character is put to, five bytes that do not leave
// with the answer (D-119). It is a constant, so it cannot be the cost of a
// copy, and it is paid where text *enters* the library rather than on any
// frame's path: a name is cleaned once per tile and a host's text once per
// hand-in (D-120). Latin text does not pay it at all.
func TestCleanKeepsCleanTextWithoutCopyingIt(t *testing.T) {
	for _, s := range []string{"", "Warning", "Tornado Warning", "Great Falls", "Ceuta y Melilla",
		"149 km", "a name with an accent: Bogot\u00E1", "\u2800\u2801"} {
		got := Clean(s)
		if got.String() != s {
			t.Errorf("%q came back as %q; it needed no cleaning", s, got.String())
		}
		long := strings.Repeat(s, 1000)
		if grown := Clean(long); grown.String() != long {
			t.Errorf("a thousand of %q came back changed", s)
		}
		once := testing.AllocsPerRun(50, func() { Clean(s) })
		thousand := testing.AllocsPerRun(50, func() { Clean(long) })
		if once > 1 || thousand != once {
			t.Errorf("cleaning %q allocates %.0f times, and a thousand of it %.0f: a cost that grows with the text is a copy", s, once, thousand)
		}
	}
	// Text that does need cleaning is still cleaned, and the fast path has
	// not made a liar of it.
	for _, c := range []struct{ in, want string }{
		{"a\x00b", "ab"},
		{"a\u202Eb", "ab"},
		{"\xffz", "\uFFFDz"},
		{"a\u2028b", "ab"},
	} {
		if got := Clean(c.in).String(); got != c.want {
			t.Errorf("Clean(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestEveryCharacterTakesTheCellsItSaysAfterAnother is the same rule put to
// every character there is, so that no future toolchain, width table or
// segmenter can quietly reintroduce the defect in a character nobody thought
// to list: whatever a cleaned character measures on its own, that is what it
// adds to a row that already holds something (NFR-8).
func TestEveryCharacterTakesTheCellsItSaysAfterAnother(t *testing.T) {
	base := Clean("x")
	for r := range rune(utf8.MaxRune + 1) {
		if r >= 0xD800 && r <= 0xDFFF { // halves of a surrogate pair are not characters
			continue
		}
		alone := Clean(string(r))
		joined := Width(Clean("x" + string(r)))
		if joined != Width(base)+Width(alone) {
			t.Fatalf("%U is %d cells alone and %d after another character", r, Width(alone), joined-Width(base))
		}
	}
}
