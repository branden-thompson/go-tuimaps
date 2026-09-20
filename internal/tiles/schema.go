package tiles

import (
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// Schema is the seam between a provider's tiles and the map (FR-35): the
// names of the layers the map draws, as this provider spells them. A change
// of schema by a provider is a new mapping here, not a new renderer.
type Schema struct {
	Name   string
	Layers []string // what the decoder keeps; every other layer is dropped while decoding
}

// openMapTiles is the one mapping in this release.
func openMapTiles() Schema {
	return Schema{Name: "openmaptiles", Layers: []string{"water", "waterway", "landcover", "park", "boundary", "transportation", "aeroway", "place", "water_name", "aerodrome_label"}}
}

// ResolveSchema picks the mapping for the layers a source declares. A source
// that declares nothing gets the default; one that declares layers of which
// the map knows none is refused.
func ResolveSchema(declared []string) (Schema, error) {
	schema := openMapTiles()
	if len(declared) == 0 {
		return schema, nil
	}
	for _, d := range declared {
		for _, l := range schema.Layers {
			if d == l {
				return schema, nil
			}
		}
	}
	return Schema{}, fault.Make(fault.UnsupportedSchema,
		textsafe.Const("the tile source was refused"),
		textsafe.Const("none of the layers it declares is one the map knows how to draw; this release reads the OpenMapTiles schema"),
		textsafe.Const("name a source that serves the OpenMapTiles schema"))
}
