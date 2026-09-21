// Package mvt is the library's own decoder for vector tiles. It treats every
// byte as hostile: each limit is checked before memory is set aside, layers
// and attributes the map does not draw are dropped while decoding and never
// materialised, and coordinates are kept as 16-bit integers (D-75, NFR-10).
// It decodes; it never encodes.
package mvt

import (
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// The protobuf wire types a vector tile may use. Groups (3 and 4) are
// deprecated and refused; 6 and 7 do not exist.
const (
	wireVarint  = 0
	wireFixed64 = 1
	wireBytes   = 2
	wireFixed32 = 5
)

const (
	maxVarintBytes = 10
	maxFieldNumber = 1<<29 - 1
)

// field is one field of a message: its number, its wire type, and either
// its value (a varint) or its body (the bytes of a length-delimited or
// fixed-width field). The body aliases the input; nothing is copied.
type field struct {
	num   uint32
	wire  uint8
	value uint64
	body  []byte
}

// malformed is the error for bytes that are not a well-formed vector tile.
// It names no address and quotes nothing from the tile.
func malformed() error {
	return fault.Make(fault.UnsupportedTile,
		textsafe.Const("the tile could not be decoded"),
		textsafe.Const("its bytes are not a well-formed vector tile"),
		textsafe.Const("check the tile source; if it serves another format, it is not supported"))
}

// uvarint reads one base-128 varint from the front of b and returns it with
// the number of bytes it took. A varint cut short, longer than ten bytes, or
// overflowing 64 bits is an error.
func uvarint(b []byte) (v uint64, n int, err error) {
	if len(b) == 0 {
		return 0, 0, malformed()
	}
	for i := 0; i < len(b) && i < maxVarintBytes; i++ {
		c := b[i]
		if i == maxVarintBytes-1 && c > 1 {
			return 0, 0, malformed() // the tenth byte may carry one bit only
		}
		v |= uint64(c&0x7f) << (7 * uint(i))
		if c < 0x80 {
			return v, i + 1, nil
		}
	}
	return 0, 0, malformed()
}

// zigzag undoes the zig-zag encoding of a signed 32-bit value.
func zigzag(v uint32) int32 {
	return int32(v>>1) ^ -int32(v&1)
}

// nextField reads the field at the front of msg and returns it with what
// follows. A length is checked against what is left before anything is
// sliced, so a nested length larger than its parent is an error and never
// an over-read.
func nextField(msg []byte) (f field, rest []byte, err error) {
	tag, n, err := uvarint(msg)
	if err != nil {
		return field{}, nil, err
	}
	if tag>>3 == 0 || tag>>3 > maxFieldNumber {
		return field{}, nil, malformed()
	}
	f.num, f.wire, rest = uint32(tag>>3), uint8(tag&7), msg[n:]
	var width uint64
	switch f.wire {
	case wireVarint:
		f.value, n, err = uvarint(rest)
		if err != nil {
			return field{}, nil, err
		}
		return f, rest[n:], nil
	case wireFixed64:
		width = 8
	case wireFixed32:
		width = 4
	case wireBytes:
		width, n, err = uvarint(rest)
		if err != nil {
			return field{}, nil, err
		}
		rest = rest[n:]
	default:
		return field{}, nil, malformed()
	}
	if width > uint64(len(rest)) {
		return field{}, nil, malformed()
	}
	f.body = rest[:width]
	return f, rest[width:], nil
}
