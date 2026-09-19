package tiles

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/fetch"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const goodTileJSON = `{"tilejson":"3.0.0","tiles":["https://tiles.example/planet/v1/{z}/{x}/{y}.pbf"],
 "minzoom":0,"maxzoom":12,"attribution":"<a href=\"https://www.openstreetmap.org/copyright\">© OpenStreetMap</a>",
 "vector_layers":[{"id":"water"},{"id":"place"},{"id":"poi"}]}`

const source = "https://tiles.example/planet"

// TestTileJSONLimits is plan task 06.16 (NFR-10): 1 MiB, nesting 64.
func TestTileJSONLimits(t *testing.T) {
	if _, err := ParseTileJSON([]byte(goodTileJSON), source, nil); err != nil {
		t.Fatal(err)
	}
	big := `{"tiles":["https://tiles.example/{z}/{x}/{y}.pbf"],"pad":"` + strings.Repeat("x", 1<<20) + `"}`
	if _, err := ParseTileJSON([]byte(big), source, nil); !isKind(err, fault.OverLimit) {
		t.Errorf("over 1 MiB: %v", err)
	}
	deep := `{"tiles":["https://tiles.example/{z}/{x}/{y}.pbf"],"x":` + strings.Repeat("[", 65) + strings.Repeat("]", 65) + `}`
	if _, err := ParseTileJSON([]byte(deep), source, nil); !isKind(err, fault.OverLimit) {
		t.Errorf("nested 66 deep: %v", err)
	}
	quoted := `{"tiles":["https://tiles.example/{z}/{x}/{y}.pbf"],"x":"` + strings.Repeat("[", 200) + `"}`
	if _, err := ParseTileJSON([]byte(quoted), source, nil); err != nil {
		t.Errorf("brackets inside a string are not nesting: %v", err)
	}
	for name, body := range map[string]string{
		"not JSON":       `<html>`,
		"no tiles":       `{"tilejson":"3.0.0"}`,
		"empty tiles":    `{"tiles":[]}`,
		"zooms reversed": `{"tiles":["https://tiles.example/{z}/{x}/{y}.pbf"],"minzoom":9,"maxzoom":3}`,
		"zoom too deep":  `{"tiles":["https://tiles.example/{z}/{x}/{y}.pbf"],"maxzoom":40}`,
	} {
		if _, err := ParseTileJSON([]byte(body), source, nil); !isKind(err, fault.FetchRefused) {
			t.Errorf("%s: %v; want the fetch-refused kind", name, err)
		}
	}
	if _, err := ParseTileJSON([]byte(`{"tiles":["https://tiles.example/{z}/{x}/{y}.pbf"],"vector_layers":[{"id":"roads"},{"id":"earth"}]}`), source, nil); !isKind(err, fault.UnsupportedSchema) {
		t.Errorf("unknown layers: %v; want the unsupported-schema kind", err)
	}
}

// TestTileJSONAddressesObeyFetchRules is the other half of 06.16 (FR-22b).
func TestTileJSONAddressesObeyFetchRules(t *testing.T) {
	info, err := ParseTileJSON([]byte(goodTileJSON), source, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Template.URL(scene.TileID{Z: 6, X: 16, Y: 26}); got != "https://tiles.example/planet/v1/6/16/26.pbf" {
		t.Errorf("address %q", got)
	}
	if info.MinZoom != 0 || info.MaxZoom != 12 || len(info.Layers) != 3 {
		t.Errorf("%+v", info)
	}
	if s := info.Attribution.String(); strings.ContainsAny(s, "\x1b") || !strings.Contains(s, "OpenStreetMap") {
		t.Errorf("attribution %q", s)
	}
	refused := map[string]string{
		"plain http":        `http://tiles.example/{z}/{x}/{y}.pbf`,
		"a user name":       `https://user:secret@tiles.example/{z}/{x}/{y}.pbf`,
		"another host":      `https://elsewhere.example/{z}/{x}/{y}.pbf`,
		"another port":      `https://tiles.example:8443/{z}/{x}/{y}.pbf`,
		"an unknown token":  `https://tiles.example/{s}/{z}/{x}/{y}.pbf`,
		"a key token":       `https://tiles.example/{z}/{x}/{y}.pbf?key={key}`,
		"an unclosed brace": `https://tiles.example/{z}/{x}/{y`,
		"no z":              `https://tiles.example/{x}/{y}.pbf`,
		"a token in host":   `https://{z}.tiles.example/{x}/{y}.pbf`,
		"not an address":    `::::`,
		"a fragment":        `https://tiles.example/tiles#{z}/{x}/{y}`,
	}
	for name, address := range refused {
		body := `{"tiles":["` + address + `"]}`
		if _, err := ParseTileJSON([]byte(body), source, nil); !isKind(err, fault.FetchRefused) {
			t.Errorf("%s: %v; want the fetch-refused kind", name, err)
		}
	}
	allowed := `{"tiles":["https://cdn.example/{z}/{x}/{y}.pbf"]}`
	if _, err := ParseTileJSON([]byte(allowed), source, []string{"cdn.example"}); err != nil {
		t.Errorf("a host the application allowed: %v", err)
	}
	for _, err := range []error{mustFail(ParseTileJSON([]byte(goodTileJSON), "", nil)), mustFail(ParseTileJSON(nil, source, nil))} {
		if err == nil {
			t.Error("an empty source or body must be refused")
		}
	}
	// No refusal repeats the address: it may hold a key.
	_, err = ParseTileJSON([]byte(`{"tiles":["https://elsewhere.example/{z}/{x}/{y}.pbf?key=SECRET"]}`), source, nil)
	if err == nil || strings.Contains(err.Error(), "SECRET") || strings.Contains(err.Error(), "elsewhere") {
		t.Errorf("%v", err)
	}
}

func mustFail(_ TileJSON, err error) error { return err }

func FuzzTileJSON(f *testing.F) {
	f.Add([]byte(goodTileJSON))
	f.Add([]byte(`{"tiles":["https://tiles.example/{z}/{x}/{y}.pbf?a={b}"]}`))
	f.Add([]byte(`{"tiles":[1],"minzoom":"x","vector_layers":{}}`))
	f.Add([]byte(strings.Repeat("[", 100)))
	f.Fuzz(func(t *testing.T, body []byte) {
		info, err := ParseTileJSON(body, source, nil)
		if err != nil {
			var own *fault.Error
			if !errors.As(err, &own) {
				t.Fatalf("a foreign error: %v", err)
			}
			return
		}
		got := info.Template.URL(scene.TileID{Z: 1, X: 1, Y: 0})
		if !strings.HasPrefix(got, "https://tiles.example/") || strings.ContainsAny(got, "{}") {
			t.Fatalf("an accepted template built %q", got)
		}
		if info.MinZoom > info.MaxZoom || info.MaxZoom > scene.MaxTileZoom {
			t.Fatalf("zooms %d to %d", info.MinZoom, info.MaxZoom)
		}
	})
}

// fetcher stands for the library's fetcher: it records each request.
type fetcher struct {
	requests []fetch.Request
	tileJSON string
	fail     error
}

func (f *fetcher) get(_ context.Context, r fetch.Request) ([]byte, error) {
	f.requests = append(f.requests, r)
	if f.fail != nil {
		return nil, f.fail
	}
	if !strings.HasSuffix(r.URL, ".pbf") {
		return []byte(f.tileJSON), nil
	}
	body, _ := assets.Tile(1, 1, 0)
	return body, nil
}

// TestParityP49_UrlModes: a source ending in a slash is a prefix; any other
// is TileJSON, whose first template is used and fetched once.
func TestParityP49_UrlModes(t *testing.T) {
	f := &fetcher{}
	prefix, err := NewRemote("https://tiles.example/planet/", f.get, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := prefix.Get(context.Background(), id(6, 16, 26)); err != nil {
		t.Fatal(err)
	}
	if len(f.requests) != 1 || f.requests[0].URL != "https://tiles.example/planet/6/16/26.pbf" {
		t.Errorf("requests %+v", f.requests)
	}

	f = &fetcher{tileJSON: goodTileJSON}
	described, err := NewRemote(source, f.get, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, max := described.Network().Zooms(); max != 14 {
		t.Errorf("before the TileJSON is read the source is taken to reach zoom %d, want 14", max)
	}
	for _, tile := range []scene.TileID{id(6, 16, 26), id(6, 17, 26)} {
		if _, err := described.Get(context.Background(), tile); err != nil {
			t.Fatal(err)
		}
	}
	var asked []string
	for _, r := range f.requests {
		asked = append(asked, r.URL)
	}
	want := []string{source, "https://tiles.example/planet/v1/6/16/26.pbf", "https://tiles.example/planet/v1/6/17/26.pbf"}
	if strings.Join(asked, " ") != strings.Join(want, " ") {
		t.Errorf("asked %v\nwant  %v; the TileJSON is read once", asked, want)
	}
	if min, max := described.Network().Zooms(); min != 0 || max != 12 {
		t.Errorf("zooms %d to %d, want the TileJSON's 0 to 12", min, max)
	}
	if got := described.Attribution().String(); !strings.Contains(got, "OpenStreetMap") {
		t.Errorf("attribution %q", got)
	}
	if described.Network().Identity != source {
		t.Error("the identity is the source as the host named it")
	}

	// A TileJSON that is refused is refused once, and asked for no more.
	f = &fetcher{tileJSON: `{"tiles":["https://elsewhere.example/{z}/{x}/{y}.pbf"]}`}
	bad, _ := NewRemote(source, f.get, nil)
	for range 3 {
		if _, err := bad.Get(context.Background(), id(1, 1, 0)); !isKind(err, fault.FetchRefused) {
			t.Errorf("%v", err)
		}
	}
	if len(f.requests) != 1 {
		t.Errorf("%d requests; a refused TileJSON is not fetched again for every tile", len(f.requests))
	}

	for _, s := range []string{"", "ftp://tiles.example/", "https://user@tiles.example/", "https:///planet/", "tiles.example/planet/"} {
		if _, err := NewRemote(s, f.get, nil); !isKind(err, fault.FetchRefused) {
			t.Errorf("source %q: %v", s, err)
		}
	}
	if _, err := NewRemote(source, nil, nil); err == nil {
		t.Error("no fetcher must be refused")
	}
}

// TestParityP50_Http: every request carries a size limit, and the fetcher's
// failure is the tile's failure - upstream checked no status and set no limit.
func TestParityP50_Http(t *testing.T) {
	f := &fetcher{tileJSON: goodTileJSON}
	r, _ := NewRemote(source, f.get, nil)
	if _, err := r.Get(context.Background(), id(1, 1, 0)); err != nil {
		t.Fatal(err)
	}
	if f.requests[0].MaxBytes != 1<<20 || f.requests[1].MaxBytes != 2<<20 {
		t.Errorf("limits %d and %d; want 1 MiB for the TileJSON and 2 MiB for a tile", f.requests[0].MaxBytes, f.requests[1].MaxBytes)
	}
	status := fault.New(fault.FetchFailed, textsafe.Const("a request failed"), textsafe.Const("the source answered with status 503"), textsafe.Const("it is tried again later"))
	f.fail = status
	if _, err := r.Get(context.Background(), id(1, 0, 0)); !isKind(err, fault.FetchFailed) {
		t.Errorf("%v; the fetcher's own error is passed on", err)
	}
	f.fail = errors.New("dial tcp 10.0.0.1: https://tiles.example/planet?key=SECRET")
	_, err := r.Get(context.Background(), id(1, 0, 0))
	if !isKind(err, fault.FetchFailed) || strings.Contains(err.Error(), "SECRET") {
		t.Errorf("%v; a replacement fetcher's error is never passed on", err)
	}
}

// TestParityP51_FetchFailure: upstream swallowed the error and left the tile
// blank for good. Here the failure is counted, the tile waits out a time and
// is tried again, and a stand-in is drawn meanwhile.
func TestParityP51_FetchFailure(t *testing.T) {
	f := &fetcher{fail: errors.New("offline")}
	r, _ := NewRemote("https://tiles.example/planet/", f.get, nil)
	p := pipeline(t, Options{Network: r.Network(), Embedded: assets.Tile, EmbeddedMaxZoom: assets.MaxZoom})
	if failed := runAll(t, p.Plan(t0, []scene.TileID{id(6, 16, 26)})); failed != 1 {
		t.Errorf("%d jobs failed, want the one network job", failed)
	}
	if w := p.TakeWarnings(); len(w) != 1 || w[0].Count != 1 {
		t.Errorf("warnings %+v", w)
	}
	if _, _, exact, ok := p.Draw(id(6, 16, 26)); !ok || exact {
		t.Error("nothing is drawn for the failed tile; a stand-in should be")
	}
	if later := p.Plan(t0, []scene.TileID{id(6, 16, 26)}).Later; len(later) != 1 {
		t.Errorf("retries %+v", later)
	}
	f.fail = nil
	runAll(t, p.Plan(t0.Add(firstRetry), []scene.TileID{id(6, 16, 26)}))
	if _, _, exact, _ := p.Draw(id(6, 16, 26)); !exact {
		t.Error("the tile did not sharpen once the source came back")
	}
}
