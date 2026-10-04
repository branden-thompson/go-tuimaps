package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// loopFile is a recorded radar loop as testdata/loops keeps one (v0.2.0 M1,
// plan task L10.9): its box, and its frames oldest first, each a PNG beside
// the file with its valid time. The frames are IEM's N0Q composite.
type loopFile struct {
	Name   string  `json:"name"`
	West   float64 `json:"west"`
	South  float64 `json:"south"`
	East   float64 `json:"east"`
	North  float64 `json:"north"`
	Frames []struct {
		Valid string `json:"valid"`
		File  string `json:"file"`
	} `json:"frames"`
}

// loopID is the overlay a recorded loop is drawn as.
const loopID = "loop"

// loadLoop reads a recorded loop from its directory as a radar overlay, read
// with IEM's published table.
func loadLoop(dir string) (tuimaps.Overlay, error) {
	body, err := os.ReadFile(filepath.Join(dir, "loop.json"))
	if err != nil {
		return tuimaps.Overlay{}, fmt.Errorf("the loop in %s could not be read: %w", dir, plainly(err))
	}
	var lf loopFile
	if err := json.Unmarshal(body, &lf); err != nil {
		return tuimaps.Overlay{}, fmt.Errorf("the loop in %s is not a loop file: %w", dir, plainly(err))
	}
	if len(lf.Frames) == 0 {
		return tuimaps.Overlay{}, errors.New("the loop in " + dir + " has no frames")
	}
	frames := make([]tuimaps.LoopFrame, 0, len(lf.Frames))
	for _, f := range lf.Frames {
		valid, err := time.Parse(time.RFC3339, f.Valid)
		if err != nil {
			return tuimaps.Overlay{}, fmt.Errorf("a frame of the loop in %s has no valid time: %w", dir, plainly(err))
		}
		png, err := os.ReadFile(filepath.Join(dir, filepath.Base(f.File)))
		if err != nil {
			return tuimaps.Overlay{}, fmt.Errorf("a frame of the loop in %s could not be read: %w", dir, plainly(err))
		}
		frames = append(frames, tuimaps.LoopFrame{Valid: valid, PNG: png})
	}
	newest := frames[len(frames)-1].Valid
	return tuimaps.RadarImage(loopID, tuimaps.Image{Frames: frames, West: lf.West, South: lf.South, East: lf.East, North: lf.North,
		Projection: tuimaps.PlateCarree, Provider: tuimaps.ProviderIEM}, newest), nil
}

// playLoop sets a recorded loop on the map, frames the view over it and
// starts it playing.
func playLoop(m *tuimaps.Map, dir string) error {
	o, err := loadLoop(dir)
	if err != nil {
		return err
	}
	if _, err := m.Set(o); err != nil {
		return err
	}
	if err := m.FitTo(nil, []string{loopID}, 2); err != nil {
		return err
	}
	if err := m.SetPlayback(tuimaps.PlaybackOn); err != nil {
		return err
	}
	return m.Play()
}
