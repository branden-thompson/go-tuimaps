package tuimaps

// detail_test.go — v0.2.0 WP-L11, L11.1 (D-82): SetDetail on the map.

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/assets"
)

func settledLines(t *testing.T, set func(m *Map)) []string {
	t.Helper()
	m, err := New(WithSize(100, 30), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Recentre(LonLat{Lon: -98, Lat: 38}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(3); err != nil {
		t.Fatal(err)
	}
	set(m)
	at := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	if _, err := m.Render(Size{Cols: 100, Rows: 30}, at); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = m.Settle(ctx)
	f, err := m.Render(Size{Cols: 100, Rows: 30}, at)
	if err != nil {
		t.Fatal(err)
	}
	return slices.Clone(f.Lines) // a frame's lines are the renderer's until the next frame is composed
}

// TestSetDetailThinsThePictureAndFullIsTheDefault: a level below Full draws
// less; Full is the picture a host that never asks gets; a value outside the
// four is refused and changes nothing.
func TestSetDetailThinsThePictureAndFullIsTheDefault(t *testing.T) {
	unset := settledLines(t, func(*Map) {})
	full := settledLines(t, func(m *Map) {
		if err := m.SetDetail(DetailFull); err != nil {
			t.Fatal(err)
		}
	})
	essential := settledLines(t, func(m *Map) {
		if err := m.SetDetail(DetailEssential); err != nil {
			t.Fatal(err)
		}
	})
	if !slices.Equal(unset, full) {
		t.Error("Full is not the picture a host that never asks gets")
	}
	if slices.Equal(essential, full) {
		t.Error("Essential drew the same picture as Full")
	}
	refused := settledLines(t, func(m *Map) {
		if err := m.SetDetail(Detail(9)); err == nil {
			t.Error("a detail outside the four was accepted")
		}
	})
	if !slices.Equal(refused, full) {
		t.Error("a refused detail changed the picture")
	}
}

// TestMajorAndMinorRoadsSwitchApartOnTheMap is L-14.2 on the map: the minor
// roads switch without the major, and a switch is an input the frame reads.
func TestMajorAndMinorRoadsSwitchApartOnTheMap(t *testing.T) {
	m, err := New(WithSize(69, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	before := m.Changed()
	m.Layers(MinorRoadLayer, false)
	if m.Changed() == before {
		t.Fatal("switching the minor roads off changed nothing the frame reads")
	}
	if !m.look.off.Has(MinorRoadLayer) || m.look.off.Has(RoadLayer) {
		t.Errorf("the switches are %b: want the minor roads off and the major on", m.look.off)
	}
	m.Layers(RoadLayer, false)
	m.Layers(MinorRoadLayer, true)
	if m.look.off.Has(MinorRoadLayer) || !m.look.off.Has(RoadLayer) {
		t.Errorf("the switches are %b: want the major roads off and the minor on", m.look.off)
	}
}

// TestALevelChangedOnADrawnMapRedraws: the renderer's reuse of the last
// frame's work sees the level, so a map already drawn thins at once, and
// comes back whole.
func TestALevelChangedOnADrawnMapRedraws(t *testing.T) {
	m, err := New(WithSize(100, 30), Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.Recentre(LonLat{Lon: -98, Lat: 38}); err != nil {
		t.Fatal(err)
	}
	if err := m.Zoom(3); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	draw := func() []string {
		t.Helper()
		if _, err := m.Render(Size{Cols: 100, Rows: 30}, at); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = m.Settle(ctx)
		f, err := m.Render(Size{Cols: 100, Rows: 30}, at)
		if err != nil {
			t.Fatal(err)
		}
		return slices.Clone(f.Lines) // a frame's lines are the renderer's until the next frame is composed
	}
	full := draw()
	if err := m.SetDetail(DetailEssential); err != nil {
		t.Fatal(err)
	}
	if slices.Equal(draw(), full) {
		t.Error("a level changed on a drawn map replayed the old picture")
	}
	if err := m.SetDetail(DetailFull); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(draw(), full) {
		t.Error("back at Full the picture is not the one it was")
	}
}
