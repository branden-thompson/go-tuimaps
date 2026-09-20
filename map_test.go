package tuimaps_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/assets"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// Example_threeCalls is plan task 12.1: the whole of a world map in a rectangle: create, settle,
// render (NFR-19). It reaches no network: the tiles are the embedded ones,
// passed as an option.
func Example_threeCalls() {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err := m.Settle(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
	frame, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, time.Time{})
	fmt.Println(len(frame.Lines), "rows,", frame.Status, err)
	// Output: 24 rows, complete <nil>
}

func isKind(err error, k fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == k
}

// TestThreeCalls is plan task 12.1 in full: the frame the three calls give is
// the world, with names, and nothing was reached to draw it.
func TestThreeCalls(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(149, 38), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	res, err := m.Settle(context.Background())
	if err != nil || res.Failed != 0 || res.Ran == 0 {
		t.Fatalf("settle: %+v, %v", res, err)
	}
	frame, err := m.Render(tuimaps.Size{Cols: 149, Rows: 38}, time.Time{})
	if err != nil || frame.Status != tuimaps.Complete || len(frame.Lines) != 38 {
		t.Fatalf("%v, status %v, %d rows", err, frame.Status, len(frame.Lines))
	}
	text := regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(strings.Join(frame.Lines, "\n"), "")
	if !strings.Contains(text, "Africa") || !strings.Contains(text, "OpenStreetMap") {
		t.Errorf("the world map lacks a continent's name or the credit line:\n%s", text)
	}
	// On a map wider than the world, names stay on the world.
	small, _ := tuimaps.New(tuimaps.WithSize(69, 12), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	defer small.Close()
	small.Settle(context.Background())
	narrow, _ := small.Render(tuimaps.Size{Cols: 69, Rows: 12}, time.Time{})
	for _, line := range narrow.Lines {
		plain := regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(line, "")
		if at := strings.Index(plain, "Asia"); at >= 0 && at < len(plain)/2 {
			t.Errorf("Asia is drawn in the western half of the map: %q", plain)
		}
	}
	if again, _ := m.Settle(context.Background()); again.Ran != 0 {
		t.Errorf("a second settle ran %d jobs; everything is on hand", again.Ran)
	}
}

// TestNewStartsNothing is plan task 12.2 (D-65, D-73): no goroutine, no
// connection, no file. The connection is the dial guard's to catch: every
// test in this package fails if anything dials.
func TestNewStartsNothing(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(dir, "cache"))
	before := runtime.NumGoroutine()
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	if after := runtime.NumGoroutine(); after != before {
		t.Errorf("%d goroutines before New, %d after", before, after)
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, time.Time{}); err != nil {
		t.Fatal(err)
	}
	if after := runtime.NumGoroutine(); after != before {
		t.Errorf("%d goroutines before, %d after a settle and a render", before, after)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Errorf("files appeared: %v, %v", entries, err)
	}
	if left := m.Close(); left != 0 {
		t.Errorf("%d calls still inside a map nobody is calling", left)
	}
}

// TestSizeIsStateBeforeSettle and TestSettleWithoutSizeRefused are plan task
// 12.4 (contract, section 3).
func TestSizeIsStateBeforeSettle(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24), tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if m.Pending() != 0 {
		t.Error("work is pending before anything asked for a view")
	}
	res, err := m.Settle(context.Background())
	if err != nil || res.Ran == 0 {
		t.Errorf("%+v, %v; with a size, Settle finds work to do before any Render", res, err)
	}
}

func TestSettleWithoutSizeRefused(t *testing.T) {
	m, err := tuimaps.New(tuimaps.Embed(assets.Tile, assets.MaxZoom))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.Settle(context.Background()); !isKind(err, fault.NoSize) {
		t.Errorf("%v; want the no-size kind", err)
	}
	for _, bad := range [][2]int{{0, 10}, {10, -1}, {100000, 100000}} {
		if _, err := tuimaps.New(tuimaps.WithSize(bad[0], bad[1])); !isKind(err, fault.NoSize) {
			t.Errorf("size %v: %v", bad, err)
		}
	}
}

// TestNoTilesSaysWhy: with no source and no embedded tiles, Settle says why
// nothing could be fetched, and the frame is the notice, not a blank (FR-23).
func TestNoTilesSaysWhy(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(80, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	res, err := m.Settle(context.Background())
	if err != nil || res.Why == nil || !strings.Contains(res.Why.Error(), "no source is named") {
		t.Errorf("%+v, %v", res, err)
	}
	frame, err := m.Render(tuimaps.Size{Cols: 80, Rows: 24}, time.Time{})
	if err != nil || frame.Status != tuimaps.NoTiles || !strings.Contains(strings.Join(frame.Lines, ""), "assets") {
		t.Errorf("%v, status %v", err, frame.Status)
	}
}

// TestRenderTakesTheSize: Render's size becomes the map's, and a closed map
// refuses every call.
func TestRenderTakesTheSize(t *testing.T) {
	m, _ := tuimaps.New(tuimaps.Embed(assets.Tile, assets.MaxZoom))
	frame, err := m.Render(tuimaps.Size{Cols: 60, Rows: 16}, time.Time{})
	if err != nil || len(frame.Lines) != 16 || frame.Status != tuimaps.NoTiles {
		t.Fatalf("%v, %d rows, status %v; nothing is on hand before any work", err, len(frame.Lines), frame.Status)
	}
	if m.Pending() == 0 {
		t.Error("the render noted nothing as wanted")
	}
	if _, err := m.Settle(context.Background()); err != nil {
		t.Fatal(err)
	}
	frame, _ = m.Render(tuimaps.Size{Cols: 60, Rows: 16}, time.Time{})
	if frame.Status != tuimaps.Complete {
		t.Errorf("after the work: %v", frame.Status)
	}
	m.Close()
	if _, err := m.Render(tuimaps.Size{Cols: 60, Rows: 16}, time.Time{}); !isKind(err, fault.Closed) {
		t.Errorf("a closed map rendered: %v", err)
	}
	if _, err := m.Settle(context.Background()); !isKind(err, fault.Closed) {
		t.Errorf("a closed map settled: %v", err)
	}
}
