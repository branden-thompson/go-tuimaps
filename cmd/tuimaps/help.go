package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// helpText is what --help prints. It carries the three things a person
// cannot find out for themselves: that the map is drawn with braille and
// their font may not have it (D-57), that this app reaches the network by
// default and how to stop it (D-65), and every accessibility switch with
// both of its ways in (PL-AX-5).
func helpText() string {
	var b strings.Builder
	b.WriteString(`tuimaps draws a map of the world in a terminal.

Usage:
  tuimaps [options]

The map is drawn with braille characters. A terminal whose font has no
braille shows boxes where the map should be. A terminal cannot be asked
what its font holds, so check your own before you judge the picture.

The network is on by default in this app: tiles are fetched from
`)
	b.WriteString(defaultSource)
	b.WriteString(`
Pass --offline to reach nothing at all and draw from the tiles built into
this program, which are the world down to zoom 3. The library itself
fetches nothing until it is told to; it is this app that tells it.

Options:
  --headless            draw one complete frame to standard output, and exit
  --describe            say in words where each place is, and draw no map
  --place NAME@LON,LAT  a place to mark and to describe; may be written again
  --scenario N          load M1 scenario N (1, 2, 3, 4, 6 or 7)
  --size COLSxROWS      the size to draw, instead of the terminal's
  --style PATH          draw the basemap by a style of your own
  --lang CODE           the language of the map's labels (default en)
  --offline             reach nothing; draw from the built-in tiles only
  --no-cache            keep no tiles on disk
  --purge               empty the tile cache on disk, and exit
  --verify              read the tile cache back, drop what is damaged, exit
  --safe-ramps          keep the library's own colours where a scale must
                        stay readable
  --reduce-motion       draw markers steadily instead of blinking
  --no-colour           draw with no colour at all
  -h, --help            print this, and exit

Accessibility. Every switch below has both a flag and a key, so that no
state of the map is reachable only one way:

`)
	for _, switched := range accessibility() {
		b.WriteString("  ")
		b.WriteString(padded(switched.what, 16))
		if switched.flag == "" {
			b.WriteString("set in the environment: a non-empty NO_COLOR means no colour\n")
			continue
		}
		b.WriteString(padded(switched.flag, 18))
		b.WriteString("key ")
		b.WriteString(switched.key)
		b.WriteString("\n")
	}
	b.WriteString(`
Keys, while the map is up:

  q, Esc      quit                     n     names on and off
  a, +        zoom in                  o     water on and off
  z, y, -     zoom out                 w     fit the whole world
  arrow keys  pan                      m     markers on and off
  h j k l     pan                      Tab   focus the next place
  ?           these keys               s r C d   the switches above

Zooming with a place focused zooms about that place, which is what a
pointer would do towards what it points at.
`)
	return b.String()
}

// padded is one column of the help, wide enough for what goes in it.
func padded(text string, width int) string {
	if width < 1 {
		return text
	}
	if len(text) >= width {
		return text + " "
	}
	return text + strings.Repeat(" ", width-len(text))
}

// saidBack is the most of a person's own words the app repeats when it
// complains about them. Long enough to recognise what was written, short
// enough that a long argument cannot fill the screen.
const saidBack = 64

// wholeComplaint is the most of a message from elsewhere - the flag
// package's own wording, which carries what was on the command line inside
// it - that the app repeats.
const wholeComplaint = 200

// printable is a piece of text made safe to print back, whoever wrote it:
// cut, with every character a terminal would act on replaced by the
// character that stands for "something was here". It is the app's half of
// the rule the library keeps on its own way out (FR-34).
func printable(text string, most int) string {
	if most < 1 {
		return ""
	}
	var b strings.Builder
	cut := false
	for i, r := range []rune(text) {
		if i >= most {
			cut = true
			break
		}
		if r == '\n' || !unhandled(r) {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(0xFFFD)
	}
	if cut {
		return b.String() + "..."
	}
	return b.String()
}

// quoted is a piece of what a person wrote, said back inside quotation
// marks so that a reader can see where it begins and ends.
func quoted(text string) string {
	return "\"" + printable(text, saidBack) + "\""
}

// clean reports whether a piece of text is one this app may put on a
// terminal: valid UTF-8, with no control character but a newline, nothing
// that moves the cursor or rewrites the line, and nothing that reorders
// what is read (FR-34's rule, applied to the app's own words).
func clean(text string) bool {
	if !utf8.ValidString(text) {
		return false
	}
	for _, r := range text {
		if r == '\n' {
			continue
		}
		if unhandled(r) {
			return false
		}
	}
	return true
}

// unhandled reports whether one character is one the app never prints.
func unhandled(r rune) bool {
	switch {
	case r < 0x20 || r == 0x7F: // C0 and delete
		return true
	case r >= 0x80 && r <= 0x9F: // C1
		return true
	case r == 0x2028 || r == 0x2029: // line and paragraph separators
		return true
	case r >= 0x202A && r <= 0x202E, r >= 0x2066 && r <= 0x2069: // bidirectional
		return true
	case r == 0x200F || r == 0x200E:
		return true
	}
	return !unicode.IsGraphic(r) && r != ' '
}
