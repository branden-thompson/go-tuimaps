package overlay

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// defaultOverlayVertices and defaultStoreVertices are the vertex caps: one
	// overlay's, and all of a map's together (NFR-20). A host may lower them.
	defaultOverlayVertices = 2_000_000
	defaultStoreVertices   = 4_000_000
	// maxWarnings is how many distinct warnings are kept until they are taken.
	maxWarnings = 64
)

// FeatureKind is what a feature's geometry is.
type FeatureKind uint8

// The kinds of feature (FR-6).
const (
	Point FeatureKind = iota + 1
	Line
	Polygon
	Circle
)

// Feature is ONE AREA of a feature overlay: a polygon's outer ring and then
// its holes, a line's one run of positions, a point's one position. A circle
// has a centre and a radius and no rings.
//
// **The first ring is the outline and every ring after it is a hole**, so
// ground that is separate is a separate feature. Two features that overlap are
// two fills, and their overlap is inside (FR-11); two areas poured into one
// feature would make the second a hole in the first.
//
// Its rings are read where they are and not copied again (FR-11): once a
// feature is handed in, the library holds these slices and a host that edits
// them is editing what is drawn. A host whose geometry is in its OWN point
// type converts it once with tuimaps.Rings, which copies at that point and
// hands over memory nothing else is holding.
type Feature struct {
	Kind     FeatureKind
	Rings    [][]project.LonLat
	Centre   project.LonLat
	RadiusKm float64
	Role     colour.Token // the token it is drawn in; for an alert, its outline's
	Label    string
	// Severity is an alert's, as data; zero means the one its role implies
	// (L-13.9). Valid and Expires are its times, zero when not given.
	Severity       Severity
	Valid, Expires time.Time
}

// Severity is how severe an alert is. The zero value means "as the role
// says": a feature drawn in an alert's colours has that alert's severity.
type Severity uint8

// The severities, least first.
const (
	SeverityUnknown Severity = iota + 1
	SeverityMinor
	SeverityModerate
	SeveritySevere
	SeverityExtreme
)

// Word is a severity as the label says it (L-8.1), and Digit as the outline
// repeats it (D-65): extreme 4, severe 3, moderate 2, minor 1, unknown ?.
func (s Severity) Word() string {
	return [...]string{"", "UNKNOWN", "MINOR", "MODERATE", "SEVERE", "EXTREME"}[min(int(s), 5)]
}

// Digit is the severity's digit, or empty for no severity.
func (s Severity) Digit() string {
	return [...]string{"", "?", "1", "2", "3", "4"}[min(int(s), 5)]
}

// LabelOf is what an alert's label says on the map: the host's label and the
// severity as a word, "Tornado Warning · EXTREME", or the word alone when
// there is no label (L-8.1). Any other feature's label is its own.
func LabelOf(f Feature) string {
	word := SeverityOf(f).Word()
	switch {
	case word == "":
		return f.Label
	case f.Label == "":
		return word
	}
	return f.Label + " · " + word
}

// SeverityOf is a feature's severity: the one it states, or the one its role
// implies; zero for a feature that is no alert.
func SeverityOf(f Feature) Severity {
	if f.Severity != 0 {
		return f.Severity
	}
	switch f.Role {
	case colour.AlertExtremeOutline, colour.AlertExtremeTint:
		return SeverityExtreme
	case colour.AlertSevereOutline, colour.AlertSevereTint:
		return SeveritySevere
	case colour.AlertModerateOutline, colour.AlertModerateTint:
		return SeverityModerate
	case colour.AlertMinorOutline, colour.AlertMinorTint:
		return SeverityMinor
	case colour.AlertUnknownOutline, colour.AlertUnknownTint:
		return SeverityUnknown
	}
	return 0
}

// Overlay is what a host hands in: an id, when its data was valid and how
// long it stays current (FR-32), a credit, and its shapes.
type Overlay struct {
	ID       string
	Valid    time.Time
	Keeps    time.Duration
	Credit   string
	Features []Feature
	Grid     *Grid  // a scalar grid; an overlay is features, or a grid, or an image, and one only
	Image    *Image // a georeferenced image with its colour table
}

// Caps are a store's limits, fixed when it is made.
type Caps struct {
	OverlayVertices int // zero means 2,000,000
	StoreVertices   int // zero means 4,000,000
	ShapeBytes      int // the shape cache's cap; zero means 250,000
	ImageBytes      int // the image cap, at one byte a pixel; zero means 250,000
	// Classified is the pictures the maps of a shared set have already
	// read. Nil means this store reads its own and shares them with nobody
	// (D-116).
	Classified *Classified
}

// SetResult is what Set returns at once (D-74, D-86).
type SetResult struct {
	Created  bool // false means an overlay of that id was replaced
	Released bool // the old geometry is no longer read by anything
}

// RemoveResult is what Remove returns at once.
type RemoveResult struct {
	Found    bool
	Released bool
}

// held is one version of an overlay's geometry, and who is reading it.
type held struct {
	overlay  Overlay
	vertices int
	readers  int
	retired  bool        // replaced or removed: released when its last reader leaves
	kind     Kind        // a grid's type, resolved at hand-in
	box      project.Box // where it is, recorded at hand-in so fit-to needs no work (D-76)
	index    []Box       // built inside Set for a shape that may be drawn from memory (D-92)
	charge   int64       // what it costs the image budget: an image's files and pixels, a grid's field
}

// Store holds a map's overlays. Set and Remove never wait: a version that is
// still being read is retired, and the reader's own return reports its
// release. It has a lock of its own, never held across slow work.
type Store struct {
	mu       sync.Mutex
	caps     Caps
	current  map[string]*held
	order    []string
	retiring map[string]int // ids with a retired version still being read, and how many
	vertices int
	warnings []fault.Warning

	prepared   map[string]map[int]*prepared    // by overlay id, then bucket
	fields     map[string]scene.Field          // prepared grids, by overlay id
	pictures   map[string][]picture            // decoded pictures, by overlay id, one a frame
	spare      map[string]map[[32]byte]picture // a replaced loop's decoded frames, by key, for its refresh to keep (L-1.7)
	decodes    int                             // pictures decoded, counted so a test can see a refresh reuse them
	budget     int64                           // what the images may hold in all (L-12.1)
	landed     uint64                          // prepared forms kept, for Work to see (D-66)
	fromMemory map[string]bool                 // overlays whose prepared form is larger than the whole cap
	views      map[*ShapeView]int              // live views, and the bucket each draws at
	released   []string
	shapeHeld  int64
	shapeNeed  int64
	overNeed   bool
	clock      uint64
}

// NewStore makes a store. Caps may be lowered, never raised.
func NewStore(c Caps) (*Store, error) {
	if c.OverlayVertices < 0 || c.StoreVertices < 0 || c.OverlayVertices > defaultOverlayVertices || c.StoreVertices > defaultStoreVertices || c.ShapeBytes < 0 || c.ShapeBytes > defaultShapeBytes || c.ImageBytes < 0 || c.ImageBytes > defaultImageBytes {
		return nil, fault.Make(fault.OverVertexCap, textsafe.Const("the overlay limits were refused"),
			textsafe.Const("a cap may be lowered, never raised: the memory the library promises is stated against it"),
			textsafe.Const("give a cap between 1 and the default, or none"))
	}
	if c.OverlayVertices == 0 {
		c.OverlayVertices = defaultOverlayVertices
	}
	if c.StoreVertices == 0 {
		c.StoreVertices = defaultStoreVertices
	}
	if c.ShapeBytes == 0 {
		c.ShapeBytes = defaultShapeBytes
	}
	if c.ImageBytes == 0 {
		c.ImageBytes = defaultImageBytes
	}
	return &Store{budget: defaultImageBudget, caps: c, current: map[string]*held{}, retiring: map[string]int{}, prepared: map[string]map[int]*prepared{}, fields: map[string]scene.Field{}, pictures: map[string][]picture{}, spare: map[string]map[[32]byte]picture{}, fromMemory: map[string]bool{}, views: map[*ShapeView]int{}}, nil
}

func refused(kind fault.Kind, why, todo textsafe.Text) error {
	return fault.Make(kind, textsafe.Const("the overlay was refused"), why, todo)
}

func grouped(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// checkRing holds one ring's positions to the globe. Positions with latitude
// out of range and longitude in it are very likely the wrong way round, and
// the message says so.
func checkRing(ring []project.LonLat) error {
	for _, p := range ring {
		if !(p.Lon >= -180 && p.Lon <= 180) || !(p.Lat >= -90 && p.Lat <= 90) {
			return refused(fault.InvalidCoordinates,
				textsafe.Const("one of its coordinates is not a number, or is off the globe; longitude and latitude the wrong way round is the usual cause"),
				textsafe.Const("give longitude from -180 to 180 first, then latitude from -90 to 90"))
		}
	}
	return nil
}

// checkFeature validates one feature and counts its vertices.
func checkFeature(f Feature) (int, error) {
	if f.Kind < Point || f.Kind > Circle {
		return 0, refused(fault.InvalidCoordinates, textsafe.Const("one of its features has no kind"),
			textsafe.Const("say whether each feature is a point, a line, a polygon or a circle"))
	}
	if f.Role.Name() == "" {
		return 0, refused(fault.UnknownPreset, textsafe.Const("one of its features names a role that is no token"),
			textsafe.Const("give each feature one of the library's tokens, such as an alert's outline"))
	}
	if f.Severity > SeverityExtreme {
		return 0, refused(fault.UnknownPreset, textsafe.Const("one of its features gives a severity that is none of the library's"),
			textsafe.Const("give unknown, minor, moderate, severe or extreme, or none for the one its role implies"))
	}
	if f.Kind == Circle && len(f.Rings) != 0 {
		return 0, refused(fault.InvalidCoordinates, textsafe.Const("a circle has a centre and a radius and no rings, and this one has rings as well"),
			textsafe.Const("hand a circle in with its centre and radius alone, or hand the rings in as a polygon"))
	}
	if f.Kind == Circle {
		if !(f.RadiusKm > 0 && f.RadiusKm <= 20040) {
			return 0, refused(fault.InvalidCoordinates, textsafe.Const("a circle's radius is not a distance on the globe"),
				textsafe.Const("give the radius in kilometres, above zero"))
		}
		return 1, checkRing([]project.LonLat{f.Centre})
	}
	least := map[FeatureKind]int{Point: 1, Line: 2, Polygon: 3}[f.Kind]
	if len(f.Rings) == 0 {
		return 0, refused(fault.RingTooShort, textsafe.Const("one of its features has no ring of positions at all"),
			textsafe.Const("give a polygon at least three positions, a line two, a point one"))
	}
	total := 0
	for _, ring := range f.Rings {
		if len(ring) < least {
			return 0, refused(fault.RingTooShort, textsafe.Const("one of its rings is too short: a polygon needs three positions, a line two, a point one"),
				textsafe.Const("check the geometry; a ring may repeat its first position to close, and need not"))
		}
		err := checkRing(ring)
		if err != nil {
			return 0, err
		}
		total += len(ring)
	}
	return total, nil
}

// check validates an overlay on hand-in (NFR-20) and counts its vertices.
func (s *Store) check(o Overlay) (int, error) {
	if _, err := textsafe.ID(o.ID); err != nil {
		return 0, refused(fault.InvalidID, textsafe.Const("its id is empty, longer than 256 bytes, or holds characters that cannot be shown safely"),
			textsafe.Const("give an id of plain text; ids are never altered, so one that would need it is refused"))
	}
	err := CheckCurrency(o.Valid, o.Keeps)
	if err != nil {
		return 0, err
	}
	kinds := 0
	for _, has := range []bool{len(o.Features) != 0, o.Grid != nil, o.Image != nil} {
		if has {
			kinds++
		}
	}
	if kinds > 1 {
		return 0, refused(fault.SizeMismatch, textsafe.Const("it has more than one of features, a grid and an image: an overlay is one kind of thing"),
			textsafe.Const("hand them in as separate overlays, each with an id of its own"))
	}
	if o.Image != nil {
		_, err := checkImage(o.Image, s.caps.ImageBytes)
		return 0, err
	}
	if o.Grid != nil {
		_, err := checkGrid(o.Grid)
		return 0, err
	}
	if len(o.Features) == 0 {
		return 0, refused(fault.SizeMismatch, textsafe.Const("there is nothing in it: no feature at all"),
			textsafe.Const("to take an overlay away, remove it by its id"))
	}
	vertices := 0
	for _, f := range o.Features {
		n, err := checkFeature(f)
		if err != nil {
			return 0, err
		}
		vertices += n
	}
	if vertices > s.caps.OverlayVertices {
		return 0, refused(fault.OverVertexCap,
			textsafe.Join(textsafe.Const("it has "), textsafe.Clean(grouped(vertices)), textsafe.Const(" vertices, over the cap of "), textsafe.Clean(grouped(s.caps.OverlayVertices)), textsafe.Const(" for one overlay")),
			textsafe.Const("simplify the geometry before handing it in, or split it between overlays"))
	}
	return vertices, nil
}

// boxOfOverlay is where an overlay is, recorded at hand-in so that fitting
// the view to it needs no work to have run (D-76). A grid's and an image's
// box is the one they state; a feature overlay's is round its geometry,
// with a circle grown by its radius in degrees of latitude.
func boxOfOverlay(o Overlay) project.Box {
	if o.Grid != nil {
		return project.Box{West: o.Grid.West, South: o.Grid.South, East: o.Grid.East, North: o.Grid.North}
	}
	if o.Image != nil {
		return project.Box{West: o.Image.West, South: o.Image.South, East: o.Image.East, North: o.Image.North}
	}
	box := project.Box{West: 180, South: 90, East: -180, North: -90}
	for _, f := range o.Features {
		for _, ring := range f.Rings {
			for _, at := range ring {
				box = grown(box, at, 0)
			}
		}
		if f.Kind == Circle {
			box = grown(box, f.Centre, f.RadiusKm/111.0)
		}
	}
	if box.West > box.East || box.South > box.North {
		return project.Box{} // nothing to fit to
	}
	return box
}

// grown is a box grown to hold a place, and a margin in degrees around it.
func grown(box project.Box, at project.LonLat, pad float64) project.Box {
	return project.Box{
		West:  min(box.West, at.Lon-pad),
		South: min(box.South, at.Lat-pad),
		East:  max(box.East, at.Lon+pad),
		North: max(box.North, at.Lat+pad),
	}
}

// Box is where an overlay is, as recorded when it was handed in (D-76).
func (s *Store) Box(id string) (project.Box, bool) {
	if s == nil || id == "" {
		return project.Box{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.current[id]
	if !ok || h.box == (project.Box{}) {
		return project.Box{}, false
	}
	return h.box, true
}

// warnLocked records a warning, counting a repeat and keeping at most 64.
func (s *Store) warnLocked(kind fault.WarningKind, subject textsafe.Text) {
	for i := range s.warnings {
		if s.warnings[i].Kind == kind && s.warnings[i].Subject == subject {
			s.warnings[i].Count++
			return
		}
	}
	if len(s.warnings) < maxWarnings {
		s.warnings = append(s.warnings, fault.Warning{Kind: kind, Subject: subject, Count: 1})
	}
}

// warnCountLocked records a warning that counts something other than how
// often it was raised: the pixels of an image that matched nothing.
func (s *Store) warnCountLocked(kind fault.WarningKind, subject textsafe.Text, count int) {
	if count <= 0 {
		return
	}
	for i := range s.warnings {
		if s.warnings[i].Kind == kind && s.warnings[i].Subject == subject {
			s.warnings[i].Count = count
			return
		}
	}
	if len(s.warnings) < maxWarnings {
		s.warnings = append(s.warnings, fault.Warning{Kind: kind, Subject: subject, Count: count})
	}
}

// TakeWarnings returns what has gone wrong since it was last called.
func (s *Store) TakeWarnings() []fault.Warning {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.warnings
	s.warnings = nil
	return out
}

// folded is an id with what a slip of the hand changes taken out of it.
func folded(id string) string {
	return strings.Map(func(r rune) rune {
		if r == '-' || r == '_' || r == ' ' || r == '.' {
			return -1
		}
		return r
	}, strings.ToLower(id))
}

// near reports whether two different ids differ only slightly: by case or
// punctuation, or by one character more or less at the end.
func near(a, b string) bool {
	if a == b {
		return false
	}
	fa, fb := folded(a), folded(b)
	if fa == fb {
		return true
	}
	if len(fa) > len(fb) {
		fa, fb = fb, fa
	}
	return len(fb)-len(fa) == 1 && strings.HasPrefix(fb, fa)
}

// retireLocked takes a version out of use. It is released at once unless
// something is still reading it.
func (s *Store) retireLocked(id string, h *held) bool {
	s.vertices -= h.vertices
	h.retired = true
	if h.readers == 0 {
		return true
	}
	s.retiring[id] += h.readers
	return false
}

// Set adds an overlay, or replaces the one of the same id. It never waits.
func (s *Store) HandIn(o Overlay) (SetResult, error) {
	if s == nil {
		return SetResult{}, refused(fault.Internal, textsafe.Const("there is no store to set it in"), textsafe.Const("this is a defect in the library; report it"))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	if o.Image != nil {
		o.Image, err = copied(o.Image) // the copy is what is checked and kept (L-1.14)
	}
	vertices := 0
	if err == nil {
		vertices, err = s.check(o)
	}
	if err == nil {
		replaced := 0
		if old, ok := s.current[o.ID]; ok {
			replaced = old.vertices
		}
		if s.vertices-replaced+vertices > s.caps.StoreVertices {
			err = refused(fault.OverVertexCap,
				textsafe.Join(textsafe.Const("with it the map's overlays would have more than "), textsafe.Clean(grouped(s.caps.StoreVertices)), textsafe.Const(" vertices in all")),
				textsafe.Const("remove an overlay first, or simplify the geometry before handing it in"))
		}
	}
	charge := chargeOf(o)
	if err == nil {
		err = s.fitsBudgetLocked(o.ID, charge)
	}
	if err != nil {
		s.warnLocked(fault.SetRefused, textsafe.Quote(o.ID)) // a discarded error still shows
		return SetResult{}, err
	}
	res := SetResult{Created: true, Released: true}
	if old, ok := s.current[o.ID]; ok {
		res.Created, res.Released = false, s.retireLocked(o.ID, old)
	} else {
		s.order = append(s.order, o.ID)
		for _, other := range s.order {
			if near(other, o.ID) {
				s.warnLocked(fault.NearDuplicateID, textsafe.Quote(o.ID))
				break
			}
		}
	}
	next := &held{overlay: o, vertices: vertices, box: boxOfOverlay(o), charge: charge}
	if o.Image != nil {
		next.kind, _ = ResolveType(o.Image.Type) // it resolved a moment ago, in check
	}
	if o.Grid != nil {
		next.kind, _ = ResolveType(o.Grid.Type) // it resolved a moment ago, in check
		if implausible(o.Grid) {
			s.warnLocked(fault.ImplausibleUnit, textsafe.Quote(o.ID))
		}
	}
	if vertices > s.IndexFrom() {
		next.index = buildIndex(o) // one linear pass, inside Set, so the shape draws next frame (D-92)
	}
	s.spareFramesLocked(o)     // a refresh keeps the frames it shares with the loop it replaces (L-1.7)
	s.dropPreparedLocked(o.ID) // what was prepared from the old geometry is not this overlay's
	s.current[o.ID] = next
	s.vertices += vertices
	return res, nil
}

// Landed counts the prepared shapes, fields and pictures kept: what a Work
// call compares before and after, to know it landed something (D-66).
func (s *Store) Landed() uint64 {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.landed
}

// chargeOf is what an overlay costs the image budget: an image's files and
// pixels, or a grid's field at one byte a cell. Features cost it nothing:
// the shape cache has its own cap.
func chargeOf(o Overlay) int64 {
	if o.Grid != nil {
		return int64(o.Grid.Cols) * int64(o.Grid.Rows)
	}
	return imageCharge(o.Image)
}

// SetBudget sets what the map's images may hold in all. Zero or less is the
// library's own, 6 MiB. Lowering it below what is held drops nothing; the
// next hand-in that does not fit is refused (D-72). A host that raises it
// owns the memory it asks for.
func (s *Store) SetBudget(bytes int64) {
	if s == nil {
		return
	}
	if bytes <= 0 {
		bytes = defaultImageBudget
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.budget = bytes
}

// ImageUse is what the budget counts now: every image's and grid's charge,
// and the shared classified set. The shared set counts in full, even where it
// holds this map's own pictures, so the count errs high, never low (L-12.6).
func (s *Store) ImageUse() int64 {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.imageUseLocked()
}

func (s *Store) imageUseLocked() int64 {
	total := int64(0)
	for _, h := range s.current {
		total += h.charge
	}
	shared, _ := s.caps.Classified.Bytes()
	return total + shared
}

// fitsBudgetLocked refuses a hand-in that would take the images over the
// budget, saying by how much. The version it replaces is not counted.
func (s *Store) fitsBudgetLocked(id string, charge int64) error {
	if charge == 0 {
		return nil
	}
	use := s.imageUseLocked() + charge
	if old, ok := s.current[id]; ok {
		use -= old.charge
	}
	if use <= s.budget {
		return nil
	}
	return refused(fault.OverImageCap,
		textsafe.Join(textsafe.Const("with it the map's images would come to "), textsafe.Clean(grouped(int(use))), textsafe.Const(" bytes, "),
			textsafe.Clean(grouped(int(use-s.budget))), textsafe.Const(" over its image budget of "), textsafe.Clean(grouped(int(s.budget)))),
		textsafe.Const("hand in fewer or smaller frames, remove an image, or raise the budget with SetImageBudget"))
}

// spareFramesLocked keeps, by key, the decoded frames of the loop an overlay
// replaces, for the new version's job to take instead of decoding them again.
// Only a loop replacing a loop keeps any; the new version's own keys decide
// which it takes, and what it does not take is dropped when it is kept.
//
// Two refreshes before any work keep what the first kept: the spare set grows
// only by frames a job decoded, and the job that keeps a version empties it,
// so it never holds more than one version's frames.
func (s *Store) spareFramesLocked(o Overlay) {
	spare := s.spare[o.ID]
	delete(s.spare, o.ID)
	if o.Image == nil || len(o.Image.Frames) == 0 {
		return
	}
	if spare == nil {
		spare = map[[32]byte]picture{}
	}
	for _, p := range s.pictures[o.ID] {
		if p.ready && p.key != ([32]byte{}) {
			spare[p.key] = p
		}
	}
	if len(spare) > 0 {
		s.spare[o.ID] = spare
	}
}

// Remove takes an overlay away. It never waits, and an id that is not set is
// not an error: the result says it was not found.
func (s *Store) Drop(id string) (RemoveResult, error) {
	if s == nil {
		return RemoveResult{}, refused(fault.Internal, textsafe.Const("there is no store to remove it from"), textsafe.Const("this is a defect in the library; report it"))
	}
	if _, err := textsafe.ID(id); err != nil {
		return RemoveResult{}, refused(fault.InvalidID, textsafe.Const("the id is empty, longer than 256 bytes, or holds characters that cannot be shown safely"),
			textsafe.Const("give the id the overlay was set with"))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.current[id]
	if !ok {
		return RemoveResult{Released: true}, nil
	}
	delete(s.current, id)
	delete(s.spare, id)
	s.dropPreparedLocked(id)
	for i, other := range s.order {
		if other == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return RemoveResult{Found: true, Released: s.retireLocked(id, old)}, nil
}

// InUse reports whether any call is still reading geometry of that id that
// has been replaced or removed.
func (s *Store) Reading(id string) bool {
	if s == nil || id == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.retiring[id] > 0
}

// IDs lists the overlays set, in the order they were first set.
func (s *Store) IDs() []string {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.order...)
}

// OwnedBytes is what the store itself holds for its overlays. Borrowed
// geometry is the host's memory and is not counted, because it is not held.
func (s *Store) OwnedBytes() int {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	total := 0
	total += int(s.shapeHeld)
	for _, f := range s.fields {
		total += len(f.Classes)
	}
	for _, pictures := range s.pictures {
		for _, p := range pictures {
			total += len(p.raster.Classes) // one byte a pixel (D-36)
		}
	}
	for _, spare := range s.spare {
		for _, p := range spare {
			total += len(p.raster.Classes) // held until the refresh's job takes or drops it
		}
	}
	for id, h := range s.current {
		total += 160 + len(id) + len(h.overlay.Credit) + 96*len(h.overlay.Features) + 16*len(h.index)
		for _, f := range h.overlay.Features {
			total += len(f.Label) + 24*len(f.Rings)
		}
	}
	return total
}

// Reader is one call's hold on a version of an overlay's geometry.
type Reader struct {
	store *Store
	id    string
	h     *held
}

// Read takes a hold on an overlay's current geometry, for a job that reads
// it off the drawing path. Done must be called when the read is over.
func (s *Store) Read(id string) (*Reader, bool) {
	if s == nil || id == "" {
		return nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.current[id]
	if !ok {
		return nil, false
	}
	h.readers++
	return &Reader{store: s, id: id, h: h}, true
}

// Overlay is the version this reader holds.
func (r *Reader) Overlay() Overlay {
	if r == nil || r.h == nil {
		return Overlay{}
	}
	return r.h.overlay
}

// Done ends the read. It returns the id if this was the last read of geometry
// that has been replaced or removed: that is how a release that Set or Remove
// could not report is reported, by the return of the call that ended it.
func (r *Reader) Done() []string {
	if r == nil || r.h == nil {
		return nil
	}
	s := r.store
	s.mu.Lock()
	defer s.mu.Unlock()
	h := r.h
	r.h = nil
	h.readers--
	if !h.retired {
		return nil
	}
	s.retiring[r.id]--
	if s.retiring[r.id] > 0 {
		return nil
	}
	delete(s.retiring, r.id)
	s.released = append(s.released, r.id)
	return []string{r.id}
}
