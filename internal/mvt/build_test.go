package mvt

// A small vector-tile writer for tests. The library never encodes; this
// lives in test files only.

func putVarint(b []byte, v uint64) []byte {
	for v >= 0x80 {
		b = append(b, byte(v)|0x80)
		v >>= 7
	}
	return append(b, byte(v))
}

func putBytes(b []byte, num uint32, body []byte) []byte {
	b = putVarint(b, uint64(num)<<3|wireBytes)
	b = putVarint(b, uint64(len(body)))
	return append(b, body...)
}

func putUint(b []byte, num uint32, v uint64) []byte {
	b = putVarint(b, uint64(num)<<3|wireVarint)
	return putVarint(b, v)
}

func packed(vals []uint32) []byte {
	var b []byte
	for _, v := range vals {
		b = putVarint(b, uint64(v))
	}
	return b
}

// testFeature builds a Feature message: type, tag pairs, geometry integers.
func testFeature(kind uint32, tags, geometry []uint32) []byte {
	var b []byte
	if len(tags) > 0 {
		b = putBytes(b, 2, packed(tags))
	}
	b = putUint(b, 3, uint64(kind))
	return putBytes(b, 4, packed(geometry))
}

// stringValue builds a Value message holding a string.
func stringValue(s string) []byte { return putBytes(nil, 1, []byte(s)) }

// intValue builds a Value message holding an int64.
func intValue(v uint64) []byte { return putUint(nil, 4, v) }

// testLayer builds a Layer message. An extent of 0 leaves the field out, so
// the decoder must apply the default.
func testLayer(name string, extent uint32, keys []string, values [][]byte, features ...[]byte) []byte {
	b := putUint(nil, 15, 2)
	b = putBytes(b, 1, []byte(name))
	for _, f := range features {
		b = putBytes(b, 2, f)
	}
	for _, k := range keys {
		b = putBytes(b, 3, []byte(k))
	}
	for _, v := range values {
		b = putBytes(b, 4, v)
	}
	if extent != 0 {
		b = putUint(b, 5, uint64(extent))
	}
	return b
}

// testTile wraps layers into a Tile message.
func testTile(layers ...[]byte) []byte {
	var b []byte
	for _, l := range layers {
		b = putBytes(b, 3, l)
	}
	return b
}

// zz zig-zag encodes a signed delta.
func zz(v int32) uint32 { return uint32(v<<1) ^ uint32(v>>31) }

// Geometry command integers.
func moveTo(n uint32) uint32 { return 1 | n<<3 }
func lineTo(n uint32) uint32 { return 2 | n<<3 }
func closePath() uint32      { return 7 | 1<<3 }
