package tuimaps_test

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

// NFR-3's lines, as ruled (D-48). They cover three maps sharing caches -
// two at 149x38 and one at 69x12 - as well as one.
const (
	liveLine = 4 << 20
	peakLine = 8 << 20
)

// withoutRadar leaves the radar image out of the tour, which is how the
// image's own share of the peak was measured.
const withoutRadar = "TUIMAPS_MEMORY_WITHOUT_RADAR"

// sampleEvery is how often the heap is looked at while work runs, which is
// what NFR-3 means by peak.
const sampleEvery = 10 * time.Millisecond

// fixtureZones are the ten Gulf-coast county shapes the fixture pins.
func fixtureZones() []string {
	return []string{"FLC015", "FLC021", "FLC029", "FLC053", "FLC057", "FLC071", "FLC081", "FLC101", "FLC103", "FLC115"}
}

// alertOverlay is the fixture's alert: ten zone shapes in one overlay,
// which is what a warning set looks like on a bad day.
func alertOverlay(t *testing.T) (tuimaps.Overlay, int) {
	t.Helper()
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	var features []tuimaps.Feature
	vertices := 0
	for _, name := range fixtureZones() {
		body, err := testkit.LoadFixture(root, "zones/"+name+".json")
		if err != nil {
			t.Fatal(err)
		}
		var zone rings
		if err := json.Unmarshal(body, &zone); err != nil {
			t.Fatal(err)
		}
		shape := zone.Geometry.runs(t)
		for _, run := range shape {
			vertices += len(run)
		}
		features = append(features, tuimaps.Feature{
			Kind: tuimaps.Polygon, Role: tuimaps.AlertSevere, Label: "Coastal Flood Warning", Rings: shape})
	}
	return tuimaps.Overlay{ID: "alerts", Valid: noon, Keeps: time.Hour,
		Credit: "National Weather Service", Features: features}, vertices
}

// radarOverlay is the fixture's single radar frame with the provider's own
// colour table.
func radarOverlay(t *testing.T) tuimaps.Overlay {
	t.Helper()
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	png, err := testkit.LoadFixture(root, "radar/gulf-2026-09-19.png")
	if err != nil {
		t.Fatal(err)
	}
	body, err := testkit.LoadFixture(root, "radar/provider-colour-table.json")
	if err != nil {
		t.Fatal(err)
	}
	var entries []struct {
		DBZ *float64 `json:"dbz"`
		RGB [3]uint8 `json:"rgb"`
	}
	if err := json.Unmarshal(body, &entries); err != nil {
		t.Fatal(err)
	}
	image := tuimaps.Image{West: -88, South: 24, East: -80, North: 31,
		Projection: tuimaps.PlateCarree, PNG: png, Tolerance: 8}
	for _, e := range entries {
		entry := tuimaps.TableEntry{Colour: tuimaps.RGB{R: e.RGB[0], G: e.RGB[1], B: e.RGB[2]}, Missing: e.DBZ == nil}
		if e.DBZ != nil {
			entry.Value = *e.DBZ
		}
		image.Table = append(image.Table, entry)
	}
	return tuimaps.RadarImage("radar", image, noon)
}

// grid64x48 is the fixture's temperature field.
func grid64x48() tuimaps.Overlay {
	const cols, rows = 64, 48
	values := make([]float64, 0, cols*rows)
	for row := range rows {
		for col := range cols {
			values = append(values, 18+8*math.Sin(float64(col)/9)+4*math.Cos(float64(row)/7))
		}
	}
	return tuimaps.TemperatureGrid("temperature", tuimaps.Grid{
		West: -88, South: 24, East: -80, North: 31, Cols: cols, Rows: rows, Values: values}, tuimaps.Celsius, noon)
}

// fixtureSource is a way of fetching that serves the fixture's own tiles
// and reaches no network.
func fixtureSource(t *testing.T) http.RoundTripper {
	t.Helper()
	root, err := testkit.FixtureRoot()
	if err != nil {
		t.Fatal(err)
	}
	dirs := []string{"tiles-gulf-z6", "tiles-midwest-z5", "tiles-urban-z14"}
	return answering(func(address string) ([]byte, error) {
		z, x, y, ok := tileOf(address)
		if !ok {
			return nil, os.ErrNotExist
		}
		name := itoa(int(z)) + "-" + itoa(int(x)) + "-" + itoa(int(y)) + ".pbf"
		for _, dir := range dirs {
			if body, err := testkit.LoadFixture(root, filepath.ToSlash(filepath.Join(dir, name))); err == nil {
				return body, nil
			}
		}
		return nil, os.ErrNotExist
	})
}

// TestFixtureMemory is plan task 14.6, and NFR-3 itself: **what the library
// holds, measured against the pinned fixture under the ruled condition** -
// three maps sharing caches, two at 149x38 and one at 69x12, with two Work
// calls running at a time (D-84, D-85).
//
// Live is the heap after a forced collection, taken after a tour that fills
// every cache; peak is the largest the heap is seen to be while the work
// runs, looked at every 10 ms. Risk RS-7 has stood at High waiting for this.
func TestFixtureMemory(t *testing.T) {
	shared, err := tuimaps.NewShared(0)
	if err != nil {
		t.Fatal(err)
	}
	live, peak, held := touring(t, shared, fixtureViews())
	t.Logf("three maps sharing caches, over the fixture region: live %.2f MB, peak %.2f MB, caches %s",
		asMB(live), asMB(peak), held)
	if live > liveLine {
		t.Errorf("live is %.2f MB; NFR-3's line is 4 MB", asMB(live))
	}
	// The peak was 8.54 MB when this was first measured, over NFR-3's line
	// by 7%, and the cost was located rather than guessed at: each map
	// decoded its own copy of the same picture. Under D-116 the maps of a
	// shared set now read a picture once between them.
	if peak > peakLine {
		t.Errorf("peak is %.2f MB; NFR-3's line is 8 MB. Set %s to see what the images cost", asMB(peak), withoutRadar)
	}

	// One map, for the same reason the line covers one: a host with a
	// single map must not pay for the arrangement that serves three.
	alone, err := tuimaps.NewShared(0)
	if err != nil {
		t.Fatal(err)
	}
	oneLive, onePeak, _ := touring(t, alone, fixtureViews()[:1])
	t.Logf("one map over the fixture region: live %.2f MB, peak %.2f MB", asMB(oneLive), asMB(onePeak))
	if oneLive > liveLine || onePeak > peakLine {
		t.Errorf("one map: live %.2f MB, peak %.2f MB", asMB(oneLive), asMB(onePeak))
	}

	// Three maps on three *different* dense views is the case D-90 says is
	// recorded and not gated: it is over the line by arithmetic, and the
	// figure is what matters, not a pass.
	apart, err := tuimaps.NewShared(0)
	if err != nil {
		t.Fatal(err)
	}
	spreadLive, spreadPeak, _ := touring(t, apart, spreadViews())
	t.Logf("three maps on three different views (recorded, not gated - D-90): live %.2f MB, peak %.2f MB",
		asMB(spreadLive), asMB(spreadPeak))
}

// placed is one map's size and where it looks.
type placed struct {
	cols, rows int
	at         tuimaps.LonLat
	zoom       float64
}

// fixtureViews is the ruled arrangement: three maps over the fixture
// region, two at 149x38 and one at 69x12.
func fixtureViews() []placed {
	tampa := tuimaps.LonLat{Lon: -82.4, Lat: 27.6}
	return []placed{
		{149, 38, tampa, 6.4},
		{149, 38, tampa, 6.4},
		{69, 12, tampa, 6.4},
	}
}

// spreadViews is three maps looking at three different dense places, which
// is the case D-90 records without gating.
func spreadViews() []placed {
	return []placed{
		{149, 38, tuimaps.LonLat{Lon: -82.4, Lat: 27.6}, 6.4},
		{149, 38, tuimaps.LonLat{Lon: -86.2, Lat: 39.8}, 5},
		{69, 12, tuimaps.LonLat{Lon: 2.37, Lat: 48.86}, 6},
	}
}

// touring makes the maps, fills every cache with a scripted tour, and
// measures. It answers the live heap, the peak heap, and what the caches
// were holding at the end.
func touring(t *testing.T, shared *tuimaps.Shared, views []placed) (live, peak uint64, held string) {
	t.Helper()
	alerts, vertices := alertOverlay(t)
	radar := radarOverlay(t)
	grid := grid64x48()
	t.Logf("the fixture's alert overlay: %d vertices in %d zone shapes", vertices, len(alerts.Features))

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	stop := make(chan struct{})
	var seen atomic.Uint64
	var watching sync.WaitGroup
	watching.Add(1)
	go func() { // the sampler: the heap, every 10 ms, while the work runs
		defer watching.Done()
		for {
			select {
			case <-stop:
				return
			case <-time.After(sampleEvery):
			}
			var now runtime.MemStats
			runtime.ReadMemStats(&now)
			if now.HeapAlloc > seen.Load() {
				seen.Store(now.HeapAlloc)
			}
		}
	}()

	maps := make([]*tuimaps.Map, 0, len(views))
	for _, view := range views {
		m, err := tuimaps.New(tuimaps.WithSize(view.cols, view.rows), tuimaps.SharedCaches(shared))
		if err != nil {
			t.Fatal(err)
		}
		defer m.Close()
		useTransport(t, m, fixtureSource(t))
		if err := m.Source("https://tiles.example.test/"); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Set(alerts); err != nil {
			t.Fatal(err)
		}
		if os.Getenv(withoutRadar) == "" {
			if _, err := m.Set(radar); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := m.Set(grid); err != nil {
			t.Fatal(err)
		}
		if err := m.Recentre(view.at); err != nil {
			t.Fatal(err)
		}
		if err := m.Zoom(view.zoom); err != nil {
			t.Fatal(err)
		}
		maps = append(maps, m)
	}

	// The tour: each map moved about its own view, zoomed in and out, and
	// drawn at every step, with two Work calls running at a time (D-84).
	for _, step := range tour() {
		for i, m := range maps {
			if err := m.Zoom(views[i].zoom + step.zoom); err != nil {
				t.Fatal(err)
			}
			if err := m.PanCells(step.cols, step.rows); err != nil {
				t.Fatal(err)
			}
			pumped(t, m)
			peakAfter(&seen)
			if _, err := m.Render(tuimaps.Size{Cols: views[i].cols, Rows: views[i].rows}, noon); err != nil {
				t.Fatal(err)
			}
		}
	}
	close(stop)
	watching.Wait()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	peak = seen.Load() - min(seen.Load(), before.HeapAlloc)
	runtime.GC()
	runtime.ReadMemStats(&after)
	live = after.HeapAlloc - min(after.HeapAlloc, before.HeapAlloc)
	use := maps[0].CacheUse()
	held = "tiles " + asBytes(use.Tiles.Held) + ", shapes " + asBytes(use.Shapes.Held)
	runtime.KeepAlive(maps)
	return live, peak, held
}

// pumped runs the work two calls at a time, which is the condition the
// peak is measured under (D-84), and reads the heap after every one.
func pumped(t *testing.T, m *tuimaps.Map) {
	t.Helper()
	for range 128 {
		var running sync.WaitGroup
		did := make([]bool, 2)
		for i := range 2 {
			running.Add(1)
			go func() {
				defer running.Done()
				ran, err := m.Work(context.Background())
				if err == nil {
					did[i] = ran
				}
			}()
		}
		running.Wait()
		if !did[0] && !did[1] {
			return
		}
	}
}

// peakAfter reads the heap once more, because NFR-3 asks for it after every
// Work call as well as on the sampler's own clock.
func peakAfter(seen *atomic.Uint64) {
	var now runtime.MemStats
	runtime.ReadMemStats(&now)
	if now.HeapAlloc > seen.Load() {
		seen.Store(now.HeapAlloc)
	}
}

// step is one move of the tour.
type step struct {
	cols, rows int
	zoom       float64
}

// tour is the scripted tour that fills every cache: in, out, and around.
func tour() []step {
	return []step{
		{0, 0, 0}, {8, 0, 0}, {0, 6, 0}, {-8, 0, 0.5}, {0, -6, 1},
		{12, 4, -0.5}, {-12, -4, 0}, {0, 0, 2}, {6, 0, -1}, {0, 0, 0},
	}
}

// asMB is a byte count in megabytes, for a message a person reads.
func asMB(b uint64) float64 { return float64(b) / (1 << 20) }

// asBytes is a byte count in kilobytes.
func asBytes(b int64) string {
	return itoa(int(b/1024)) + " KB"
}
