package scene

import (
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
	"testing"
	"unsafe"
)

func TestByteAccountingMatchesTheRealSizes(t *testing.T) {
	if got := int(unsafe.Sizeof(Tile{})); got > tileBytes {
		t.Errorf("a Tile is %d bytes; the accounting says %d", got, tileBytes)
	}
	if got := int(unsafe.Sizeof(Layer{})); got > layerBytes {
		t.Errorf("a Layer is %d bytes; the accounting says %d", got, layerBytes)
	}
	if got := int(unsafe.Sizeof(Feature{})); got > featureBytes {
		t.Errorf("a Feature is %d bytes; the accounting says %d", got, featureBytes)
	}
}

func TestTileBytes(t *testing.T) {
	var none *Tile
	if none.Bytes() != 0 {
		t.Error("a nil tile costs nothing")
	}
	tile := &Tile{Layers: []Layer{{
		Name:     "water",
		Coords:   make([]int16, 10),
		Parts:    make([]uint32, 2),
		Features: []Feature{{Class: "lake", Name: textsafe.Clean("Erie")}, {Class: "lake"}, {Class: "river", Name: textsafe.Clean("Maumee")}},
	}}}
	want := tileBytes + 20 + 8 + layerBytes + len("water") + 3*featureBytes + len("Erie") + len("Maumee") + len("lake") + len("river")
	if got := tile.Bytes(); got != want {
		t.Errorf("Bytes() = %d, want %d: a class shared by neighbouring features is counted once", got, want)
	}
}

func TestLayerPart(t *testing.T) {
	l := Layer{Coords: []int16{1, 2, 3, 4, 5, 6, 7, 8}, Parts: []uint32{2, 8}}
	first, err := l.Part(0)
	if err != nil || len(first) != 2 || first[0] != 1 {
		t.Errorf("part 0: %v, %v", first, err)
	}
	second, err := l.Part(1)
	if err != nil || len(second) != 6 || second[0] != 3 || second[5] != 8 {
		t.Errorf("part 1: %v, %v", second, err)
	}
	if _, err := l.Part(2); err == nil {
		t.Error("a part past the end must be an error")
	}
	for name, bad := range map[string]Layer{
		"an end before its start": {Coords: make([]int16, 8), Parts: []uint32{6, 2}},
		"an end past the slab":    {Coords: make([]int16, 4), Parts: []uint32{6}},
		"an odd number of values": {Coords: make([]int16, 8), Parts: []uint32{3}},
	} {
		if _, err := bad.Part(len(bad.Parts) - 1); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
