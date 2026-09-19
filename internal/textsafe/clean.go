// Package textsafe makes text safe wherever it leaves the library, whatever
// its source: tile names, source metadata, style text, a host's labels,
// credits, legends, errors and warnings (FR-34). It is the one way out, and
// the one package that may import third-party code: the width table the
// first host measures with, and the grapheme segmenter that table uses.
package textsafe

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clipperhouse/uax29/v2/graphemes"
	"github.com/mattn/go-runewidth"
)

// Text is text that has been cleaned. Its field is unexported, so no other
// package can make one from a plain string: Clean is the only way in for
// outside text.
type Text struct {
	s string
}

// String returns the cleaned text.
func (t Text) String() string { return t.s }

// Clean makes s safe to show or to hand back as data, in four steps:
//
//  1. invalid UTF-8 becomes U+FFFD, one for each bad byte;
//  2. control characters are dropped: C0, DEL, C1, and the line and
//     paragraph separators U+2028 and U+2029;
//  3. bidirectional controls are dropped: U+061C, U+200E, U+200F, U+202A to
//     U+202E, U+2066 to U+2069;
//  4. the text is cut into grapheme clusters. A cluster made only of
//     zero-width characters - one standing on its own - is dropped, and so
//     is any cluster that occupies no cell. Combining marks, joiners and
//     variation selectors inside a cluster stay with it. Tag characters,
//     U+E0000 to U+E007F, stay only in a cluster that begins with the black
//     flag U+1F3F4, the one place they belong; anywhere else they are an
//     invisible channel, and are removed from the cluster.
//
// Clean is idempotent, and its result never holds an escape byte.
func Clean(s string) Text {
	var kept strings.Builder
	kept.Grow(len(s))
	for _, r := range s { // an invalid byte arrives as utf8.RuneError
		if isControl(r) || isBidiControl(r) {
			continue
		}
		kept.WriteRune(r)
	}
	var out strings.Builder
	out.Grow(kept.Len())
	clusters := graphemes.FromString(kept.String())
	for range kept.Len() { // a cluster is at least one byte, so this bounds the walk
		if !clusters.Next() {
			break
		}
		cluster := stripStrayTags(clusters.Value())
		if onlyZeroWidth(cluster) || clusterWidth(cluster) == 0 {
			continue
		}
		out.WriteString(cluster)
	}
	return Text{s: out.String()}
}

// isControl reports the characters of step 2.
func isControl(r rune) bool {
	switch {
	case r < 0x20, r == 0x7F: // C0 and DEL
		return true
	case r >= 0x80 && r <= 0x9F: // C1
		return true
	case r == 0x2028, r == 0x2029: // line and paragraph separators
		return true
	}
	return false
}

// isBidiControl reports the characters of step 3.
func isBidiControl(r rune) bool {
	switch {
	case r == 0x061C, r == 0x200E, r == 0x200F:
		return true
	case r >= 0x202A && r <= 0x202E:
		return true
	case r >= 0x2066 && r <= 0x2069:
		return true
	}
	return false
}

// isZeroWidth is the documented table of characters that occupy no cell:
// every nonspacing and enclosing mark, every format character - among them
// the zero-width space U+200B, the joiners U+200C and U+200D, the word
// joiner U+2060, the byte-order mark U+FEFF, the soft hyphen U+00AD, the
// interlinear annotation marks U+FFF9 to U+FFFB and the tag characters -
// and the variation selectors.
func isZeroWidth(r rune) bool {
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf, unicode.Variation_Selector)
}

// onlyZeroWidth reports whether a cluster has nothing in it that occupies a
// cell: a zero-width character standing outside any real cluster.
func onlyZeroWidth(cluster string) bool {
	for _, r := range cluster {
		if !isZeroWidth(r) {
			return false
		}
	}
	return true
}

// stripStrayTags removes tag characters from a cluster that does not begin
// with the black flag, the base of every valid tag sequence.
func stripStrayTags(cluster string) string {
	const blackFlag, firstTag, lastTag = 0x1F3F4, 0xE0000, 0xE007F
	if first, _ := utf8.DecodeRuneInString(cluster); first == blackFlag {
		return cluster
	}
	return strings.Map(func(r rune) rune {
		if r >= firstTag && r <= lastTag {
			return -1
		}
		return r
	}, cluster)
}

// clusterWidth is the number of cells one grapheme cluster occupies under
// the pinned width table, in the fixed narrow condition: ambiguous-width
// characters count as one column (NFR-8). U+FFFD is ambiguous, so one.
func clusterWidth(cluster string) int {
	if cluster == "" {
		return 0
	}
	if r, _ := utf8.DecodeRuneInString(cluster); r == utf8.RuneError {
		return 1
	}
	narrow := runewidth.Condition{EastAsianWidth: false, StrictEmojiNeutral: true}
	return narrow.StringWidth(cluster)
}

// constant is a string type no other package can name. A caller elsewhere
// can therefore hand Const an untyped string constant and nothing else: not
// a variable, not text that came from outside.
type constant string

// Const makes Text from a string constant written in the library's own
// source: its notices, its error messages. Such text needs no cleaning; it
// is the library's, and the static rules keep invisible characters out of
// the source it is written in.
func Const(s constant) Text {
	return Text{s: string(s)}
}
