package colour

import "testing"

// TestTheDefaultDepthFollowsTheTerminal is L-7.2 (OW-25): with no hint from
// the host and NO_COLOR unset, the depth is the terminal's own word for it -
// COLORTERM's truecolor, then TERM's - and truecolor where it says nothing;
// a hint and NO_COLOR still come first.
func TestTheDefaultDepthFollowsTheTerminal(t *testing.T) {
	for _, c := range []struct {
		colorterm, term, noColor string
		want                     Depth
	}{
		{"truecolor", "xterm-256color", "", Truecolor},
		{"24bit", "screen", "", Truecolor},
		{"", "xterm-direct", "", Truecolor},
		{"", "xterm-256color", "", Colours256},
		{"", "tmux-256color", "", Colours256},
		{"", "xterm", "", Colours16},
		{"", "linux", "", Colours16},
		{"", "dumb", "", NoColour},
		{"", "", "", Truecolor},
		{"truecolor", "xterm-256color", "1", NoColour},
	} {
		env := map[string]string{"COLORTERM": c.colorterm, "TERM": c.term, "NO_COLOR": c.noColor}
		if got := ChooseDepth(Truecolor, false, func(k string) string { return env[k] }); got != c.want {
			t.Errorf("COLORTERM %q, TERM %q, NO_COLOR %q: %v; want %v", c.colorterm, c.term, c.noColor, got, c.want)
		}
	}
	env := map[string]string{"TERM": "dumb"}
	if got := ChooseDepth(Colours256, true, func(k string) string { return env[k] }); got != Colours256 {
		t.Errorf("a hint from the host and TERM dumb: %v; the host's hint wins", got)
	}
}
