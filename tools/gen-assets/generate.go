package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// maxZoom is the deepest zoom embedded: 85 tiles in all (D-33).
const maxZoom = 3

// Source returns a tile's stored bytes from the planet archive.
type Source func(ctx context.Context, id scene.TileID) ([]byte, bool, error)

// Pin names the archive the tiles were cut from. The pin, the hash list and
// the tiles change together or not at all (FR-28a).
type Pin struct {
	Archive   string // the archive's versioned name; never its address, which may hold a key
	Length    int64  // as the source reports it
	EntityTag string // as the source reports it; it changes when the file does
}

func (p Pin) text() string {
	return fmt.Sprintf("archive %s\nlength %d\nentity-tag %s\n", p.Archive, p.Length, p.EntityTag)
}

func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// Generate reads every tile of zoom 0 to maxZoom from source, strips each to
// what the library reads, and writes the tiles, the pin and the hash list
// under dir. Two runs over the same source write the same bytes.
func Generate(ctx context.Context, source Source, pin Pin, dir string) error {
	if source == nil || ctx == nil {
		return errors.New("gen-assets: no source")
	}
	if pin.Archive == "" || pin.Length <= 0 || pin.EntityTag == "" {
		return errors.New("gen-assets: the pin must name the archive, its length and its entity tag")
	}
	if dir == "" {
		return errors.New("gen-assets: no output directory")
	}
	if err := os.MkdirAll(filepath.Join(dir, "tiles"), 0o755); err != nil {
		return err
	}
	list := []string{"# pin " + sum([]byte(pin.text()))}
	for z := uint8(0); z <= maxZoom; z++ {
		for x := uint32(0); x < 1<<z; x++ {
			for y := uint32(0); y < 1<<z; y++ {
				lines, err := one(ctx, source, scene.TileID{Z: z, X: x, Y: y}, dir)
				if err != nil {
					return err
				}
				list = append(list, lines...)
			}
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "PIN"), []byte(pin.text()), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "HASHES"), []byte(strings.Join(list, "\n")+"\n"), 0o644)
}

// one strips and writes one tile, and returns its two lines of the list:
// the source tile's hash and the output tile's.
func one(ctx context.Context, source Source, id scene.TileID, dir string) ([]string, error) {
	if source == nil || ctx == nil {
		return nil, errors.New("gen-assets: no source")
	}
	if dir == "" {
		return nil, errors.New("gen-assets: no output directory")
	}
	if id.Z > maxZoom {
		return nil, errors.New("gen-assets: a tile deeper than the embedded set goes")
	}
	name := fmt.Sprintf("%d-%d-%d", id.Z, id.X, id.Y)
	stored, ok, err := source(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("gen-assets: tile %s: %w", name, err)
	}
	if !ok {
		return nil, fmt.Errorf("gen-assets: the archive has no tile %s; the embedded set must be whole", name)
	}
	kept, err := mvt.Decode(stored, drawn(), mvt.DefaultLimits())
	if err != nil {
		return nil, fmt.Errorf("gen-assets: tile %s: %w", name, err)
	}
	stripped, err := Encode(kept)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	zw, err := gzip.NewWriterLevel(&out, gzip.BestCompression) // no name, no time: the same bytes every run
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(stripped); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	rel := "tiles/" + name + ".pbf.gz"
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(rel)), out.Bytes(), 0o644); err != nil {
		return nil, err
	}
	return []string{
		fmt.Sprintf("%s  source/%s  %d", sum(stored), name, len(stored)),
		fmt.Sprintf("%s  %s  %d", sum(out.Bytes()), rel, out.Len()),
	}, nil
}

// Verify checks that the pin, the hash list and the tiles under dir agree:
// the list names the pin, every listed tile matches its hash, and no tile is
// present that the list does not name.
func Verify(dir string) error {
	if dir == "" {
		return errors.New("gen-assets: no directory")
	}
	pin, err := os.ReadFile(filepath.Join(dir, "PIN"))
	if err != nil {
		return err
	}
	list, err := os.ReadFile(filepath.Join(dir, "HASHES"))
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(list)), "\n")
	if len(lines) == 0 || lines[0] != "# pin "+sum(pin) {
		return errors.New("gen-assets: the hash list was made for another pin; regenerate the tiles")
	}
	listed := map[string]bool{}
	for _, l := range lines[1:] {
		f := strings.Fields(l)
		if len(f) != 3 || !strings.HasPrefix(f[1], "tiles/") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f[1])))
		if err != nil {
			return err
		}
		if sum(data) != f[0] {
			return fmt.Errorf("gen-assets: %s does not match the hash list", f[1])
		}
		listed[f[1]] = true
	}
	found, err := filepath.Glob(filepath.Join(dir, "tiles", "*"))
	if err != nil {
		return err
	}
	sort.Strings(found)
	for _, p := range found {
		if rel := "tiles/" + filepath.Base(p); !listed[rel] {
			return fmt.Errorf("gen-assets: %s is not in the hash list", rel)
		}
	}
	if len(listed) != 85 {
		return fmt.Errorf("gen-assets: %d tiles listed, want 85", len(listed))
	}
	return nil
}
