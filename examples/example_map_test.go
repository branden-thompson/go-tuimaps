package examples_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
)

// noon is the moment these examples draw at, so their output never moves.
var noon = time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)

// Example_worldMap is the whole of a map of the world: create, settle,
// render. The tiles are the ones the assets package carries, so nothing is
// fetched and no connection is opened.
func Example_worldMap() {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer m.Close()
	if _, err := m.Settle(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
	frame, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(frame.Lines), "rows,", frame.Status)
	// Output: 24 rows, complete
}

// Example_alertArea hands in one feature overlay: an alert area, drawn as
// its tint with its outline over it, and labelled.
func Example_alertArea() {
	m, _ := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	defer m.Close()
	area := tuimaps.Overlay{
		ID:     "alerts",
		Valid:  noon,
		Keeps:  time.Hour,
		Credit: "National Weather Service",
		Features: []tuimaps.Feature{{
			Kind:  tuimaps.Polygon,
			Rings: [][]tuimaps.LonLat{{{Lon: -86, Lat: 24}, {Lon: -82, Lat: 24}, {Lon: -82, Lat: 28}, {Lon: -86, Lat: 28}, {Lon: -86, Lat: 24}}},
			Role:  tuimaps.AlertSevere,
			Label: "Tornado Warning",
		}},
	}
	res, err := m.Set(area)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("created:", res.Created, "· released at once:", res.Released)
	// The embedded tiles hold the world down to zoom 3; fitting closely to a
	// small area would want deeper tiles, which need a source (Source).
	if err := m.Recentre(tuimaps.LonLat{Lon: -84, Lat: 26}); err != nil {
		fmt.Println(err)
		return
	}
	if err := m.Zoom(3); err != nil {
		fmt.Println(err)
		return
	}
	if _, err := m.Settle(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
	frame, _ := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, noon)
	fmt.Println("status:", frame.Status)
	fmt.Println("credits:", len(m.Credits()))
	// Output:
	// created: true · released at once: true
	// status: complete
	// credits: 2
}

// Example_temperatureGrid is a scalar field in one call: the preset carries
// the breaks and the colours, and the host says only which unit it is in.
func Example_temperatureGrid() {
	m, _ := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	defer m.Close()
	grid := tuimaps.Grid{
		West: -100, South: 20, East: -70, North: 35,
		Cols: 2, Rows: 2,
		Values: []float64{18, 21, 24, 27},
	}
	if _, err := m.Set(tuimaps.TemperatureGrid("temperature", grid, tuimaps.Celsius, noon)); err != nil {
		fmt.Println(err)
		return
	}
	if _, err := m.Settle(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
	legend := m.Legend()
	fmt.Println("legend entries:", len(legend), "· unit:", legend[0].Unit, "· classes:", len(legend[0].Classes))
	// Output: legend entries: 1 · unit: C · classes: 17
}

// Example_places shows the host's own places: a name and a position are all
// that has to be written, and the id comes back for removing it later.
func Example_places() {
	m, _ := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	defer m.Close()
	ids, err := m.SetPlaces([]tuimaps.Place{
		{Name: "Miami", At: tuimaps.LonLat{Lon: -80.19, Lat: 25.77}},
		{ID: "home", Name: "Home", At: tuimaps.LonLat{Lon: -84.39, Lat: 33.75}, Marker: tuimaps.MarkerRing, Radius: 4},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ids[0])
	fmt.Println(ids[1])
	gone, _ := m.RemovePlace("home")
	fmt.Println("removed:", gone)
	// Output:
	// 25.770000,-80.190000
	// home
	// removed: 1
}

// Example_radarImage is a radar image with the table that says what its
// colours mean (D-45). **The provider's terms and its credit are the host's
// to honour**: this example names the United States' National Weather
// Service, whose radar imagery is public domain and asks to be credited,
// and a host using another provider must carry that provider's own terms
// and credit in its place.
func Example_radarImage() {
	m, _ := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	defer m.Close()

	// One pixel a class: the table maps the provider's own colours to the
	// classes the library draws, and every colour in the picture must be in
	// it or the hand-in says how many were not.
	image := tuimaps.Image{
		West: -100, South: 20, East: -70, North: 35,
		Projection: tuimaps.PlateCarree,
		PNG:        lightRainPNG(),
		Type:       tuimaps.Type{Preset: "radar", Unit: "dBZ"},
		Table: []tuimaps.TableEntry{
			{Colour: tuimaps.RGB{R: 0x04, G: 0xe9, B: 0xe7}, Value: 15}, // light rain, in dBZ
		},
	}
	radar := tuimaps.RadarImage("radar", image, noon)
	radar.Credit = "NOAA/NWS radar imagery, public domain"
	res, err := m.Set(radar)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("created:", res.Created)
	for _, credit := range m.Credits() {
		fmt.Println(credit)
	}
	// Output:
	// created: true
	// OpenFreeMap (c) OpenMapTiles Data from OpenStreetMap
	// NOAA/NWS radar imagery, public domain
}

// lightRainPNG is one pixel of the provider's own lightest rain colour: the
// smallest picture that shows the shape of the call.
func lightRainPNG() []byte {
	one := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	one.Set(0, 0, color.NRGBA{R: 0x04, G: 0xe9, B: 0xe7, A: 0xff})
	var out bytes.Buffer
	if err := png.Encode(&out, one); err != nil {
		panic(err)
	}
	return out.Bytes()
}
