package main

import (
	"math"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// The keys, by the name the app knows them by. Arrow keys arrive as escape
// sequences and are given plain names here, so that everything above this
// file deals in words and never in bytes.
const (
	keyQuit           = "quit"
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
	keySafeRamps      = "safe-ramps"
	keyReduce         = "reduce-motion"
	keyColour         = "colour"
	keyDescribe       = "describe"
	keyHelp           = "help"
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
	case escape:
		return keyQuit // Esc on its own quits, as upstream has it
	}
	return keyUnknown
}

// decode turns what was read from the terminal into key names. A terminal
// sends an arrow key as three bytes and may send several keys in one read,
// so the whole chunk is read through rather than one byte of it.
func decode(chunk []byte) []string {
	var keys []string
	for i := 0; i < len(chunk); i++ {
		if chunk[i] == escape && i+2 < len(chunk) && chunk[i+1] == '[' {
			if key := arrow(chunk[i+2]); key != keyUnknown {
				keys, i = append(keys, key), i+2
				continue
			}
			i += 2 // an escape sequence the app has no use for, passed over
			continue
		}
		if key := pressed(chunk[i]); key != keyUnknown {
			keys = append(keys, key)
		}
	}
	return keys
}

// arrow is the key an escape sequence's last byte stands for.
func arrow(c byte) string {
	switch c {
	case 'A':
		return keyUp
	case 'B':
		return keyDown
	case 'C':
		return keyRight
	case 'D':
		return keyLeft
	case 'Z':
		return keyFocus // shift and tab, which some terminals send this way
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
	case keyZoomIn:
		return a.zoomed(1), false
	case keyZoomOut:
		return a.zoomed(-1), false
	case keyLeft, keyRight, keyUp, keyDown:
		return a.panned(key), false
	case keyWorld:
		return a.note(a.m.FitWorld()), false
	case keyFocus:
		return a.focusNext(), false
	case keyDescribe, keyHelp:
		return a.panel(key), false
	}
	return a.switched(key), false
}

// switched is the keys that turn something on and off: the basemap layers
// upstream switches, and the accessibility settings this app adds.
func (a *app) switched(key string) bool {
	switch key {
	case keyNames:
		a.labels = !a.labels
		a.m.Layers(tuimaps.LabelLayer, a.labels)
	case keyWater:
		a.water = !a.water
		a.m.Layers(tuimaps.WaterLayer, a.water)
	case keyMarkers:
		a.markers = !a.markers
		return a.note(a.showPlaces())
	case keySafeRamps:
		a.safeRamps = !a.safeRamps
		a.m.SafeRamps(a.safeRamps)
	case keyReduce:
		a.reduceMotion = !a.reduceMotion
		a.m.ReduceMotion(a.reduceMotion)
	case keyColour:
		a.noColour = !a.noColour
		a.m.ColourDepth(a.depth())
	default:
		return false
	}
	return true
}

// depth is the colour the map is drawn in: none while the switch is on, and
// otherwise whatever the library would choose for itself.
func (a *app) depth() tuimaps.Depth {
	if a.noColour {
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
