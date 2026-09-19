package testkit

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestLoopbackOnlyRefusesAPublicAddress is plan task 00.5. The address is
// from the range reserved for documentation, and the refusal comes before
// any packet is sent.
func TestLoopbackOnlyRefusesAPublicAddress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, addr := range []string{"192.0.2.1:80", "[2001:db8::1]:443", "10.0.0.1:80", "169.254.1.1:80"} {
		conn, err := Dialer().DialContext(ctx, "tcp", addr)
		if err == nil {
			conn.Close()
			t.Fatalf("dial %s succeeded; tests may reach loopback only", addr)
		}
		if !errors.Is(err, ErrNotLoopback) {
			t.Errorf("dial %s: %v; want the loopback-only refusal", addr, err)
		}
	}
}

func TestLoopbackOnlyAllowsLoopback(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, err := Dialer().DialContext(ctx, "tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("a loopback dial was refused: %v", err)
	}
	conn.Close()
}

func TestCheckLoopbackAddress(t *testing.T) {
	ok := []string{"127.0.0.1:80", "127.9.9.9:1", "[::1]:443"}
	bad := []string{"", "example.com:80", "localhost:80", "8.8.8.8:53", "[::ffff:8.8.8.8]:53", "0.0.0.0:80", "127.0.0.1"}
	for _, a := range ok {
		if err := CheckLoopback(a); err != nil {
			t.Errorf("%s refused: %v", a, err)
		}
	}
	for _, a := range bad {
		if err := CheckLoopback(a); err == nil {
			t.Errorf("%q allowed; only a literal loopback address with a port may pass", a)
		}
	}
}

func TestInstallGuardsTheDefaultTransport(t *testing.T) {
	restore, err := InstallLoopbackOnly()
	if err != nil {
		t.Fatal(err)
	}
	defer restore()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://192.0.2.1/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("the default client reached a public address")
	}
	if !errors.Is(err, ErrNotLoopback) {
		t.Errorf("got %v; want the loopback-only refusal", err)
	}
}

// TestBlockingTransportNeverReturnsUntilCancelled is plan task 00.6.
func TestBlockingTransportNeverReturnsUntilCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1/tile", nil)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		resp, err := BlockingTransport{}.RoundTrip(req)
		if resp != nil {
			resp.Body.Close()
		}
		done <- err
	}()
	select {
	case err := <-done:
		t.Fatalf("the transport returned before it was cancelled: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("got %v; want the context's own error", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the transport did not return after it was cancelled")
	}
}

func TestBlockingTransportRefusesAMalformedRequest(t *testing.T) {
	if _, err := (BlockingTransport{}).RoundTrip(nil); err == nil {
		t.Error("a nil request must be an error, not a block for ever")
	}
	if _, err := (BlockingTransport{}).RoundTrip(&http.Request{}); err == nil {
		t.Error("a request with no address must be an error")
	}
}
