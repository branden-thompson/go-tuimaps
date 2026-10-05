package main

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// headless draws one complete frame to standard output and stops. It is
// how a map goes into a file, a pipe or a bug report, and it is how M1's
// reference frames are made at an exact size.
func headless(s settings, out, errs io.Writer) int {
	m, err := build(s)
	if err != nil {
		return complain(errs, err, exitMistake)
	}
	defer m.Close()
	if err := settled(m); err != nil {
		return complain(errs, err, exitFailed)
	}
	cols, rows := sizeOf(s)
	frame, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, time.Now())
	if err != nil {
		return complain(errs, err, exitFailed)
	}
	for _, line := range frame.Lines {
		fmt.Fprintln(out, line)
	}
	return exitFine
}

// describing prints what the map says in words and draws no map at all: no
// braille, no colour, no terminal control. It is the path of a person who
// does not read the picture (FR-5, D-52).
func describing(s settings, out, errs io.Writer) int {
	m, err := build(s)
	if err != nil {
		return complain(errs, err, exitMistake)
	}
	defer m.Close()
	if err := settled(m); err != nil {
		return complain(errs, err, exitFailed)
	}
	// One frame gives the map the wall clock, by which Report judges what is
	// out of date (D-130); the frame itself is not shown.
	cols, rows := sizeOf(s)
	if _, err := m.Render(tuimaps.Size{Cols: cols, Rows: rows}, time.Now()); err != nil {
		return complain(errs, err, exitFailed)
	}
	said, err := m.Report(nil)
	if err != nil {
		return complain(errs, err, exitFailed)
	}
	lines := reported(said, m.Legend())
	if len(lines) == 0 {
		return complain(errs, errNothingToDescribe(), exitMistake)
	}
	for _, line := range lines {
		fmt.Fprintln(out, line)
	}
	for _, warning := range m.Warnings() {
		fmt.Fprintln(errs, "tuimaps: "+noted(warning))
	}
	return exitFine
}

// noted is one of the library's warnings in a line: what it is about, what
// it was about it, and how many times, since a warning is de-duplicated and
// counted rather than repeated.
func noted(w tuimaps.Warning) string {
	said := w.Kind.String() + ": " + w.Subject.String()
	if w.Count > 1 {
		return said + " (" + strconv.Itoa(w.Count) + " times)"
	}
	return said
}

// maintain is --purge and --verify: the two things a person may do to the
// tiles this app has kept on their disk (FR-21b).
func maintain(s settings, out, errs io.Writer) int {
	m, err := build(s)
	if err != nil {
		return complain(errs, err, exitMistake)
	}
	defer m.Close()
	if s.purge {
		report, err := m.Purge()
		if err != nil {
			return complain(errs, err, exitFailed)
		}
		fmt.Fprintln(out, "the tile cache at "+printable(s.cacheRoot, wholeComplaint)+" is empty: "+strconv.Itoa(report.Removed)+" files removed")
	}
	if s.verify {
		checked, removed, err := m.Verify()
		if err != nil {
			return complain(errs, err, exitFailed)
		}
		fmt.Fprintln(out, "the tile cache at "+printable(s.cacheRoot, wholeComplaint)+" holds "+strconv.Itoa(checked)+
			" tiles; "+strconv.Itoa(removed)+" were damaged and were dropped")
	}
	return exitFine
}

// sizeOf is the size a mode draws at: what was asked for, or the
// terminal's.
func sizeOf(s settings) (int, int) {
	if s.cols > 0 && s.rows > 0 {
		return s.cols, s.rows
	}
	return terminalSize()
}

// reported is a report as the app prints it, a place and then a line for
// each alert and each other answer: the words --describe writes and the
// describe panel shows (FR-5, D-52). An image's class and a field's band are
// said as the legend's range (L5.7, D-47).
func reported(r tuimaps.Report, legend []tuimaps.LegendEntry) []string {
	lines := make([]string, 0, len(r.Places)*2)
	for _, one := range r.Places {
		lines = append(lines, one.Place)
		for _, alert := range one.Alerts {
			lines = append(lines, "  "+warned(alert))
		}
		for _, answer := range one.Answers {
			lines = append(lines, "  "+answer.Overlay+": "+sentence(answer, legend))
		}
		if len(one.Alerts) == 0 && len(one.Answers) == 0 {
			lines = append(lines, "  nothing is set over this place")
		}
	}
	if len(r.Alerts) > 0 {
		lines = append(lines, "in view")
		for _, a := range r.Alerts {
			lines = append(lines, "  "+inView(a))
		}
	}
	for _, mo := range r.Motion {
		lines = append(lines, moved(mo))
	}
	return lines
}

// inView is one alert in view, named with no place needed (L-13.5), with
// its times.
func inView(a tuimaps.AlertShown) string {
	what := alertName(a.Overlay, a.Label, a.Severity)
	if times := alertTimes(a.Valid, a.Expires); times != "" {
		what += ", " + times
	}
	if a.Stale {
		what += " (out of date)"
	}
	return what
}

// alertName is an alert as both lists name it: its overlay, its label where
// the label says more than the overlay's name, and its severity in the same
// word every time.
func alertName(overlay, label string, severity tuimaps.Severity) string {
	what := overlay
	if label != "" && label != overlay {
		what += ": " + label
	}
	if word := strings.ToLower(severity.Word()); word != "" {
		what += ", " + word
	}
	return what
}

// alertTimes is when an alert's data was valid and when it expires, each
// where it is given, with the zone named.
func alertTimes(valid, expires time.Time) string {
	said := ""
	if !valid.IsZero() {
		said = "valid from " + valid.Format("15:04 MST")
	}
	if expires.IsZero() {
		return said
	}
	if said == "" {
		return "until " + expires.Format("15:04 MST")
	}
	return said + " until " + expires.Format("15:04 MST")
}

// moved is a loop's observed motion in words (L-1.12, D-122): the way the
// heavier rain near the place moved and how fast, over the loop, then where
// the nearest of it is now and whether it came closer. With no motion to
// tell, it says why (D-111). Observation, never forecast.
func moved(mo tuimaps.MotionReport) string {
	near := mo.Place
	if near == "" {
		near = "the view's centre"
	}
	old := ""
	if mo.Stale {
		old = "; this data is out of date" // never told as now (D-130)
	}
	if mo.Missing != 0 {
		return mo.Overlay + ": no motion to tell near " + near + ": " + mo.Missing.String() + old
	}
	over := "over the " + strconv.Itoa(int(mo.Span.Round(time.Minute).Minutes())) + " minutes to " + mo.To.Valid.Format("15:04 MST")
	how := "has barely moved " + over
	if mo.Moving {
		how = "is moving " + mo.HeadingCompass + " at about " + strconv.Itoa(int(math.Round(mo.SpeedKmh/5))*5) + " kilometres an hour, " + over
	}
	said := mo.Overlay + ": heavier rain near " + near + " " + how + "; "
	distance := strconv.FormatFloat(mo.To.Distance, 'f', 0, 64)
	if distance == "0" && mo.Stale {
		return said + "it was over " + near + " at " + mo.To.Valid.Format("15:04 MST") + old
	}
	if distance == "0" {
		return said + "it is over " + near + " now"
	}
	went := map[tuimaps.Trend]string{tuimaps.Closer: "coming closer", tuimaps.Away: "moving away", tuimaps.Held: "holding its distance"}[mo.Trend]
	unit := mo.To.Unit
	if distance == "1" {
		unit = strings.TrimSuffix(unit, "s") // one kilometre, one mile
	}
	return said + "the nearest is " + distance + " " + unit + " " + mo.To.Compass + ", " + went + old
}

// warned is one alert against a place, in words: what it is, how severe,
// whether the place is inside it, near it or outside it, and when its data
// was valid.
func warned(a tuimaps.PlaceAlert) string {
	what := alertName(a.Overlay, a.Label, a.Severity)
	where := "outside this area"
	switch a.Where {
	case tuimaps.Inside:
		where = "inside this area"
	case tuimaps.Nearby:
		where = "near this area, outside it"
	}
	parts := []string{what + ": " + where + ", " + reach("the nearest edge is", a.Distance, a.Unit, a.Compass)}
	if a.UnderOneCell {
		parts = append(parts, "closer than one cell of this map, so only these words can settle it")
	}
	if times := alertTimes(a.Valid, time.Time{}); times != "" {
		parts = append(parts, times)
	}
	if a.Stale {
		parts = append(parts, "this data is out of date")
	}
	return strings.Join(parts, "; ")
}

// sentence is one answer in words. The library hands out data and never a
// sentence of its own (D-52); this is the app making one, and it is the
// app's to change without the library moving.
func sentence(a tuimaps.Answer, legend []tuimaps.LegendEntry) string {
	var entry *tuimaps.LegendEntry
	for i := range legend {
		if legend[i].ID == a.Overlay {
			entry = &legend[i]
		}
	}
	parts := []string{body(a, entry)}
	if a.UnderOneCell {
		parts = append(parts, "closer than one cell of this map, so only these words can settle it")
	}
	if a.Stale {
		parts = append(parts, "this data is out of date")
	}
	return strings.Join(parts, "; ")
}

// body is what an answer says about the place, by the shape it is, with the
// overlay's legend entry where it has one.
func body(a tuimaps.Answer, entry *tuimaps.LegendEntry) string {
	if a.NoData && a.Form != "image" {
		return "no reading here"
	}
	switch a.Form {
	case "area":
		return relation(a) + ", " + reach("the nearest edge is", a.Distance, a.Unit, a.Compass)
	case "points", "line":
		return named(a) + reach("is", a.Distance, a.Unit, a.Compass)
	case "field":
		return reading(a, entry)
	case "image":
		return picture(a, entry)
	}
	return "nothing is known about this"
}

// relation is whether the place is in the area or out of it.
func relation(a tuimaps.Answer) string {
	if a.Relation.String() == "inside" {
		return "inside this area"
	}
	return "outside this area"
}

// named is what the nearest thing is called, where it is called anything of
// its own. A label that only repeats the overlay's name says nothing twice,
// so the sentence says where the nearest part of it is instead.
func named(a tuimaps.Answer) string {
	if a.Label == "" || a.Label == a.Overlay {
		return "the nearest part of it "
	}
	return a.Label + " "
}

// reach is a distance and a direction in words a voice can read.
func reach(lead string, distance float64, unit, compass string) string {
	if unit == "" {
		return lead + " an unknown distance away"
	}
	said := lead + " " + number(rounded(distance)) + " " + unit
	if compass == "" {
		return said
	}
	return said + " to the " + compass
}

// reading is what a field says here: its value, the legend's range it falls
// in, and which way it rises. Wind says where it blows from instead, since a
// wind answer carries no rise.
func reading(a tuimaps.Answer, entry *tuimaps.LegendEntry) string {
	said := number(rounded(a.Value))
	if a.ValueUnit != "" {
		said += " " + a.ValueUnit
	}
	if a.From != "" {
		said = "from the " + a.From + " at " + said
	}
	if within, ok := legendRange(entry, a.Band); ok {
		said += ", in the legend's " + within
	} else {
		said += ", band " + strconv.Itoa(a.Band)
	}
	switch {
	case a.From != "":
		return said
	case a.Rises == "":
		return said + ", with no change across this view"
	}
	return said + ", rising to the " + a.Rises
}

// picture is what a classified image says here: the legend's range for its
// class, and how far the nearest heavier reading is. Radar's readings are
// rain.
func picture(a tuimaps.Answer, entry *tuimaps.LegendEntry) string {
	what := "reading"
	if entry != nil && entry.Preset == "radar" {
		what = "rain"
	}
	said, ok := legendRange(entry, a.Class)
	switch {
	case a.NoData:
		said = "no reading here"
	case !ok:
		said = "class " + strconv.Itoa(a.Class)
	}
	switch {
	case a.Heavier == "" && a.NoData:
		return said + ", with no " + what + " anywhere on it"
	case a.Heavier == "":
		return said + ", with no heavier " + what + " anywhere on it"
	case a.NoData:
		return said + ", " + reach("the nearest "+what+" is", a.HeavierAt, a.Unit, a.Heavier)
	}
	return said + ", " + reach("the nearest "+what+" heavier than here is", a.HeavierAt, a.Unit, a.Heavier)
}

// legendRange is the range a class covers, as the legend words it, in the
// overlay's own unit: "10 to 20 dBZ", "under 10 °C", "20 mph and above".
// False where the overlay has no legend or the class is not on it.
func legendRange(entry *tuimaps.LegendEntry, class int) (string, bool) {
	if entry == nil || class < 0 || class >= len(entry.Classes) {
		return "", false
	}
	label, unit := entry.Classes[class].Label, entry.Unit
	if spoken, ok := unitWords[unit]; ok {
		unit = spoken
	}
	if unit == "" || label == "all values" {
		return label, true
	}
	if above, ok := strings.CutSuffix(label, " and above"); ok {
		return above + " " + unit + " and above", true
	}
	return label + " " + unit, true
}

// unitWords are the legend's units a voice would misread, in words.
var unitWords = map[string]string{"C": "degrees Celsius", "F": "degrees Fahrenheit"}

// rounded is a number said to a length a voice can carry: whole units above
// ten, one decimal place below. "Twelve point three seven kilometres" has
// said nothing more than "twelve kilometres".
func rounded(v float64) float64 {
	if math.Abs(v) >= 10 {
		return math.Round(v)
	}
	return math.Round(v*10) / 10
}

// number is a number written out, with no trailing zero of its own.
func number(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
