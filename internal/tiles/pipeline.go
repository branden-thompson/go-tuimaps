package tiles

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/mvt"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// embeddedIdentity is the source identity of the embedded set, and of
	// "no source at all": with no source named, the embedded set is the
	// only place a tile could come from.
	embeddedIdentity = "embedded"
	// defaultMaxZoom is the deepest zoom of a source that does not say.
	defaultMaxZoom = 14
	// firstRetry and lastRetry bound the wait after a failure (FR-23).
	firstRetry = 30 * time.Second
	lastRetry  = 10 * time.Minute
)

// Network is the source the host named. Nothing is reached until there is
// one (D-65).
type Network struct {
	// Identity is what the source is, in full; it is part of every cache key
	// and is never shown.
	Identity string
	// Get returns a tile's bytes as the source stores them.
	Get func(ctx context.Context, tile scene.TileID) ([]byte, error)
	// MinZoom and MaxZoom are the zooms the source holds. A MaxZoom of zero
	// means 14. Above MaxZoom the MaxZoom tile is drawn scaled.
	MinZoom, MaxZoom uint8
	// Zooms, if set, is asked in their place: a source that learns its zooms
	// from a document it has yet to read answers 0 to 14 until it has.
	Zooms func() (lo, hi uint8)
}

// zooms is the range the source holds, as far as it is known.
func (n *Network) zooms() (lo, hi uint8) {
	if n == nil {
		return 0, scene.MaxTileZoom
	}
	if n.Zooms != nil {
		lo, hi = n.Zooms()
		if lo <= hi && hi <= scene.MaxTileZoom {
			return lo, hi
		}
	}
	if n.MaxZoom == 0 {
		return n.MinZoom, defaultMaxZoom
	}
	return n.MinZoom, n.MaxZoom
}

// EmbeddedFunc returns an embedded tile's stored bytes, as assets.Tile does.
type EmbeddedFunc func(z uint8, x, y uint32) ([]byte, bool)

// Options is what a pipeline is made from.
type Options struct {
	Cache           *Cache       // required; may be shared between maps
	Network         *Network     // nil: no source is named
	Embedded        EmbeddedFunc // nil: no embedded tiles were passed
	EmbeddedMaxZoom uint8
	Disk            *Disk      // nil: the disk cache is off (FR-21b)
	Language        string     // the one label language kept; "en" if empty
	Limits          mvt.Limits // the zero value means the defaults
	Schema          Schema     // the zero value means the default mapping
}

// Later is a failed tile's job and the time it may be tried again.
type Later struct {
	Job       scene.Job
	NotBefore time.Time
}

// Plan is what one view needs done: jobs to queue now, ancestors first, and
// failed tiles that are waiting out their time.
type Plan struct {
	Jobs  []scene.Job
	Later []Later
}

type track struct {
	state     State
	failures  int
	notBefore time.Time // zero until the owner call after the failure stamps it
}

// Pipeline is one map's tile pipeline. Plan, Draw, State and the setters are
// owner calls; a job's Run may be on any goroutine. Its lock is never held
// across a fetch or a decode.
type Pipeline struct {
	landed   uint64 // tiles that passed the gate into the cache, for Work to see (D-66)
	mu       sync.Mutex
	opts     Options
	pins     *PinSet
	tracks   map[Key]*track
	wanted   map[string]bool // the last plan's job keys
	now      time.Time       // the host's clock at the last plan that had one; jobs date what they write by it
	failed   int
	lastFail scene.TileID
}

func refusedOptions(why textsafe.Text) error {
	return fault.Make(fault.Internal, textsafe.Const("the tile pipeline could not be made"), why, textsafe.Const("this is a defect in the library; report it"))
}

func checkNetwork(n *Network) error {
	if n == nil {
		return nil
	}
	if n.Identity == "" || n.Get == nil {
		return refusedOptions(textsafe.Const("a named source needs an identity, which is its cache key, and a way to fetch"))
	}
	if n.MaxZoom > scene.MaxTileZoom || n.MinZoom > n.MaxZoom && n.MaxZoom != 0 {
		return refusedOptions(textsafe.Const("a named source's zooms are out of order or deeper than any tile"))
	}
	return nil
}

// NewPipeline makes a pipeline. It reaches nothing.
func NewPipeline(o Options) (*Pipeline, error) {
	if o.Cache == nil {
		return nil, refusedOptions(textsafe.Const("it was given no memory cache"))
	}
	if o.EmbeddedMaxZoom > scene.MaxTileZoom {
		return nil, refusedOptions(textsafe.Const("the embedded tiles claim a zoom deeper than any tile"))
	}
	err := checkNetwork(o.Network)
	if err != nil {
		return nil, err
	}
	if o.Language == "" {
		o.Language = "en"
	}
	if o.Limits == (mvt.Limits{}) {
		o.Limits = mvt.DefaultLimits()
	}
	err = o.Limits.Validate()
	if err != nil {
		return nil, err
	}
	if len(o.Schema.Layers) == 0 {
		o.Schema = openMapTiles()
	}
	return &Pipeline{opts: o, pins: o.Cache.Register(), tracks: map[Key]*track{}, wanted: map[string]bool{}}, nil
}

// Release ends the pipeline's live view: what only it needed becomes spare.
func (p *Pipeline) Release() {
	if p == nil {
		return
	}
	p.pins.Withdraw()
}

// SetNetwork names a source, or with nil names none. Tiles that were
// unavailable are wanted again.
func (p *Pipeline) SetNetwork(n *Network) error {
	if p == nil {
		return refusedOptions(textsafe.Const("there is no pipeline"))
	}
	err := checkNetwork(n)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.opts.Network = n
	p.sourcesChangedLocked()
	return nil
}

// SetLanguage sets the one label language kept when a tile is decoded. It is
// part of a tile's cache key (D-82), so tiles already on hand are another
// language's and are wanted again.
func (p *Pipeline) SetLanguage(code string) error {
	if p == nil {
		return refusedOptions(textsafe.Const("there is no pipeline"))
	}
	if code == "" {
		code = "en"
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.opts.Language == code {
		return nil
	}
	p.opts.Language = code
	p.sourcesChangedLocked()
	return nil
}

// Purge empties the memory cache of fetched tiles. The tiles the view
// needs are fetched again, through the disk cache if it still holds them.
func (p *Pipeline) Purge() {
	if p == nil {
		return
	}
	p.mu.Lock()
	cache := p.opts.Cache
	p.mu.Unlock()
	cache.EmptyFetched()
}

// SetDisk gives the pipeline a disk cache, or with nil takes it away
// (FR-21b). What is already in memory stays there.
func (p *Pipeline) SetDisk(d *Disk) error {
	if p == nil {
		return refusedOptions(textsafe.Const("there is no pipeline"))
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.opts.Disk = d
	return nil
}

// Holding is what the tile cache in memory needs, holds and may hold (D-90).
func (p *Pipeline) Holding() Use {
	if p == nil {
		return Use{}
	}
	p.mu.Lock()
	cache := p.opts.Cache
	p.mu.Unlock()
	return cache.Bytes()
}

// SetEmbedded passes the embedded tiles, or with nil takes them away.
func (p *Pipeline) SetEmbedded(get EmbeddedFunc, maxZoom uint8) error {
	if p == nil {
		return refusedOptions(textsafe.Const("there is no pipeline"))
	}
	if maxZoom > scene.MaxTileZoom {
		return refusedOptions(textsafe.Const("the embedded tiles claim a zoom deeper than any tile"))
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.opts.Embedded, p.opts.EmbeddedMaxZoom = get, maxZoom
	p.sourcesChangedLocked()
	return nil
}

func (p *Pipeline) sourcesChangedLocked() {
	for _, tr := range p.tracks {
		tr.state, _ = Next(tr.state, SourcesChanged)
	}
}

// maxZoomLocked is the deepest zoom the chosen source holds.
func (p *Pipeline) maxZoomLocked() uint8 {
	if p.opts.Network == nil {
		return scene.MaxTileZoom
	}
	_, hi := p.opts.Network.zooms()
	return hi
}

// Landed counts the tiles that have passed the gate into the cache: what a
// Work call compares before and after, to know it landed something (D-66).
func (p *Pipeline) Landed() uint64 {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.landed
}

// Deepest is the deepest tile zoom this pipeline can actually draw at: the
// chosen source's, or the embedded tiles' when there is no source. It is what
// a host needs to know before asking for a zoom it cannot be given.
func (p *Pipeline) Deepest() uint8 {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.opts.Network != nil {
		_, hi := p.opts.Network.zooms()
		return hi
	}
	return p.opts.EmbeddedMaxZoom
}

// ancestorAt returns the tile's ancestor at zoom z, or the tile itself if it
// is no deeper.
func ancestorAt(tile scene.TileID, z uint8) scene.TileID {
	if tile.Z <= z {
		return tile
	}
	shift := tile.Z - z
	return scene.TileID{Z: z, X: tile.X >> shift, Y: tile.Y >> shift}
}

func (p *Pipeline) key(source string, tile scene.TileID) Key {
	return Key{Source: source, Language: p.opts.Language, Tile: tile}
}

// resolveLocked gives, for a wanted tile, the key of the tile itself - from
// the chosen source, at no deeper than the source goes - and the chain of
// what may be drawn for it: at each zoom the chosen source's tile, then the
// embedded one, which never overrides it (L-13).
func (p *Pipeline) resolveLocked(tile scene.TileID) (Key, []Key) {
	tile = ancestorAt(tile, p.maxZoomLocked())
	source := embeddedIdentity
	if p.opts.Network != nil {
		source = p.opts.Network.Identity
	}
	primary := p.key(source, tile)
	chain := make([]Key, 0, 2*(int(tile.Z)+1))
	for _, k := range Lineage(primary) {
		chain = append(chain, k)
		if source != embeddedIdentity && p.opts.Embedded != nil && k.Tile.Z <= p.opts.EmbeddedMaxZoom {
			chain = append(chain, p.key(embeddedIdentity, k.Tile))
		}
	}
	return primary, chain
}

// standInSourceLocked is the nearest ancestor a source can supply of a tile
// with nothing on hand to stand in for it: the embedded one if embedded
// tiles were passed, which is the cheapest job there is; else the chosen
// source's parent tile.
func (p *Pipeline) standInSourceLocked(primary Key) (Key, bool) {
	if p.opts.Embedded != nil {
		k := p.key(embeddedIdentity, ancestorAt(primary.Tile, p.opts.EmbeddedMaxZoom))
		return k, k != primary
	}
	if lo, _ := p.opts.Network.zooms(); p.opts.Network == nil || primary.Tile.Z == 0 || primary.Tile.Z <= lo {
		return Key{}, false
	}
	return p.key(primary.Source, ancestorAt(primary.Tile, primary.Tile.Z-1)), true
}

// Plan notes what a view wants, tells the cache what that view needs, and
// returns the work: jobs to queue now and failed tiles waiting out their
// time. now is the host's wall clock; nothing here reads one.
func (p *Pipeline) Plan(now time.Time, wanted []scene.TileID) Plan {
	if p == nil {
		return Plan{}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !now.IsZero() {
		p.now = now // a plan with no clock, as Settle's, keeps the last one known
	}
	for k, tr := range p.tracks {
		if tr.state == Dropped {
			delete(p.tracks, k)
		}
	}
	var first, then []Key // stand-ins to fetch, then the tiles themselves
	var chains [][]Key
	var onDisk []scene.TileID // the source's tiles the view needs, which the disk cache keeps over its cap
	seen := map[Key]bool{}
	for _, tile := range wanted {
		if tile.Validate() != nil {
			continue
		}
		primary, chain := p.resolveLocked(tile)
		if seen[primary] {
			continue
		}
		seen[primary] = true
		chains = append(chains, chain)
		then = append(then, primary)
		if p.opts.Network != nil && primary.Source == p.opts.Network.Identity {
			onDisk = append(onDisk, primary.Tile)
		}
		_, _, onHand := p.opts.Cache.First(chain)
		if onHand {
			continue
		}
		if k, ok := p.standInSourceLocked(primary); ok && !seen[k] {
			seen[k] = true
			first = append(first, k)
		}
	}
	p.pins.Publish(chains)
	if p.opts.Network != nil {
		p.opts.Disk.InView(p.opts.Network.Identity, onDisk) // memory only: a plan does no input or output
	}
	plan := Plan{}
	p.wanted = map[string]bool{}
	for _, k := range append(first, then...) {
		p.planOneLocked(now, k, &plan)
	}
	p.releaseLocked(seen)
	return plan
}

// planOneLocked moves one needed tile along and adds its work to the plan.
func (p *Pipeline) planOneLocked(now time.Time, k Key, plan *Plan) {
	if plan == nil {
		return
	}
	tr := p.trackLocked(k)
	if tr.state == Evicted {
		tr.state, _ = Next(tr.state, WantedAgain)
	}
	job := &tileJob{p: p, k: k}
	switch tr.state {
	case Wanted:
		tr.state, _ = Next(tr.state, Enqueue)
		plan.Jobs = append(plan.Jobs, job)
	case Queued:
		plan.Jobs = append(plan.Jobs, job) // the same key: the queue joins it to the first
	case Waiting:
		if tr.notBefore.IsZero() {
			tr.notBefore = now.Add(backoff(tr.failures))
		}
		if tr.notBefore.After(now) {
			plan.Later = append(plan.Later, Later{Job: job, NotBefore: tr.notBefore})
			break
		}
		tr.state, _ = Next(tr.state, RetryDue)
		plan.Jobs = append(plan.Jobs, job)
	}
	if tr.state == Queued || tr.state == Waiting || tr.state == Loading {
		p.wanted[job.Key()] = true
	}
}

// trackLocked returns a tile's record, brought up to date with the cache: a
// tile recorded as on hand that the cache no longer holds was evicted.
func (p *Pipeline) trackLocked(k Key) *track {
	tr, ok := p.tracks[k]
	if !ok {
		tr = &track{state: Wanted}
		if p.opts.Cache.Has(k) {
			tr.state = OnHand // decoded for another map that shares the cache
		}
		p.tracks[k] = tr
	}
	if tr.state == OnHand && !p.opts.Cache.Has(k) {
		tr.state, _ = Next(tr.state, Evict)
	}
	return tr
}

// releaseLocked lets go of every tile the view no longer needs.
func (p *Pipeline) releaseLocked(needed map[Key]bool) {
	if needed == nil {
		return
	}
	for k := range p.tracks {
		if needed[k] {
			continue
		}
		tr := p.trackLocked(k)
		switch tr.state {
		case Queued, Waiting:
			tr.state, _ = Next(tr.state, LeftView)
		case Wanted, Unavailable:
			delete(p.tracks, k)
		}
	}
}

// backoff is the wait after the nth failure in a row: 30 s, doubling, to 10 min.
func backoff(failures int) time.Duration {
	if failures <= 1 {
		return firstRetry
	}
	wait := firstRetry
	for range failures - 1 {
		wait *= 2
		if wait >= lastRetry {
			return lastRetry
		}
	}
	return wait
}

// StillWanted reports whether a job key was part of the last plan; the queue
// asks it which of a map's jobs to keep.
func (p *Pipeline) StillWanted(jobKey string) bool {
	if p == nil || jobKey == "" {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.wanted[jobKey]
}

// State reports where a tile is in its life, as the chosen source's tile.
func (p *Pipeline) State(tile scene.TileID) State {
	if p == nil || tile.Validate() != nil {
		return Wanted
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	primary, _ := p.resolveLocked(tile)
	if _, ok := p.tracks[primary]; !ok {
		return Wanted
	}
	return p.trackLocked(primary).state
}

// Draw returns what is drawn for a tile now: the tile itself, or the nearest
// thing on hand that stands in for it (D-30). at is the tile returned, which
// the renderer scales from; exact is false for a stand-in.
func (p *Pipeline) Draw(tile scene.TileID) (got *scene.Tile, at scene.TileID, exact, ok bool) {
	if p == nil || tile.Validate() != nil {
		return nil, scene.TileID{}, false, false
	}
	p.mu.Lock()
	primary, chain := p.resolveLocked(tile)
	p.mu.Unlock()
	got, k, ok := p.opts.Cache.First(chain)
	if !ok {
		return nil, scene.TileID{}, false, false
	}
	return got, k.Tile, k == primary, true
}

// TakeWarnings returns, once, what the tiles have to tell the host: the disk
// cache's warnings, the memory cache's need over its cap, and how many tiles
// failed since it was last called, with the last of them.
func (p *Pipeline) TakeWarnings() []fault.Warning {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	out := p.opts.Disk.TakeWarnings()
	if w, ok := p.opts.Cache.TakeWarning(); ok {
		out = append(out, w) // need alone over the cap: told once, by whichever map sharing the cache asks first
	}
	if p.failed == 0 {
		return out
	}
	w := fault.Warning{Kind: fault.TileFailed, Subject: textsafe.Clean(tileName(p.lastFail)), Count: p.failed}
	p.failed = 0
	return append(out, w)
}

func tileName(t scene.TileID) string {
	return strconv.Itoa(int(t.Z)) + "/" + strconv.FormatUint(uint64(t.X), 10) + "/" + strconv.FormatUint(uint64(t.Y), 10)
}

// tileJob fetches and decodes one tile from one source.
type tileJob struct {
	p *Pipeline
	k Key
}

// Kind says this is a tile job.
func (j *tileJob) Kind() scene.JobKind { return scene.KindTile }

// Key names the work. The source appears as a short hash: its identity may
// hold a key, and a job's name may be shown.
func (j *tileJob) Key() string {
	if j == nil {
		return ""
	}
	source := j.k.Source
	if source != embeddedIdentity {
		sum := sha256.Sum256([]byte(source))
		source = hex.EncodeToString(sum[:4])
	}
	return "tile/" + source + "/" + j.k.Language + "/" + tileName(j.k.Tile)
}

func noSource() error {
	return fault.Make(fault.FetchRefused,
		textsafe.Const("no tile could be fetched"),
		textsafe.Const("no source is named, and the embedded tiles do not hold this tile or were not passed"),
		textsafe.Const("name a source, or pass the assets package's tiles as an option"))
}

func cancelled() error {
	return fault.Make(fault.Cancelled,
		textsafe.Const("a tile was abandoned"),
		textsafe.Const("it left the view, or the work it was part of ran out of time"),
		textsafe.Const("nothing; it is asked for again if it is still wanted"))
}

func sourceFailed() error {
	return fault.Make(fault.FetchFailed,
		textsafe.Const("a tile could not be fetched"),
		textsafe.Const("the source's fetcher failed without saying why in the library's own terms"),
		textsafe.Const("check the network and the source; the tile is tried again later"))
}

// Run does the slow part: bytes from the source, through the gate, into the
// cache. It holds the pipeline's lock only to read its options and to
// publish the outcome.
func (j *tileJob) Run(ctx context.Context) error {
	if j == nil || j.p == nil || ctx == nil {
		return refusedOptions(textsafe.Const("a tile job was run with no pipeline or no context"))
	}
	p := j.p
	p.mu.Lock()
	o, now := p.opts, p.now
	tr := p.trackLocked(j.k)
	if tr.state == OnHand {
		p.mu.Unlock()
		return nil
	}
	tr.state = Loading
	p.mu.Unlock()

	tile, err := j.load(ctx, o, now)

	p.mu.Lock()
	defer p.mu.Unlock()
	tr = p.trackLocked(j.k)
	tr.state = Loading // whatever a plan made of it meanwhile, this is where the outcome applies
	switch {
	case err == nil:
		o.Cache.Put(j.k, tile)
		p.landed++
		tr.state, _ = Next(tr.state, PassedGate)
		tr.failures, tr.notBefore = 0, time.Time{}
	case isKind(err, fault.Cancelled):
		tr.state, _ = Next(tr.state, Cancelled)
	case isKind(err, fault.FetchRefused) && j.k.Source == embeddedIdentity:
		tr.state, _ = Next(tr.state, NoSource)
	default:
		tr.state, _ = Next(tr.state, SourceFailed)
		tr.failures++
		tr.notBefore = time.Time{}
		p.failed++
		p.lastFail = j.k.Tile
	}
	return err
}

func isKind(err error, kind fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == kind
}

// load gets the tile's bytes and decodes them. Every error it returns is the
// library's own.
func (j *tileJob) load(ctx context.Context, o Options, now time.Time) (*scene.Tile, error) {
	if ctx.Err() != nil {
		return nil, cancelled()
	}
	want := mvt.Want{Layers: o.Schema.Layers, Language: o.Language}
	if j.k.Source == embeddedIdentity {
		if o.Embedded == nil || j.k.Tile.Z > o.EmbeddedMaxZoom {
			return nil, noSource()
		}
		body, ok := o.Embedded(j.k.Tile.Z, j.k.Tile.X, j.k.Tile.Y)
		if !ok {
			return nil, noSource()
		}
		return mvt.Decode(body, want, o.Limits)
	}
	if o.Network == nil || o.Network.Identity != j.k.Source {
		return nil, cancelled() // the source was changed while this waited
	}
	gen := o.Disk.Generation() // a purge or a release after this, and nothing is written (L-9.3)
	o.Disk.Expire(now)
	if body, ok := o.Disk.Load(j.k.Source, j.k.Tile, now); ok {
		tile, err := mvt.Decode(body, want, o.Limits)
		if err == nil {
			return tile, nil
		}
		o.Disk.Delete(j.k.Source, j.k.Tile) // a cached file that does not decode is deleted and fetched again (FR-22a)
	}
	body, err := o.Network.Get(ctx, j.k.Tile)
	if err != nil {
		return nil, own(ctx, err)
	}
	tile, err := mvt.Decode(body, want, o.Limits)
	if err != nil {
		return nil, err
	}
	if o.Disk != nil {
		_ = o.Disk.Store(j.k.Source, j.k.Tile, body, now, gen) // only after a complete decode; a cache that cannot be written costs the tile nothing, and says so
	}
	return tile, nil
}
