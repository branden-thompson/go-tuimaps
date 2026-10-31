package tuimaps_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// TestTheCheckedDialerExemptsOnlyTheNamedProxy is L-7.4 (OW-19, D-117): a
// host behind a proxy at a private address keeps the check. The dialer
// reaches the proxy once the proxy function names it, and refuses it before;
// any other private address stays refused; a request through a transport
// set with both goes through the proxy.
func TestTheCheckedDialerExemptsOnlyTheNamedProxy(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Host != "tiles.example.org" {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = io.WriteString(w, "through the proxy")
	}))
	defer proxy.Close()
	other := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer other.Close()
	proxyURL, _ := url.Parse(proxy.URL)
	dial, choose := tuimaps.CheckedDialerFor(func(*http.Request) (*url.URL, error) { return proxyURL, nil })
	ctx := context.Background()
	if c, err := dial(ctx, "tcp", proxyURL.Host); err == nil {
		_ = c.Close()
		t.Fatal("the proxy, at a loopback address, was dialled before the proxy function named it")
	}
	req, _ := http.NewRequest(http.MethodGet, "http://tiles.example.org/0/0/0.pbf", nil)
	if u, err := choose(req); err != nil || u.Host != proxyURL.Host {
		t.Fatalf("the proxy function answered %v, %v; want the host's proxy", u, err)
	}
	c, err := dial(ctx, "tcp", proxyURL.Host)
	if err != nil {
		t.Fatalf("the named proxy was refused: %v", err)
	}
	_ = c.Close()
	otherHost := other.Listener.Addr().(*net.TCPAddr).String()
	if c, err := dial(ctx, "tcp", otherHost); err == nil {
		_ = c.Close()
		t.Error("a loopback address that is not the proxy was dialled")
	}
	if c, err := tuimaps.CheckedDialer()(ctx, "tcp", proxyURL.Host); err == nil {
		_ = c.Close()
		t.Error("the plain checked dialer reached the private proxy")
	}
	if d, p := tuimaps.CheckedDialerFor(nil); d == nil || p != nil {
		t.Error("with no proxy function: want the plain checked dialer and no proxy function")
	}
	_, blank := tuimaps.CheckedDialerFor(func(*http.Request) (*url.URL, error) { return &url.URL{Scheme: "http"}, nil })
	if u, err := blank(req); err == nil {
		t.Errorf("a proxy named with no address was taken: %v", u)
	}
	client := &http.Client{Transport: &http.Transport{Proxy: choose, DialContext: dial}}
	resp, err := client.Get("http://tiles.example.org/0/0/0.pbf")
	if err != nil {
		t.Fatalf("a request through the proxy failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != "through the proxy" {
		t.Errorf("the request did not go through the proxy: %d %q", resp.StatusCode, body)
	}
}
