package textsafe

// InGlyphList reports whether r is one of the characters the renderer may
// emit of its own accord. The list is closed (NFR-8): each member measures
// one cell under the pinned width table, and each was seen, at the right
// width, on the terminal test card of the PLAN entry checks. Text that
// comes from outside - names, labels, credits - is not held to this list;
// it is cleaned and measured instead.
//
// For v0.1.0 the list is: the space and printable ASCII, for the library's
// own words and numbers; braille, the map itself; six marker glyphs; eight
// box-drawing characters for the scale mark and frames; three hatch
// strokes and three block shades for the forms drawn with no colour; and
// U+FFFD, which cleaning emits. Arrows and block quadrants are on the test
// card and join the list with wind and the block renderer.
func InGlyphList(r rune) bool {
	if r >= 0x20 && r <= 0x7E { // the space and printable ASCII
		return true
	}
	if r >= 0x2800 && r <= 0x28FF { // braille patterns
		return true
	}
	switch r {
	case 0x00B7, 0x2022, 0x25C6, 0x25C9, 0x25CB, 0x25CF: // markers: middle dot, bullet, diamond, fisheye, circles
		return true
	case 0x2500, 0x2502, 0x250C, 0x2510, 0x2514, 0x2518, 0x251C, 0x2524: // box-drawing, light
		return true
	case 0x2571, 0x2572, 0x2573: // hatch strokes
		return true
	case 0x2591, 0x2592, 0x2593: // block shades
		return true
	case 0xFFFD: // the replacement character
		return true
	}
	return false
}
