package examples_test

import (
	"context"
	"fmt"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// playbackHost is the whole of a host's share of playback: the map, and
// nothing else. Every key and every Settings row goes straight to the map,
// and whatever the host shows is read back from it (L-1.13). A rules test
// holds this struct to its one field.
type playbackHost struct {
	m *tuimaps.Map
}

// key is the listener's controls.
func (h playbackHost) key(k string) error {
	switch k {
	case "space":
		if h.m.Loop().Playing {
			return h.m.Stop()
		}
		return h.m.Play()
	case "left":
		return h.m.Step(-1)
	case "right":
		return h.m.Step(1)
	case "home":
		return h.m.Reset()
	}
	return nil
}

// settings is the Settings rows: playback on or off, and its step.
func (h playbackHost) settings(on bool, step time.Duration) error {
	p := tuimaps.PlaybackOff
	if on {
		p = tuimaps.PlaybackOn
	}
	if err := h.m.SetPlayback(p); err != nil {
		return err
	}
	return h.m.SetPlaybackStep(step)
}

// status is the line a host shows under the map, read from the map.
func (h playbackHost) status() string {
	st := h.m.Loop()
	state := "stopped"
	if st.Playing {
		state = "playing"
	}
	return fmt.Sprintf("%s  frame %d of %d  %s", st.At.Format("15:04"), st.Index+1, st.Count, state)
}

// Example_playback is a host wiring playback with no state of its own. It
// renders when Changed moves or when NextCall says time has something due.
func Example_playback() {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer m.Close()
	var frames []tuimaps.LoopFrame
	for i := 3; i >= 0; i-- {
		frames = append(frames, tuimaps.LoopFrame{Valid: noon.Add(-time.Duration(i) * 5 * time.Minute), PNG: lightRainPNG()})
	}
	loop := tuimaps.RadarImage("radar", tuimaps.Image{Frames: frames, West: -90, South: 30, East: -80, North: 40,
		Projection: tuimaps.PlateCarree, Table: []tuimaps.TableEntry{{Colour: tuimaps.RGB{R: 0x04, G: 0xe9, B: 0xe7}, Value: 25}}, Exact: true}, noon)
	if _, err := m.Set(loop); err != nil {
		fmt.Println(err)
		return
	}
	size := tuimaps.Size{Cols: 80, Rows: 24}
	if _, err := m.Settle(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
	h := playbackHost{m: m}
	if err := h.settings(true, 0); err != nil {
		fmt.Println(err)
		return
	}
	render := func(wall time.Time) {
		if _, err := m.Render(size, wall); err != nil {
			fmt.Println(err)
		}
	}
	render(noon)
	fmt.Println(h.status())

	before := m.Changed()
	_ = h.key("space")
	fmt.Println("a key moved Changed:", m.Changed() != before)
	render(noon)
	fmt.Println(h.status())

	due, _ := m.NextCall(noon)
	fmt.Println("next call in", due.Sub(noon))
	render(due)
	fmt.Println(h.status())

	_ = h.key("right")
	fmt.Println(h.status())
	_ = h.key("home")
	fmt.Println(h.status())
	// Output:
	// 12:00  frame 4 of 4  stopped
	// a key moved Changed: true
	// 11:45  frame 1 of 4  playing
	// next call in 500ms
	// 11:50  frame 2 of 4  playing
	// 11:55  frame 3 of 4  stopped
	// 12:00  frame 4 of 4  stopped
}
