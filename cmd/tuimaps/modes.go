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
	said, err := m.Report(nil)
	if err != nil {
		return complain(errs, err, exitFailed)
	}
	if len(said.Places) == 0 {
		return complain(errs, errNothingToDescribe(), exitMistake)
	}
	for _, line := range reported(said) {
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
		if err := m.Purge(); err != nil {
			return complain(errs, err, exitFailed)
		}
		fmt.Fprintln(out, "the tile cache at "+s.cacheRoot+" is empty")
	}
	if s.verify {
		checked, removed, err := m.Verify()
		if err != nil {
			return complain(errs, err, exitFailed)
		}
		fmt.Fprintln(out, "the tile cache at "+s.cacheRoot+" holds "+strconv.Itoa(checked)+
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

// sentence is one answer in words. The library hands out data and never a
// sentence of its own (D-52); this is the app making one, and it is the
// app's to change without the library moving.
// reported is a report as the app prints it, a place and then a line for
// each alert and each other answer: the words --describe writes and the
// describe panel shows (FR-5, D-52).
func reported(r tuimaps.Report) []string {
	lines := make([]string, 0, len(r.Places)*2)
	for _, one := range r.Places {
		lines = append(lines, one.Place)
		for _, alert := range one.Alerts {
			lines = append(lines, "  "+alert.Overlay+": "+warned(alert))
		}
		for _, answer := range one.Answers {
			lines = append(lines, "  "+answer.Overlay+": "+sentence(answer))
		}
		if len(one.Alerts) == 0 && len(one.Answers) == 0 {
			lines = append(lines, "  nothing is set over this place")
		}
	}
	return lines
}

// warned is one alert against a place, in words: what it is, how severe,
// and whether the place is inside it, near it or outside it.
func warned(a tuimaps.PlaceAlert) string {
	what := a.Label
	if what == "" {
		what = "an alert"
	}
	if word := strings.ToLower(a.Severity.Word()); word != "" {
		what += " (" + word + ")"
	}
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
	if a.Stale {
		parts = append(parts, "this data is out of date")
	}
	return strings.Join(parts, "; ")
}

func sentence(a tuimaps.Answer) string {
	parts := []string{body(a)}
	if a.UnderOneCell {
		parts = append(parts, "closer than one cell of this map, so only these words can settle it")
	}
	if a.Stale {
		parts = append(parts, "this data is out of date")
	}
	return strings.Join(parts, "; ")
}

// body is what an answer says about the place, by the shape it is.
func body(a tuimaps.Answer) string {
	if a.NoData && a.Form != "image" {
		return "no data here"
	}
	switch a.Form {
	case "area":
		return relation(a) + ", " + reach("the nearest edge is", a.Distance, a.Unit, a.Compass)
	case "points", "line":
		return named(a) + reach("is", a.Distance, a.Unit, a.Compass)
	case "field":
		return reading(a)
	case "image":
		return picture(a)
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

// reading is what a scalar field says here.
func reading(a tuimaps.Answer) string {
	said := number(rounded(a.Value))
	if a.ValueUnit != "" {
		said += " " + a.ValueUnit
	}
	said += ", band " + strconv.Itoa(a.Band)
	if a.Rises == "" {
		return said + ", with no change across this view"
	}
	return said + ", rising to the " + a.Rises
}

// picture is what a classified image says here.
func picture(a tuimaps.Answer) string {
	said := "class " + strconv.Itoa(a.Class)
	if a.NoData {
		said = "no data here"
	}
	if a.Heavier == "" {
		return said + ", with nothing heavier anywhere on it"
	}
	return said + ", " + reach("heavier", a.HeavierAt, a.Unit, a.Heavier)
}

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
