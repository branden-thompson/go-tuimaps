package mvt

import (
	"errors"
	"math"
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

func isKind(err error, k fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == k
}

// TestVarint is plan task 03.1.
func TestVarint(t *testing.T) {
	good := []struct {
		in   []byte
		want uint64
		n    int
	}{
		{[]byte{0x00}, 0, 1},
		{[]byte{0x01}, 1, 1},
		{[]byte{0x7f}, 127, 1},
		{[]byte{0x80, 0x01}, 128, 2},
		{[]byte{0xac, 0x02, 0xff}, 300, 2}, // trailing bytes are the caller's
		{[]byte{0xff, 0xff, 0xff, 0xff, 0x0f}, math.MaxUint32, 5},
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}, math.MaxUint64, 10},
	}
	for _, c := range good {
		v, n, err := uvarint(c.in)
		if err != nil || v != c.want || n != c.n {
			t.Errorf("uvarint(% x) = %d, %d, %v; want %d, %d", c.in, v, n, err, c.want, c.n)
		}
	}
	bad := map[string][]byte{
		"empty":                {},
		"cut short":            {0x80},
		"cut short after nine": {0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		"eleven bytes":         {0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01},
		"tenth byte overflows": {0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x02},
	}
	for name, in := range bad {
		if _, _, err := uvarint(in); !isKind(err, fault.UnsupportedTile) {
			t.Errorf("%s: %v; want an unsupported-tile error", name, err)
		}
	}
}

func TestZigZag(t *testing.T) {
	for in, want := range map[uint32]int32{0: 0, 1: -1, 2: 1, 3: -2, 4: 2, 4294967294: 2147483647, 4294967295: -2147483648} {
		if got := zigzag(in); got != want {
			t.Errorf("zigzag(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestField(t *testing.T) {
	// field 3, length-delimited, three bytes of body, then field 15 as a varint.
	msg := []byte{0x1a, 0x03, 'a', 'b', 'c', 0x78, 0x02}
	f, rest, err := nextField(msg)
	if err != nil || f.num != 3 || f.wire != wireBytes || string(f.body) != "abc" {
		t.Fatalf("first field: %+v, %v", f, err)
	}
	f, rest, err = nextField(rest)
	if err != nil || f.num != 15 || f.wire != wireVarint || f.value != 2 || len(rest) != 0 {
		t.Fatalf("second field: %+v, rest %d, %v", f, len(rest), err)
	}
	// fixed-width fields are skipped over at their width.
	f, rest, err = nextField([]byte{0x15, 1, 2, 3, 4, 0x19, 1, 2, 3, 4, 5, 6, 7, 8})
	if err != nil || f.wire != wireFixed32 || len(f.body) != 4 || len(rest) != 9 {
		t.Fatalf("fixed32: %+v, rest %d, %v", f, len(rest), err)
	}
	f, rest, err = nextField(rest)
	if err != nil || f.wire != wireFixed64 || len(f.body) != 8 || len(rest) != 0 {
		t.Fatalf("fixed64: %+v, rest %d, %v", f, len(rest), err)
	}
}

// TestWirePitfalls is the wire layer's share of plan task 03.17.
func TestWirePitfalls(t *testing.T) {
	bad := map[string][]byte{
		"a group start":                          {0x0b},
		"a group end":                            {0x0c},
		"wire type 6":                            {0x0e},
		"wire type 7":                            {0x0f},
		"field number zero":                      {0x00, 0x00},
		"a length longer than what is left":      {0x1a, 0x05, 'a', 'b'},
		"a length beyond the platform's int":     {0x1a, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
		"a fixed32 cut short":                    {0x15, 1, 2},
		"a fixed64 cut short":                    {0x19, 1, 2, 3, 4},
		"a tag cut short":                        {0x80},
		"a field number beyond the format's max": {0xf8, 0xff, 0xff, 0xff, 0xff, 0x01, 0x00},
	}
	for name, in := range bad {
		if _, _, err := nextField(in); !isKind(err, fault.UnsupportedTile) {
			t.Errorf("%s: %v; want an unsupported-tile error", name, err)
		}
	}
}
