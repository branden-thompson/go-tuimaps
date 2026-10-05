package main

import (
	"strconv"
	"strings"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// chromeRows is how many rows of the terminal the app keeps for itself: a
// row of keys and a row of status, as upstream has them (P-72a). The
// library draws the exact rectangle it is given and reserves nothing of its
// own (L-17 g), so these rows are the app's business and no one else's.
const chromeRows = 2

// mapRows is how tall the map itself is on a terminal of this height.
func mapRows(rows int) int {
	if rows <= chromeRows {
		return 1 // a terminal too short for chrome still gets a map
	}
	return rows - chromeRows
}

// The key row in four lengths. A terminal is as narrow as 69 columns
// (NFR-7), and a row of keys cut off in the middle of a word tells a person
// less than a shorter row that fits. Every length names quit, play and the
// key that lists the rest.
const (
	keysLong   = "q quit  a/z zoom  arrows pan  n names  o water  w world  m markers  Tab/S-Tab focus  p play  [ ] step  0 now  d describe  ? keys"
	keysWide   = "q quit  a/z zoom  arrows pan  n o w m  Tab focus  p play  [ ] step  0 now  d describe  ? keys"
	keysMiddle = "q quit  a/z zoom  arrows pan  p play  [ ] step  d describe  ? keys"
	keysShort  = "q quit  p play  ? keys"
)

// keyRow is the row of keys, as upstream draws it (P-72a): what a person
// can press, in one line, in the longest form the terminal has room for.
func keyRow(cols int) string {
	for _, row := range []string{keysLong, keysWide, keysMiddle, keysShort} {
		if len(row) <= cols {
			return row
		}
	}
	return cut(keysShort, cols)
}

// screenful is everything the terminal shows: the map or a panel, then the
// key row, then the status row. It is always exactly as many rows as the
// terminal has, so that nothing scrolls and nothing is left behind.
func (a *app) screenful(frame tuimaps.Frame) []string {
	rows := mapRows(a.rows)
	body := frame.Lines
	if panel, whole := a.panelText(); panel != nil {
		body = fitted(panel, rows, a.cols, whole)
	}
	shown := make([]string, 0, a.rows)
	for i := range rows {
		if i < len(body) {
			shown = append(shown, body[i])
			continue
		}
		shown = append(shown, "")
	}
	if a.rows <= chromeRows {
		return shown
	}
	return append(shown, keyRow(a.cols), cut(a.statusRow(frame), a.cols))
}

// statusRow is upstream's own, where the map is and how far in (P-57,
// P-72a), with everything else a person needs to know put before it: what a
// key did or what went wrong, the focused place, the loop's state, and why
// the map is not whole. Each is added to the others and none replaces them;
// the row is cut to the terminal's width, so the most pressing comes first.
// A whole frame clears a tile failure kept from before.
func (a *app) statusRow(frame tuimaps.Frame) string {
	if frame.Status == tuimaps.Complete {
		a.failed = ""
	}
	var parts []string
	if a.said != "" {
		parts = append(parts, a.said)
	}
	if place, ok := a.focusedPlace(); ok {
		parts = append(parts, "focused: "+place.Name)
	}
	if st := a.m.Loop(); st.Count > 0 {
		parts = append(parts, loopWords(st))
	}
	if why := a.unwhole(frame.Status); why != "" {
		parts = append(parts, why)
	}
	if footer := a.m.Footer(); footer != "" {
		parts = append(parts, footer)
	}
	return strings.Join(parts, "  ")
}

// unwhole is why the map is not whole, or nothing when it is. Offline, the
// built-in map ends at its deepest zoom, and the notice is for a view past
// it or with nothing at all; online, a tile that failed is said with what to
// do; anything else is still loading.
func (a *app) unwhole(status tuimaps.Status) string {
	switch {
	case status == tuimaps.Complete:
		return ""
	case a.offline:
		if _, zoom := a.m.Centre(); status == tuimaps.NoTiles || zoom > assets.MaxZoom {
			return offlineNotice
		}
	case a.failed != "":
		return a.failed
	}
	return "loading"
}

// loopWords is where a recorded loop is and whether it moves (D-109):
// "playing" only while its frames advance, "held" while play is pressed and
// the picture is still, and with playback off, why. A forecast or a gap
// frame says so, and the time names its zone.
func loopWords(st tuimaps.LoopState) string {
	state := "stopped"
	switch {
	case st.Off == tuimaps.OffReduceMotion:
		state = "off: reduced motion"
	case st.Off == tuimaps.OffByHost:
		state = "off"
	case st.Off == tuimaps.OffByDefault:
		state = "off until turned on"
	case st.Advancing:
		state = "playing"
	case st.Playing:
		state = "held"
	}
	at := st.At.Format("15:04 MST")
	switch {
	case st.Forecast:
		at = "forecast " + at // never told as observed (D-42)
	case st.Gap:
		at = "gap " + at
	}
	return at + " frame " + strconv.Itoa(st.Index+1) + " of " + strconv.Itoa(st.Count) + " " + state
}

// offlineNotice is why the map is not all there. The tiles built into the
// app are the world down to zoom 3, so a person who zooms to their own
// street sees an empty screen and has no way of knowing why (D-65, D-83).
var offlineNotice = "no map at this zoom - the built-in map reaches zoom " + strconv.Itoa(assets.MaxZoom) +
	"; run without --offline to fetch the rest"

// panelText is what is shown instead of the map, where anything is - the
// description in words, or the key table - and the command that prints it
// whole.
func (a *app) panelText() ([]string, string) {
	switch {
	case a.describeUp:
		return a.describedLines(), "tuimaps --describe"
	case a.helpUp:
		return strings.Split(strings.TrimRight(keysText(), "\n"), "\n"), "tuimaps --help"
	}
	return nil, ""
}

// describedLines is the description as the app would print it, shown on the
// map's own rows: the same words --describe writes, so that a person who
// reads rather than looks sees no less (FR-5, D-52).
func (a *app) describedLines() []string {
	said, err := a.m.Report(nil)
	if err != nil {
		return []string{printable(err.Error(), wholeComplaint)}
	}
	if lines := reported(said, a.m.Legend()); len(lines) > 0 {
		return lines
	}
	return []string{"no place is named, and no alert or rain motion is in view"}
}

// cut is a line of the app's own, cut to the width of the terminal. The
// app's chrome is its own plain words, so a character is a cell here; the
// map's own rows are the library's and are already exact (NFR-8).
func cut(line string, cols int) string {
	if cols < 1 {
		return ""
	}
	runes := []rune(line)
	if len(runes) <= cols {
		return line
	}
	return string(runes[:cols])
}

// fitted is a panel's lines on the rows a panel has: each cut to the
// terminal's width, so that none wraps, and when there are more lines than
// rows the last row says how many are not shown and which command prints
// them all, so that nothing is dropped without a word (A11y, BUILD-exit red
// team).
func fitted(lines []string, rows, cols int, whole string) []string {
	out := make([]string, 0, min(len(lines), rows))
	for _, l := range lines {
		out = append(out, cut(l, cols))
	}
	if rows < 1 || len(out) <= rows {
		return out
	}
	hidden := len(out) - (rows - 1)
	return append(out[:rows-1], cut(strconv.Itoa(hidden)+" more lines; "+whole+" prints them all", cols))
}
