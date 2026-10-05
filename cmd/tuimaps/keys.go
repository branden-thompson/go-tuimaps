package main

import (
	"math"
	"strings"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// The keys, by the name the app knows them by. Arrow keys arrive as escape
// sequences and are given plain names here, so that everything above this
// file deals in words and never in bytes.
const (
	keyQuit           = "quit"
	keyEscape         = "escape"
	keyZoomIn         = "zoom-in"
	keyZoomOut        = "zoom-out"
	keyLeft           = "left"
	keyRight          = "right"
	keyUp             = "up"
	keyDown           = "down"
	keyNames          = "names"
	keyWater          = "water"
	keyWorld          = "world"
	keyMarkers        = "markers"
	keyFocus          = "focus"
	keyFocusBack      = "focus-back"
	keyNow            = "right-now"
	keySafeRamps      = "safe-ramps"
	keyReduce         = "reduce-motion"
	keyColour         = "colour"
	keyDescribe       = "describe"
	keyHelp           = "help"
	keyPlay           = "play"
	keyStepBack       = "step-back"
	keyStepOn         = "step-on"
	keyUnknown        = ""
	escape       byte = 0x1b
)

// pressed is the key one character stands for, or nothing where the app has
// no use for it. It is upstream's mapping (P-68a) with the app's own
// accessibility switches added, each of which also has a flag (NFR-15).
func pressed(c byte) string {
	switch c {
	case 'q', 0x03: // q, and the interrupt a terminal sends
		return keyQuit
	case 'a', '+', '=':
		return keyZoomIn
	case 'z', 'y', '-':
		return keyZoomOut
	case 'h':
		return keyLeft
	case 'l':
		return keyRight
	case 'k':
		return keyUp
	case 'j':
		return keyDown
	case 'n':
		return keyNames
	case 'o':
		return keyWater
	case 'w':
		return keyWorld
	case 'm':
		return keyMarkers
	case '\t':
		return keyFocus
	case 's':
		return keySafeRamps
	case 'r':
		return keyReduce
	case 'C':
		return keyColour
	case 'd':
		return keyDescribe
	case '?':
		return keyHelp
	case 'p':
		return keyPlay
	case '[':
		return keyStepBack
	case ']':
		return keyStepOn
	case '0':
		return keyNow
	case escape:
		return keyEscape // Esc closes an open panel, and otherwise quits, as upstream has it
	}
	return keyUnknown
}

// decode is the keys in one read from the terminal, read on its own: a
// sequence cut off at its end is dropped.
func decode(chunk []byte) []string {
	var r keyReader
	return r.read(chunk)
}

// mostHeld is the longest escape sequence the reader holds while it waits
// for the rest. A sequence longer than any key sends is dropped.
const mostHeld = 32

// keyReader turns what is read from the terminal into key names. A terminal
// sends an arrow, a function key or Alt and a key as an escape sequence, may
// send several keys in one read, and may split a sequence across two reads,
// so each sequence is read whole and one cut off at the end of a read is
// held until the next.
type keyReader struct {
	held []byte
}

// read is the keys in one more read from the terminal.
func (r *keyReader) read(chunk []byte) []string {
	data := append(r.held, chunk...)
	r.held = nil
	var keys []string
	for i := 0; i < len(data); {
		if data[i] != escape {
			if key := pressed(data[i]); key != keyUnknown {
				keys = append(keys, key)
			}
			i++
			continue
		}
		key, size, whole := sequence(data[i:])
		if !whole {
			if len(data)-i <= mostHeld {
				r.held = append([]byte(nil), data[i:]...)
			}
			break
		}
		if key != keyUnknown {
			keys = append(keys, key)
		}
		i += size
	}
	return keys
}

// sequence is the key an escape sequence at the start of data stands for,
// how many bytes it takes, and false when data ends before it does. Esc with
// nothing after it in the read is the Esc key; Esc and a printable character
// is Alt and that key, which the app has no use for; CSI (Esc [) runs to its
// final byte, 0x40 to 0x7E, and SS3 (Esc O) is one byte more.
func sequence(data []byte) (key string, size int, whole bool) {
	if len(data) == 1 {
		return keyEscape, 1, true
	}
	switch data[1] {
	case '[':
		for i := 2; i < len(data); i++ {
			switch c := data[i]; {
			case c >= 0x40 && c <= 0x7E:
				return csiKey(data[2:i], c), i + 1, true
			case c < 0x20 || c > 0x7E:
				return keyUnknown, i, true // not a sequence after all: what came before this byte is passed over
			}
		}
		return keyUnknown, 0, false
	case 'O':
		if len(data) < 3 {
			return keyUnknown, 0, false
		}
		return cursorKey(data[2]), 3, true
	}
	if data[1] >= 0x20 && data[1] <= 0x7E {
		return keyUnknown, 2, true // Alt and a key
	}
	return keyEscape, 1, true
}

// csiKey is the key a CSI sequence stands for, by its parameters and its
// final byte. A modifier on an arrow (Esc [ 1 ; 5 C) leaves it the arrow.
func csiKey(params []byte, final byte) string {
	switch final {
	case 'Z':
		return keyFocusBack // shift and tab
	case '~':
		first, _, _ := strings.Cut(string(params), ";")
		if first == "1" || first == "7" {
			return keyNow // home, as some terminals send it
		}
		return keyUnknown
	}
	return cursorKey(final)
}

// cursorKey is the key an arrow's or home's last byte stands for.
func cursorKey(c byte) string {
	switch c {
	case 'A':
		return keyUp
	case 'B':
		return keyDown
	case 'C':
		return keyRight
	case 'D':
		return keyLeft
	case 'H':
		return keyNow // home
	}
	return keyUnknown
}

// The pan step, as upstream's (P-69): a fraction of the world that halves
// with every zoom level, so a key moves the same share of the screen
// wherever the map is.
const (
	panLon = 8.0
	panLat = 6.0
)

// step is how far one press of a pan key moves the centre, in degrees.
func step(zoom float64) (float64, float64) {
	share := math.Pow(2, zoom)
	return panLon / share, panLat / share
}

// act does what one key asks of the map. It answers whether the frame would
// differ and whether the app is finished.
func (a *app) act(key string) (redraw, done bool) {
	switch key {
	case keyQuit:
		return false, true
	case keyEscape:
		if a.describeUp || a.helpUp {
			a.describeUp, a.helpUp = false, false
			return true, false // a panel open: Esc closes it, and never quits from inside it
		}
		return false, true
	case keyZoomIn:
		return a.zoomed(1), false
	case keyZoomOut:
		return a.zoomed(-1), false
	case keyLeft, keyRight, keyUp, keyDown:
		return a.panned(key), false
	case keyWorld:
		return a.note(a.m.FitWorld()), false
	case keyFocus:
		return a.focusBy(1), false
	case keyFocusBack:
		return a.focusBy(-1), false
	case keyDescribe, keyHelp:
		return a.panel(key), false
	case keyPlay, keyStepBack, keyStepOn, keyNow:
		return a.played(key), false
	}
	return a.switched(key), false
}

// noLoop is what a loop's key says with no loop on the map.
const noLoop = "no loop is on the map; tuimaps --loop DIR opens a recorded one"

// played is the loop's keys: play and stop, a step either way, which stops
// it where it lands, and back to "right now", stopped.
func (a *app) played(key string) bool {
	if a.m.Loop().Count == 0 {
		a.said = noLoop
		return true
	}
	switch key {
	case keyPlay:
		if a.m.Loop().Playing {
			return a.note(a.m.Stop())
		}
		return a.note(a.m.Play())
	case keyStepBack:
		return a.note(a.m.Step(-1))
	case keyNow:
		if a.note(a.m.Reset()); a.said == "" {
			a.said = "right now: the newest frame, stopped"
		}
		return true
	}
	return a.note(a.m.Step(1))
}

// switched is the keys that turn something on and off: the basemap layers
// upstream switches, and the accessibility settings this app adds. Each says
// in the status row what it did.
func (a *app) switched(key string) bool {
	switch key {
	case keyNames:
		a.labels = !a.labels
		a.m.Layers(tuimaps.LabelLayer, a.labels)
		a.said = onOff("names", a.labels)
	case keyWater:
		a.water = !a.water
		a.m.Layers(tuimaps.WaterLayer, a.water)
		a.said = onOff("water", a.water)
	case keyMarkers:
		a.markers = !a.markers
		if a.note(a.showPlaces()); a.said == "" {
			a.said = onOff("markers", a.markers)
		}
	case keySafeRamps:
		a.safeRamps = !a.safeRamps
		a.m.SafeRamps(a.safeRamps)
		a.said = onOff("safe ramps", a.safeRamps)
	case keyReduce:
		a.reduceMotion = !a.reduceMotion
		a.m.ReduceMotion(a.reduceMotion)
		a.said = onOff("reduce motion", a.reduceMotion)
	case keyColour:
		a.noColour = !a.noColour
		a.m.ColourDepth(a.depth())
		a.said = onOff("colour", !a.noColour)
		if !a.noColour && a.colour == tuimaps.NoColour {
			a.said = "colour stays off while NO_COLOR is set; unset it and start again to draw in colour"
		}
	default:
		return false
	}
	return true
}

// onOff is a switch's name and where it stands.
func onOff(what string, on bool) string {
	if on {
		return what + " on"
	}
	return what + " off"
}

// depth is the colour the map is drawn in: none while the switch is on, and
// otherwise the depth the app started with, the library's own choice.
func (a *app) depth() tuimaps.Depth {
	if a.noColour {
		return tuimaps.NoColour
	}
	return a.colour
}

// startingDepth is the depth the library draws at when the app hints at
// none: no colour under a non-empty NO_COLOR, and otherwise truecolor.
func startingDepth(getenv func(string) string) tuimaps.Depth {
	if getenv("NO_COLOR") != "" {
		return tuimaps.NoColour
	}
	return tuimaps.Truecolor
}

// panel opens or closes the two things the app draws instead of the map:
// what the description says, and which key does what.
func (a *app) panel(key string) bool {
	if key == keyHelp {
		a.helpUp, a.describeUp = !a.helpUp, false
		return true
	}
	a.describeUp, a.helpUp = !a.describeUp, false
	return true
}

// zoomed zooms in or out. With a place focused it zooms about that place,
// which is what a pointer does towards what it points at - the keyboard's
// equivalent, so that nothing needs a pointer to reach (FR-5, D-17).
func (a *app) zoomed(levels float64) bool {
	at, zoom := a.m.Centre()
	focus, focused := a.focused()
	if !focused {
		return a.note(a.m.ZoomBy(levels))
	}
	// Half the distance to the place for each level in, twice it for each
	// level out: the place keeps its place on the screen and the middle
	// moves towards it.
	share := math.Pow(2, -levels)
	towards := tuimaps.LonLat{
		Lon: focus.Lon + eastward(focus.Lon, at.Lon)*share,
		Lat: focus.Lat + (at.Lat-focus.Lat)*share,
	}
	if err := a.m.Zoom(zoom + levels); err != nil {
		return a.note(err)
	}
	return a.note(a.m.Recentre(towards))
}

// panned moves the centre by upstream's own step (P-69).
func (a *app) panned(key string) bool {
	at, zoom := a.m.Centre()
	east, north := step(zoom)
	switch key {
	case keyLeft:
		at.Lon -= east
	case keyRight:
		at.Lon += east
	case keyUp:
		at.Lat += north
	case keyDown:
		at.Lat -= north
	}
	at.Lon = wrapped(at.Lon)
	at.Lat = math.Max(-85, math.Min(85, at.Lat))
	return a.note(a.m.Recentre(at))
}

// eastward is how far east one longitude is from another, the short way
// round, which is the only way a step near the seam makes sense.
func eastward(from, to float64) float64 {
	return math.Mod(to-from+540, 360) - 180
}

// wrapped is a longitude put back on the world.
func wrapped(lon float64) float64 {
	return math.Mod(lon+540, 360) - 180
}
