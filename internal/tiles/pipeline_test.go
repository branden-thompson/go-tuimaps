package tiles

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

var t0 = time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)

func id(z uint8, x, y uint32) scene.TileID { return scene.TileID{Z: z, X: x, Y: y} }

// network is a named source that serves real tile bytes and records what it
// was asked for. Any tile's bytes will do for any tile: the decoder does not
// know where a tile belongs.
type network struct {
	asked []scene.TileID
	fail  error
}

func (n *network) get(_ context.Context, tile scene.TileID) ([]byte, error) {
	n.asked = append(n.asked, tile)
	if n.fail != nil {
		return nil, n.fail
	}
	body, _ := assets.Tile(2, 1, 1)
	return body, nil
}

func (n *network) source() *Network {
	return &Network{Identity: "https://tiles.example/planet", Get: n.get, MaxZoom: 14}
}

func pipeline(t *testing.T, o Options) *Pipeline {
	t.Helper()
	if o.Cache == nil {
		o.Cache = mustCache(t, 4<<20)
	}
	p, err := NewPipeline(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

// runAll runs every job of a plan, as a host's Work calls would.
func runAll(t *testing.T, plan Plan) (failed int) {
	t.Helper()
	for _, job := range plan.Jobs {
		if job.Kind() != scene.KindTile {
			t.Errorf("%s is of kind %d", job.Key(), job.Kind())
		}
		if err := job.Run(context.Background()); err != nil {
			failed++
		}
	}
	return failed
}

// TestNoSourceNoConnection is plan task 06.1 (D-65): nothing is registered by
// default, so a map with no named source has nothing to dial with.
func TestNoSourceNoConnection(t *testing.T) {
	p := pipeline(t, Options{})
	plan := p.Plan(t0, []scene.TileID{id(6, 16, 26)})
	if len(plan.Jobs) != 1 {
		t.Fatalf("%d jobs; the one job is what finds that there is no source", len(plan.Jobs))
	}
	err := plan.Jobs[0].Run(context.Background())
	if !isKind(err, fault.FetchRefused) || !strings.Contains(err.Error(), "no source is named") {
		t.Errorf("%v; the job must say why nothing could be fetched", err)
	}
	if got := p.State(id(6, 16, 26)); got != Unavailable {
		t.Errorf("state %v, want unavailable", got)
	}
	again := p.Plan(t0.Add(time.Hour), []scene.TileID{id(6, 16, 26)})
	if len(again.Jobs) != 0 || len(again.Later) != 0 {
		t.Errorf("%d jobs and %d retries for an unavailable tile; it is never retried on a timer", len(again.Jobs), len(again.Later))
	}
}

// TestSourceOrder is plan task 06.2, and TestAssetsNeverOverrideChosenSource
// is 06.20 (L-13): the named source is asked for the tile itself; the
// embedded set only ever supplies what stands in until it arrives.
func TestSourceOrder(t *testing.T) {
	net := &network{}
	p := pipeline(t, Options{Network: net.source(), Embedded: assets.Tile, EmbeddedMaxZoom: assets.MaxZoom})
	plan := p.Plan(t0, []scene.TileID{id(2, 1, 1)})
	if len(plan.Jobs) != 2 || !strings.Contains(plan.Jobs[0].Key(), "embedded") {
		t.Fatalf("jobs %v; the embedded tile is the cheapest job and sorts first", keys(plan.Jobs))
	}
	if err := plan.Jobs[0].Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, at, exact, ok := p.Draw(id(2, 1, 1)); !ok || exact || at != id(2, 1, 1) {
		t.Errorf("at %v exact=%v ok=%v; the embedded tile is drawn, as a stand-in, until the network's arrives", at, exact, ok)
	}
	if len(net.asked) != 0 {
		t.Error("the embedded job asked the network")
	}
	if err := plan.Jobs[1].Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, _, exact, _ := p.Draw(id(2, 1, 1)); !exact {
		t.Error("the network's tile arrived and did not win")
	}
	if len(net.asked) != 1 || net.asked[0] != id(2, 1, 1) {
		t.Errorf("the network was asked for %v", net.asked)
	}
}

func TestAssetsNeverOverrideChosenSource(t *testing.T) {
	net := &network{fail: errors.New("offline")}
	p := pipeline(t, Options{Network: net.source(), Embedded: assets.Tile, EmbeddedMaxZoom: assets.MaxZoom})
	runAll(t, p.Plan(t0, []scene.TileID{id(2, 1, 1), id(6, 16, 26)}))
	if got := p.State(id(2, 1, 1)); got != Waiting {
		t.Errorf("state %v; an embedded tile on hand does not settle a tile the chosen source failed to give", got)
	}
	if _, at, exact, ok := p.Draw(id(6, 16, 26)); !ok || exact || at != id(3, 2, 3) {
		t.Errorf("zoom 6 drawn from %v exact=%v ok=%v; want the embedded zoom-3 ancestor as a stand-in", at, exact, ok)
	}

	// With no source named, the embedded set is the source for the zooms it holds.
	alone := pipeline(t, Options{Embedded: assets.Tile, EmbeddedMaxZoom: assets.MaxZoom})
	runAll(t, alone.Plan(t0, []scene.TileID{id(2, 1, 1), id(6, 16, 26)}))
	if _, _, exact, ok := alone.Draw(id(2, 1, 1)); !ok || !exact {
		t.Error("with no source named, an embedded tile is the tile")
	}
	if got := alone.State(id(6, 16, 26)); got != Unavailable {
		t.Errorf("zoom 6 with no source: state %v, want unavailable", got)
	}
	if _, at, _, ok := alone.Draw(id(6, 16, 26)); !ok || at != id(3, 2, 3) {
		t.Errorf("zoom 6 drawn from %v; want the embedded ancestor", at)
	}
}

func keys(jobs []scene.Job) []string {
	var out []string
	for _, j := range jobs {
		out = append(out, j.Key())
	}
	return out
}

// TestAncestorsAreWanted is plan task 06.19: with nothing on hand to stand
// in, a missing tile also wants the nearest ancestor a source can supply,
// and that job sorts first.
func TestAncestorsAreWanted(t *testing.T) {
	net := &network{}
	p := pipeline(t, Options{Network: net.source()})
	plan := p.Plan(t0, []scene.TileID{id(6, 16, 26), id(6, 17, 26)})
	if len(plan.Jobs) != 3 {
		t.Fatalf("jobs %v; want the shared parent once, then the two tiles", keys(plan.Jobs))
	}
	if !strings.HasSuffix(plan.Jobs[0].Key(), "/5/8/13") {
		t.Errorf("first job %s; the ancestor sorts first", plan.Jobs[0].Key())
	}
	runAll(t, Plan{Jobs: plan.Jobs[:1]})
	again := p.Plan(t0, []scene.TileID{id(6, 16, 26), id(6, 17, 26), id(6, 16, 27)})
	for _, k := range keys(again.Jobs) {
		if strings.Contains(k, "/5/") || strings.Contains(k, "/4/") {
			t.Errorf("%s wanted though a stand-in is already on hand", k)
		}
	}
	world := pipeline(t, Options{Network: net.source()})
	if jobs := world.Plan(t0, []scene.TileID{id(0, 0, 0)}).Jobs; len(jobs) != 1 {
		t.Errorf("jobs %v; the world tile has no ancestor", keys(jobs))
	}
}

// TestDedupInFlight is plan task 06.15: two wants for one tile make one job,
// and a tile already being worked on makes none.
func TestDedupInFlight(t *testing.T) {
	net := &network{}
	p := pipeline(t, Options{Network: net.source()})
	plan := p.Plan(t0, []scene.TileID{id(0, 0, 0), id(0, 0, 0)})
	if len(plan.Jobs) != 1 {
		t.Fatalf("jobs %v", keys(plan.Jobs))
	}
	if again := p.Plan(t0, []scene.TileID{id(0, 0, 0)}); len(again.Jobs) != 1 || again.Jobs[0].Key() != plan.Jobs[0].Key() {
		t.Errorf("jobs %v; a queued tile is offered again under the same key, which the queue joins to the first", keys(again.Jobs))
	}
	runAll(t, plan)
	if again := p.Plan(t0, []scene.TileID{id(0, 0, 0)}); len(again.Jobs) != 0 {
		t.Errorf("jobs %v for a tile on hand", keys(again.Jobs))
	}
	if len(net.asked) != 1 {
		t.Errorf("the network was asked %d times", len(net.asked))
	}
}

// TestNotBeforeDoubles is plan task 06.7 (FR-23): 30 s doubling to 10 min.
// The time is stamped at the owner call after the failure, from the clock
// the host passes in; nothing here reads a clock or makes a timer.
func TestNotBeforeDoubles(t *testing.T) {
	net := &network{fail: errors.New("offline")}
	p := pipeline(t, Options{Network: net.source()})
	now := t0
	want := []time.Duration{30 * time.Second, time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute, 10 * time.Minute, 10 * time.Minute}
	plan := p.Plan(now, []scene.TileID{id(0, 0, 0)})
	for i, delay := range want {
		if len(plan.Jobs) != 1 {
			t.Fatalf("round %d: jobs %v", i, keys(plan.Jobs))
		}
		if runAll(t, plan) != 1 {
			t.Fatalf("round %d: the job did not fail", i)
		}
		plan = p.Plan(now, []scene.TileID{id(0, 0, 0)})
		if len(plan.Jobs) != 0 || len(plan.Later) != 1 || !plan.Later[0].NotBefore.Equal(now.Add(delay)) {
			t.Fatalf("round %d: %d jobs, retries %+v; want one retry %v later", i, len(plan.Jobs), plan.Later, delay)
		}
		if early := p.Plan(now.Add(delay-time.Second), []scene.TileID{id(0, 0, 0)}); len(early.Jobs) != 0 {
			t.Fatalf("round %d: retried a second early", i)
		}
		now = now.Add(delay)
		plan = p.Plan(now, []scene.TileID{id(0, 0, 0)})
	}
	net.fail = nil
	runAll(t, plan)
	if got := p.State(id(0, 0, 0)); got != OnHand {
		t.Errorf("state %v after the source came back", got)
	}
}

// TestTileStates is plan task 06.6: every transition of the state diagram
// (L3 States, machine 1), and no other.
func TestTileStates(t *testing.T) {
	allowed := map[State]map[Event]State{
		Wanted:      {Enqueue: Queued},
		Queued:      {LeftView: Dropped, CapDropped: Dropped, PickedUp: Loading},
		Loading:     {Cancelled: Dropped, PassedGate: OnHand, NoSource: Unavailable, SourceFailed: Waiting},
		Waiting:     {RetryDue: Queued, LeftView: Dropped},
		Unavailable: {SourcesChanged: Wanted},
		OnHand:      {Evict: Evicted},
		Evicted:     {WantedAgain: Wanted},
		Dropped:     {},
	}
	for from := Wanted; from <= Dropped; from++ {
		for ev := Enqueue; ev <= WantedAgain; ev++ {
			got, ok := Next(from, ev)
			want, legal := allowed[from][ev]
			if ok != legal || (legal && got != want) {
				t.Errorf("%v on %v: %v, %v; want %v, %v", from, ev, got, ok, want, legal)
			}
			if !ok && got != from {
				t.Errorf("%v on %v: an event that does not apply moved the state to %v", from, ev, got)
			}
		}
	}

	// The pipeline walks the same machine.
	net := &network{}
	p := pipeline(t, Options{Network: net.source(), Cache: mustCache(t, 1)}) // a cache too small to keep a spare
	if got := p.State(id(1, 0, 0)); got != Wanted {
		t.Errorf("a tile never asked about: %v", got)
	}
	plan := p.Plan(t0, []scene.TileID{id(1, 0, 0)})
	if got := p.State(id(1, 0, 0)); got != Queued {
		t.Errorf("after the plan: %v", got)
	}
	runAll(t, plan)
	if got := p.State(id(1, 0, 0)); got != OnHand {
		t.Errorf("after the job: %v", got)
	}
	if got := p.State(id(0, 0, 0)); got != Evicted {
		t.Errorf("the world tile stood in until the tile arrived, and is a spare since: %v, want evicted", got)
	}
	// The view moves away: nothing needs the tile, and the cache is over its cap.
	p.Plan(t0, []scene.TileID{id(1, 1, 1)})
	if got := p.State(id(1, 0, 0)); got != Evicted {
		t.Errorf("on hand, unneeded and over the cap: %v, want evicted (D-90)", got)
	}
	if got := p.State(id(0, 0, 0)); got != Queued {
		t.Errorf("the stand-in for the new view, evicted and wanted again: %v, want queued", got)
	}
	if got := p.State(id(1, 1, 1)); got != Queued {
		t.Errorf("the new view's tile: %v", got)
	}
	back := p.Plan(t0, []scene.TileID{id(1, 0, 0)})
	if got := p.State(id(1, 0, 0)); got != Queued || len(back.Jobs) == 0 {
		t.Errorf("wanted again: %v with %d jobs", got, len(back.Jobs))
	}
	if got := p.State(id(1, 1, 1)); got != Dropped {
		t.Errorf("queued, then left the view: %v, want dropped", got)
	}
	if !p.StillWanted(back.Jobs[0].Key()) || p.StillWanted("tile/other/en/1/1/1") {
		t.Error("StillWanted must answer for the last plan's jobs, and for no others")
	}

	// Cancelled while loading: dropped, and not a failure to retry.
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	q := pipeline(t, Options{Network: &Network{Identity: "https://slow.example", MaxZoom: 14, Get: func(ctx context.Context, _ scene.TileID) ([]byte, error) {
		return nil, ctx.Err()
	}}})
	plan = q.Plan(t0, []scene.TileID{id(0, 0, 0)})
	if err := plan.Jobs[0].Run(cancelled); !isKind(err, fault.Cancelled) {
		t.Errorf("%v; want the cancelled kind", err)
	}
	if got := q.State(id(0, 0, 0)); got != Dropped {
		t.Errorf("cancelled while loading: %v, want dropped", got)
	}
	if later := q.Plan(t0, nil).Later; len(later) != 0 {
		t.Errorf("a cancelled tile was given a retry time: %+v", later)
	}
}

// TestSourcesChanged: an unavailable tile is wanted again once a source is
// named or embedded tiles are passed.
func TestSourcesChanged(t *testing.T) {
	p := pipeline(t, Options{})
	runAll(t, p.Plan(t0, []scene.TileID{id(2, 1, 1)}))
	if got := p.State(id(2, 1, 1)); got != Unavailable {
		t.Fatalf("state %v", got)
	}
	if err := p.SetEmbedded(assets.Tile, assets.MaxZoom); err != nil {
		t.Fatal(err)
	}
	plan := p.Plan(t0, []scene.TileID{id(2, 1, 1)})
	if runAll(t, plan) != 0 || p.State(id(2, 1, 1)) != OnHand {
		t.Errorf("after embedded tiles were passed: %v with jobs %v", p.State(id(2, 1, 1)), keys(plan.Jobs))
	}
	net := &network{}
	if err := p.SetNetwork(net.source()); err != nil {
		t.Fatal(err)
	}
	if got := p.State(id(2, 1, 1)); got != Wanted {
		t.Errorf("a source was named: state %v; the chosen source's tile is now the one wanted", got)
	}
}

// TestOverzoomAboveSourceMax is plan task 06.18 (parity P-18): above the
// source's deepest zoom its deepest tile is what is asked for and drawn.
func TestOverzoomAboveSourceMax(t *testing.T) {
	net := &network{}
	p := pipeline(t, Options{Network: net.source()})
	deep := id(16, 4*8299+1, 4*5636+2)
	plan := p.Plan(t0, []scene.TileID{deep, id(16, 4*8299+2, 4*5636+2)})
	var asked []string
	for _, k := range keys(plan.Jobs) {
		if strings.HasSuffix(k, "/14/8299/5636") {
			asked = append(asked, k)
		}
		if strings.Contains(k, "/16/") || strings.Contains(k, "/15/") {
			t.Errorf("%s asked of a source that stops at zoom 14", k)
		}
	}
	if len(asked) != 1 {
		t.Errorf("jobs %v; both tiles are one zoom-14 tile, asked for once", keys(plan.Jobs))
	}
	runAll(t, plan)
	if _, at, exact, ok := p.Draw(deep); !ok || !exact || at != id(14, 8299, 5636) {
		t.Errorf("drawn from %v exact=%v ok=%v; the zoom-14 tile, scaled, is the tile", at, exact, ok)
	}
}

// TestSchemaMappingSeam is plan task 06.14 (D-46, FR-35).
func TestSchemaMappingSeam(t *testing.T) {
	declared := []string{"water", "waterway", "landcover", "landuse", "park", "boundary", "transportation", "place", "poi"}
	schema, err := ResolveSchema(declared)
	if err != nil || schema.Name != "openmaptiles" {
		t.Fatalf("%+v, %v", schema, err)
	}
	for _, l := range schema.Layers {
		if l == "poi" || l == "landuse" {
			t.Errorf("%s is kept, though the map does not draw it", l)
		}
	}
	if s, err := ResolveSchema(nil); err != nil || s.Name != "openmaptiles" {
		t.Errorf("a source that declares nothing: %+v, %v; the default mapping applies", s, err)
	}
	if _, err := ResolveSchema([]string{"roads", "buildings", "earth"}); !isKind(err, fault.UnsupportedSchema) {
		t.Errorf("unknown layers: %v; want the unsupported-schema kind", err)
	}
}

// TestBadTileFromTheSource: bytes that do not pass the gate are a failure of
// the source, retried like any other.
func TestBadTileFromTheSource(t *testing.T) {
	p := pipeline(t, Options{Network: &Network{Identity: "https://bad.example", MaxZoom: 14, Get: func(context.Context, scene.TileID) ([]byte, error) {
		return []byte("<html>not a tile</html>"), nil
	}}})
	if runAll(t, p.Plan(t0, []scene.TileID{id(0, 0, 0)})) != 1 {
		t.Error("bytes that are not a tile did not fail the job")
	}
	if got := p.State(id(0, 0, 0)); got != Waiting {
		t.Errorf("state %v, want waiting", got)
	}
	if w := p.TakeWarnings(); len(w) != 1 || w[0].Kind != fault.TileFailed {
		t.Errorf("warnings %+v; want one tile-failed", w)
	}
}

func TestPipelineRefusals(t *testing.T) {
	if _, err := NewPipeline(Options{Cache: mustCache(t, 1000), Network: &Network{Identity: "", Get: (&network{}).get}}); err == nil {
		t.Error("a source with no identity must be refused: the identity is the cache key")
	}
	if _, err := NewPipeline(Options{Cache: mustCache(t, 1000), Network: &Network{Identity: "x"}}); err == nil {
		t.Error("a source with nothing to fetch with must be refused")
	}
	if _, err := NewPipeline(Options{}); err == nil {
		t.Error("no cache must be refused")
	}
	p := pipeline(t, Options{})
	if plan := p.Plan(t0, []scene.TileID{{Z: 40}}); len(plan.Jobs) != 0 {
		t.Error("a tile deeper than any zoom was planned")
	}
}
