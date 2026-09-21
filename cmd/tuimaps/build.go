package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// errNoCacheToMaintain and errNothingToDescribe are the two things a person
// can ask for that the flags themselves rule out.
func errNoCacheToMaintain() error {
	return errors.New("there is no tile cache to work on, because --no-cache turned it off")
}

func errNothingToDescribe() error {
	return errors.New("there is nothing to describe; name a place with --place Home@-84.51,33.82, or load one with --scenario")
}

// sourceFor is where tiles come from for these settings: the app's default,
// or nowhere at all when it was told to stay offline (D-65).
func sourceFor(s settings) string {
	if s.offline {
		return ""
	}
	return defaultSource
}

// build makes the map the settings ask for. It is the one place the app
// turns what a person wrote into calls on the library, so that every mode
// starts from the same map.
func build(s settings) (*tuimaps.Map, error) {
	cols, rows := sizeOf(s)
	if cols < 1 || rows < 1 {
		return nil, fmt.Errorf("a map of %d by %d cells cannot be drawn; a map is at least 1 column by 1 row", cols, rows)
	}
	m, err := tuimaps.New(tuimaps.WithSize(cols, rows), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		return nil, err
	}
	if err := dress(m, s); err != nil {
		m.Close()
		return nil, err
	}
	return m, nil
}

// dress puts everything the settings said onto a map that has just been
// made: how it looks, where it fetches from, what it holds, and what is on
// it.
func dress(m *tuimaps.Map, s settings) error {
	m.SafeRamps(s.safeRamps)
	m.ReduceMotion(s.reduceMotion)
	if s.noColour {
		m.ColourDepth(tuimaps.NoColour)
	}
	if err := m.LabelLanguage(s.language); err != nil {
		return err
	}
	if err := useStyleFile(m, s.stylePath); err != nil {
		return err
	}
	if address := sourceFor(s); address != "" {
		if err := m.Source(address); err != nil {
			return err
		}
	}
	if s.cacheRoot != "" {
		if err := m.CacheRoot(s.cacheRoot, cacheCap); err != nil {
			return err
		}
	}
	return put(m, s)
}

// put sets the places and the overlays: the scenario's, then the person's
// own, which are added and never replace what the scenario carries.
func put(m *tuimaps.Map, s settings) error {
	places := s.places
	if s.scenario != 0 {
		loaded, err := loadScenario(s.scenario)
		if err != nil {
			return err
		}
		if err := loaded.onto(m); err != nil {
			return err
		}
		places = append(append([]tuimaps.Place(nil), loaded.places...), places...)
	}
	if len(places) == 0 {
		return nil
	}
	if _, err := m.SetPlaces(places); err != nil {
		return err
	}
	return frame(m, places)
}

// frame puts the view where what was asked about is: over the places, with
// the overlays they are to be judged against, and a margin so that nothing
// sits on the edge of the rectangle (D-76).
func frame(m *tuimaps.Map, places []tuimaps.Place) error {
	at := make([]tuimaps.LonLat, 0, len(places))
	for _, p := range places {
		at = append(at, p.At)
	}
	if err := m.FitTo(at, m.Overlays(), 2); err != nil {
		return err
	}
	return nil
}

// useStyleFile reads a style from disk and hands the bytes to the library,
// which never opens a style file itself (FR-20, D-94).
func useStyleFile(m *tuimaps.Map, path string) error {
	if path == "" {
		return nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("the style file %s could not be read: %w", path, plainly(err))
	}
	if len(body) == 0 {
		return fmt.Errorf("the style file %s is empty", path)
	}
	if err := m.SetStyle(body); err != nil {
		return fmt.Errorf("the style file %s was refused: %w", path, err)
	}
	return nil
}

// plainly is a file error with the operating system's own wording kept but
// the path taken out of it, since the app has already said which file it
// was and says it once.
func plainly(err error) error {
	var path *os.PathError
	if errors.As(err, &path) {
		return path.Err
	}
	return err
}

// settled is the map with every job run, which is what a mode that draws
// once needs before it draws (FR-30).
func settled(m *tuimaps.Map) error {
	if _, err := m.Settle(context.Background()); err != nil {
		return err
	}
	return nil
}
