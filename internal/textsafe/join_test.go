package textsafe

import "testing"

// TestJoin: cleaned pieces joined are still clean. A piece cannot smuggle
// anything in by leaning on its neighbour.
func TestJoin(t *testing.T) {
	got := Join(Const("the source "), Quote("https://tiles.example"), Const(" refused the request"))
	if got.String() != "the source https://tiles.example refused the request" {
		t.Errorf("%q", got)
	}
	if Join().String() != "" || Join(Const("one")).String() != "one" {
		t.Error("joining nothing gives nothing; joining one piece gives that piece")
	}
	// The last piece ends in a right-to-left override, written as its UTF-8 bytes.
	hostile := Join(Const("a"), Clean("\x1b[2J"), Clean("b\xe2\x80\xae"))
	for _, r := range hostile.String() {
		if isControl(r) || isBidiControl(r) {
			t.Errorf("%q holds %U", hostile, r)
		}
	}
	if again := Clean(got.String()); again != got {
		t.Errorf("a joined text is not already clean: %q became %q", got, again)
	}
}
