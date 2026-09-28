package work

// basemap_test.go — v0.2.0 L-22 (watchpost UAT-2 U2-34): the basemap is
// never starved by the overlays. With hundreds of overlays handed in, the
// queue - full at its cap - dropped its oldest job, the view's tiles, and ran
// the overlays first; the map drew markers over no basemap at all.

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// TestTheBasemapIsNeverStarved is L-22.1: a tile asked for before a flood of
// overlays past the cap stays queued - an overlay is dropped in its place -
// and is run before any overlay.
func TestTheBasemapIsNeverStarved(t *testing.T) {
	q := NewQueue(8)
	m, _ := q.Join()
	var ran []string
	var mu sync.Mutex
	record := func(kind scene.JobKind, key string) scene.Job {
		return job{kind: kind, key: key, run: func(context.Context) error { mu.Lock(); ran = append(ran, key); mu.Unlock(); return nil }}
	}
	add(t, m, record(scene.KindTile, "tile/1"))
	for i := range 20 {
		add(t, m, record(scene.KindOverlayPrepare, "overlay/"+strconv.Itoa(i)))
	}
	add(t, m, record(scene.KindTile, "tile/2"))
	if got := m.Backlog(); got != 8 {
		t.Fatalf("Pending() = %d with a cap of 8", got)
	}
	drain(t, m)
	if len(ran) < 2 || ran[0] != "tile/1" || ran[1] != "tile/2" {
		t.Errorf("the queue ran %v; want both tiles first, and neither dropped", ran)
	}
}

// TestAViewNoLongerShownGoesFirst is L-22.1's other half: past the cap, what
// goes first is a job of a view no map shows any more - even its tile -
// before an overlay of the view on screen.
func TestAViewNoLongerShownGoesFirst(t *testing.T) {
	q := NewQueue(3)
	m, _ := q.Join()
	var ran []string
	record := func(kind scene.JobKind, key string) scene.Job {
		return job{kind: kind, key: key, run: func(context.Context) error { ran = append(ran, key); return nil }}
	}
	add(t, m, record(scene.KindTile, "old/tile"))
	m.NewView()
	add(t, m, record(scene.KindOverlayPrepare, "overlay/1"), record(scene.KindOverlayPrepare, "overlay/2"), record(scene.KindTile, "tile"))
	drain(t, m)
	if strings.Join(ran, " ") != "tile overlay/1 overlay/2" {
		t.Errorf("the queue ran %v; want the old view's tile dropped, and the view shown kept whole", ran)
	}
}
