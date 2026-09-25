package main

import (
	"strings"

	tuimaps "github.com/branden-thompson/go-tuimaps"
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

// The key row in three lengths. A terminal is as narrow as 69 columns
// (NFR-7), and a row of keys cut off in the middle of a word tells a person
// less than a shorter row that fits.
const (
	keysLong   = "q quit  a/z zoom  arrows pan  n names  o water  w world  m markers  Tab focus  ? keys"
	keysMiddle = "q quit  a/z zoom  arrows pan  n o w m  Tab focus  ? keys"
	keysShort  = "q quit  ? keys"
)

// keyRow is the row of keys, as upstream draws it (P-72a): what a person
// can press, in one line, in the longest form the terminal has room for.
func keyRow(cols int) string {
	for _, row := range []string{keysLong, keysMiddle, keysShort} {
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
	if panel := a.panelText(); panel != nil {
		body = panel
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

// statusRow is upstream's own: the focused marker's label after two arrows,
// and otherwise where the map is and how far in (P-57, P-72a). What went
// wrong is said here too, because it is the one line a person is already
// looking at.
func (a *app) statusRow(frame tuimaps.Frame) string {
	if a.said != "" {
		return a.said
	}
	if place, ok := a.focusedPlace(); ok {
		return ">> " + place.Name
	}
	said := a.m.Footer()
	if frame.Status == tuimaps.Complete {
		return said
	}
	if a.offline {
		// The notice stands alone: it is longer than the room beside the
		// footer, and it is what a person needs to read.
		return offlineNotice
	}
	return said + "  " + frame.Status.String()
}

// offlineNotice is why the map is not all there. The tiles built into the
// app are the world down to zoom 3, so a person who zooms to their own
// street sees an empty screen and has no way of knowing why (D-65, D-83).
const offlineNotice = "no tiles at this zoom - run without --offline to fetch them"

// panelText is what is shown instead of the map, where anything is: the
// description in words, or the keys in full.
func (a *app) panelText() []string {
	switch {
	case a.describeUp:
		return a.describedLines()
	case a.helpUp:
		return strings.Split(strings.TrimRight(helpText(), "\n"), "\n")
	}
	return nil
}

// describedLines is the description as the app would print it, shown on the
// map's own rows: the same words --describe writes, so that a person who
// reads rather than looks sees no less (FR-5, D-52).
func (a *app) describedLines() []string {
	said, err := a.m.Report(nil)
	if err != nil {
		return []string{printable(err.Error(), wholeComplaint)}
	}
	if len(said.Places) == 0 {
		return []string{"no place is named; the map has nothing to say about anywhere"}
	}
	return reported(said)
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
