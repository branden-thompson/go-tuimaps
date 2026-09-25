package tuimaps_test

// confine_test.go — v0.2.0 L8.4-L8.5 (L-10.1-L-10.3): one allow-list for
// fetching and for a TileJSON's tile addresses, through Map.Source, with the
// library's transport and with a host's; and the checked dialer a host
// transport can keep the private-address refusal with.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tuimaps "github.com/branden-thompson/go-tuimaps"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestOneAllowListThroughSource is L8.5 (L-10.1, L-10.2): a TileJSON source
// whose tiles are at another host, another port, or over another scheme is
// refused, and nothing is asked of the other place - with the library's own
// transport and with a host's.
func TestOneAllowListThroughSource(t *testing.T) {
	var elsewhere atomic.Int32
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { elsewhere.Add(1) }))
	defer other.Close()
	otherPort := other.URL[strings.LastIndex(other.URL, ":")+1:]
	var docs http.HandlerFunc
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { docs(w, r) }))
	defer source.Close()
	sourcePort := source.URL[strings.LastIndex(source.URL, ":")+1:]
	for name, template := range map[string]string{
		"another port":   "http://127.0.0.1:" + otherPort + "/{z}/{x}/{y}.pbf",
		"another host":   "http://localhost:" + sourcePort + "/{z}/{x}/{y}.pbf",
		"another scheme": "https://127.0.0.1:" + sourcePort + "/{z}/{x}/{y}.pbf",
	} {
		docs = func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"tilejson":"3.0.0","tiles":[%q],"minzoom":0,"maxzoom":6}`, template)
		}
		for _, transport := range []http.RoundTripper{nil, http.DefaultTransport} {
			m, err := tuimaps.New(tuimaps.WithSize(40, 12))
			if err != nil {
				t.Fatal(err)
			}
			must(t, m.SetFetchOptions(tuimaps.FetchOptions{Transport: transport}))
			must(t, m.Source(source.URL+"/tiles.json"))
			if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil {
				t.Fatal(err)
			}
			res, err := m.Settle(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if res.Failed == 0 || !isKind(res.Why, fault.FetchRefused) {
				t.Errorf("%s, host transport %v: %+v; want the source refused", name, transport != nil, res)
			}
			m.Close()
		}
	}
	if n := elsewhere.Load(); n != 0 {
		t.Errorf("the other server was asked %d times", n)
	}
}

// TestCheckedDialerRefusesAPrivateAddress is L8.4 (L-10.3): the dialer a host
// transport can use to keep the library's refusal of private addresses.
func TestCheckedDialerRefusesAPrivateAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	dial := tuimaps.CheckedDialer()
	if dial == nil {
		t.Fatal("no dialer")
	}
	address := strings.TrimPrefix(server.URL, "http://")
	if conn, err := dial(context.Background(), "tcp", address); err == nil {
		conn.Close()
		t.Errorf("the checked dialer connected to %s, a loopback address", address)
	}
	// A host transport built on it refuses the same through the library.
	transport := &http.Transport{DialContext: tuimaps.CheckedDialer()}
	m, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	must(t, m.SetFetchOptions(tuimaps.FetchOptions{Transport: transport, AllowHTTP: []string{"127.0.0.1"}}))
	must(t, m.Source(server.URL+"/"))
	if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil {
		t.Fatal(err)
	}
	if res, err := m.Settle(context.Background()); err != nil || res.Failed == 0 || !isKind(res.Why, fault.FetchRefused) {
		t.Errorf("a host transport on the checked dialer, to loopback: %+v, %v; want refused", res, err)
	}
}

// stubborn is a transport that ignores its request's context: it answers
// with the embedded tiles, but only after a pause the request has no say in.
type stubborn struct{ tiles http.RoundTripper }

func (s stubborn) RoundTrip(r *http.Request) (*http.Response, error) {
	time.Sleep(120 * time.Millisecond)
	return s.tiles.RoundTrip(r)
}

// TestALateAnswerIsNeverUsed is L8.3 through the public Map (L-7.3): a host
// transport that ignores its context holds only its own work, and what it
// brings back after its request's timeout never lands.
func TestALateAnswerIsNeverUsed(t *testing.T) {
	m, err := tuimaps.New(tuimaps.WithSize(40, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	asked := 0
	must(t, m.SetFetchOptions(tuimaps.FetchOptions{Transport: stubborn{tiles: served(t, &asked)}, Timeout: 30 * time.Millisecond}))
	must(t, m.Source("https://tiles.example.test/"))
	if _, err := m.Render(tuimaps.Size{Cols: 40, Rows: 12}, noon); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := m.Settle(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if asked == 0 {
		t.Fatal("the transport was never asked, so this proves nothing")
	}
	if res.Failed == 0 || m.CacheUse().Tiles.Held != 0 {
		t.Errorf("late answers: %+v, %d bytes of tiles held; want every one failed and none kept", res, m.CacheUse().Tiles.Held)
	}
}
