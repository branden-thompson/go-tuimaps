package testkit

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestShapedServer is plan task 05.12.
func TestShapedServer(t *testing.T) {
	// NFR-5's link: 8 Mbit/s, 150 ms a round trip, a pump two wide.
	link := Link{Setup: 300 * time.Millisecond, RoundTrip: 150 * time.Millisecond, BitsPerSecond: 8_000_000, Width: 2}
	sizes := map[string]int{"/5/7/11": 291348, "/5/7/12": 322978, "/5/8/11": 283958, "/5/8/12": 346879}
	srv, err := NewShapedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, strings.Repeat("x", sizes[r.URL.Path]))
	}), link)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Shutdown()

	// Its trust anchor is its own: a client that does not have it is refused.
	stranger := &http.Client{Transport: &http.Transport{DialContext: Dialer().DialContext, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}}}
	if resp, err := stranger.Get(srv.URL + "/5/7/11"); err == nil {
		resp.Body.Close()
		t.Fatal("a client without the server's trust anchor was served")
	}
	if st := srv.Stats(); st.Requests != 0 || st.VirtualElapsed != 0 {
		t.Errorf("a refused handshake cost %d requests and %v", st.Requests, st.VirtualElapsed)
	}

	client := &http.Client{Transport: &http.Transport{DialContext: Dialer().DialContext, TLSClientConfig: &tls.Config{RootCAs: srv.RootCAs, MinVersion: tls.VersionTLS12}}}
	start := time.Now()
	for path, size := range sizes {
		resp, err := client.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		n, _ := io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if int(n) != size {
			t.Errorf("%s: %d bytes, want %d", path, n, size)
		}
	}
	if real := time.Since(start); real > 3*time.Second {
		t.Errorf("the server really took %v; its clock is virtual, and nothing should wait", real)
	}
	// Setup once, two rounds of two requests, 1,245,163 bytes at 8 Mbit/s:
	// 0.300 + 0.300 + 1.245163 s - the constants file's own arithmetic.
	want := 300*time.Millisecond + 2*150*time.Millisecond + time.Duration(1245163*8)*time.Second/8_000_000
	st := srv.Stats()
	if st.VirtualElapsed != want {
		t.Errorf("virtual time %v, want %v", st.VirtualElapsed, want)
	}
	if st.Requests != 4 || st.Bytes != 1245163 {
		t.Errorf("%d requests, %d bytes", st.Requests, st.Bytes)
	}
}

func TestShapedServerRefusesABadLink(t *testing.T) {
	ok := http.NotFoundHandler()
	for name, link := range map[string]Link{
		"no bandwidth":   {RoundTrip: time.Millisecond, Width: 1},
		"no width":       {RoundTrip: time.Millisecond, BitsPerSecond: 1},
		"negative delay": {RoundTrip: -1, BitsPerSecond: 1, Width: 1},
	} {
		if srv, err := NewShapedServer(ok, link); err == nil {
			srv.Shutdown()
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := NewShapedServer(nil, Link{BitsPerSecond: 1, Width: 1}); err == nil {
		t.Error("no handler: accepted")
	}
}
