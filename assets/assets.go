// Package assets holds a small map of the whole world - the 85 vector tiles
// of zoom 0 to 3 - so that a map can draw with no network at all. Importing
// it changes nothing: a host passes Tile to the map as an option, and only
// then are the tiles used. They never override a source the host names.
//
// The tiles are cut from the planet file named in the pin by the generator
// in tools/gen-assets, and keep only what the library reads: the layers it
// draws, and place names in the local language and in English. The data is
// OpenStreetMap's; Notice returns the credit that must travel with it.
package assets

import (
	"embed"
	"strconv"
)

// MaxZoom is the deepest zoom embedded.
const MaxZoom = 3

//go:embed tiles PIN HASHES NOTICE
var files embed.FS

// Tile returns an embedded tile as it is stored - a gzipped vector tile - or
// false if it is not one of the 85. The bytes are the caller's own copy.
func Tile(z uint8, x, y uint32) ([]byte, bool) {
	if z > MaxZoom {
		return nil, false
	}
	if x >= 1<<z || y >= 1<<z {
		return nil, false
	}
	name := "tiles/" + strconv.Itoa(int(z)) + "-" + strconv.FormatUint(uint64(x), 10) + "-" + strconv.FormatUint(uint64(y), 10) + ".pbf.gz"
	body, err := files.ReadFile(name)
	if err != nil {
		return nil, false
	}
	return body, true
}

// text returns one of the embedded text files, or nothing if this build
// somehow lacks it; the tests hold every one of them to its content.
func text(name string) string {
	if name == "" {
		return ""
	}
	body, err := files.ReadFile(name)
	if err != nil {
		return ""
	}
	return string(body)
}

// Pin names the planet file the tiles were cut from: its versioned name, and
// the length and entity tag its source reported.
func Pin() string { return text("PIN") }

// Hashes lists the SHA-256 of every source tile and every embedded tile. It
// shows that the tiles are unchanged since they were generated, not that the
// source was authentic.
func Hashes() string { return text("HASHES") }

// Notice is the data notice: who the map data belongs to, under what terms,
// and what the hash list does and does not prove.
func Notice() string { return text("NOTICE") }
