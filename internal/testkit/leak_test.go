package testkit

import (
	"strings"
	"testing"
)

// leaky parks until released. It stands in for a goroutine the library
// might wrongly leave behind: its frames are inside this module.
func leaky(started, release chan struct{}) {
	close(started)
	<-release
}

// TestLeakCheckFindsAGoroutineInsideTheLibrary is plan task 00.4.
func TestLeakCheckFindsAGoroutineInsideTheLibrary(t *testing.T) {
	if err := LeakCheck(ModulePath); err != nil {
		t.Fatalf("a leak was reported before any goroutine was started: %v", err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	go leaky(started, release)
	<-started
	err := LeakCheck(ModulePath)
	if err == nil {
		t.Fatal("a parked goroutine inside the module was not reported")
	}
	if !strings.Contains(err.Error(), "leaky") {
		t.Errorf("the report %q does not name the function that is parked", err)
	}
	close(release)
	if err := LeakCheck(ModulePath); err != nil {
		t.Errorf("after release the goroutine was still reported: %v", err)
	}
}

func TestLeakCheckIgnoresGoroutinesOutsideTheLibrary(t *testing.T) {
	const dump = "goroutine 1 [running]:\n" +
		"testing.tRunner(0x1, 0x2)\n\t/go/src/testing/testing.go:1\n\n" +
		"goroutine 7 [IO wait]:\n" +
		"internal/poll.runtime_pollWait(0x1, 0x72)\n\t/go/src/runtime/netpoll.go:1\n" +
		"net/http.(*persistConn).readLoop(0x1)\n\t/go/src/net/http/transport.go:1\n" +
		"created by net/http.(*Transport).dialConn in goroutine 5\n\t/go/src/net/http/transport.go:2\n\n" +
		"goroutine 9 [chan receive]:\n" +
		"example.com/lib/internal/tiles.(*cache).loop(0x1)\n\t/src/cache.go:1\n" +
		"created by example.com/lib.New in goroutine 1\n\t/src/map.go:2\n\n" +
		"goroutine 11 [chan receive]:\n" +
		"net/http.(*persistConn).writeLoop(0x1)\n\t/go/src/net/http/transport.go:3\n" +
		"created by example.com/lib/internal/fetch.dial in goroutine 1\n\t/src/fetch.go:2\n"
	got := leakedIn(dump, "example.com/lib")
	if len(got) != 1 || !strings.Contains(got[0], "tiles.(*cache).loop") {
		t.Errorf("got %q; want only goroutine 9 — the network stack's own goroutines are not the library's, even when the library's call started them", got)
	}
}
