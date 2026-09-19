package archive

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// TestRealArchive reads every tile of zoom 0 to 3 out of a real archive. The
// archive is 16 MB and is not committed: the test runs only when
// TUIMAPS_REAL_ARCHIVE names one, and says so when it does not.
func TestRealArchive(t *testing.T) {
	path := os.Getenv("TUIMAPS_REAL_ARCHIVE")
	if path == "" {
		t.Skip("NOT RUN: set TUIMAPS_REAL_ARCHIVE to a version 3 archive holding zoom 0 to 3")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	reads := 0
	read := func(_ context.Context, offset, length int64) ([]byte, error) {
		reads++
		b := make([]byte, length)
		_, err := io.ReadFull(io.NewSectionReader(f, offset, length), b)
		return b, err
	}
	a, err := Open(context.Background(), read, info.Size())
	if err != nil {
		t.Fatal(err)
	}
	tiles, bytes := 0, 0
	for z := uint8(0); z <= 3; z++ {
		for x := uint32(0); x < 1<<z; x++ {
			for y := uint32(0); y < 1<<z; y++ {
				data, ok, err := a.Tile(context.Background(), scene.TileID{Z: z, X: x, Y: y})
				if err != nil || !ok {
					t.Fatalf("%d/%d/%d: ok=%v, %v", z, x, y, ok, err)
				}
				if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
					t.Errorf("%d/%d/%d: the stored tile is not gzip", z, x, y)
				}
				tiles++
				bytes += len(data)
			}
		}
	}
	t.Logf("%d tiles, %d bytes stored, %d reads, header %+v", tiles, bytes, reads, a.Header())
	if tiles != 85 {
		t.Errorf("%d tiles, want 85", tiles)
	}
}
