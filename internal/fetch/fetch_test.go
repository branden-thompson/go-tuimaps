package fetch

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

func isKind(err error, k fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == k
}

// secure starts a loopback server with secure transport and returns it with
// options that trust its certificate.
func secure(t *testing.T, h http.Handler) (*httptest.Server, Options) {
	t.Helper()
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	return srv, Options{RootCAs: pool, Token: "watchpost-test"}
}

func body(s string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, s) })
}

func get(t *testing.T, f *Fetcher, rawURL string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return f.Fetch(ctx, Request{URL: rawURL, MaxBytes: 1 << 20})
}

func TestGetOverSecureTransport(t *testing.T) {
	srv, opts := secure(t, body("tile bytes"))
	f, err := ForSource(srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	got, err := get(t, f, srv.URL+"/6/16/26.pbf")
	if err != nil || string(got) != "tile bytes" {
		t.Errorf("%q, %v", got, err)
	}
}

// TestPlainHTTPRefused is plan task 05.1.
func TestPlainHTTPRefused(t *testing.T) {
	if _, err := ForSource("http://tiles.example/planet", Options{}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a plain source on a public name: %v", err)
	}
	if _, err := ForSource("http://localhost:8080/planet", Options{}); !isKind(err, fault.FetchRefused) {
		t.Errorf("localhost is a name, and a name can resolve anywhere: %v", err)
	}
	if _, err := ForSource("ftp://tiles.example/planet", Options{AllowHTTP: []string{"tiles.example"}}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a scheme that is neither: %v", err)
	}
	if _, err := ForSource("https://user:secret@tiles.example/planet", Options{}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a source with a user-info part: %v", err)
	}
	if _, err := ForSource("http://tiles.example/planet", Options{AllowHTTP: []string{"tiles.example"}}); err != nil {
		t.Errorf("plain to a host the options name was refused: %v", err)
	}
	if _, err := ForSource("http://tiles.example:8080/planet", Options{AllowHTTP: []string{"tiles.example:8080"}}); err != nil {
		t.Errorf("plain to a host and port the options name was refused: %v", err)
	}
	if _, err := ForSource("http://tiles.example:8080/planet", Options{AllowHTTP: []string{"tiles.example:9090"}}); !isKind(err, fault.FetchRefused) {
		t.Errorf("plain to a port the options do not name: %v", err)
	}
	if _, err := ForSource("http://other.example/planet", Options{AllowHTTP: []string{"tiles.example"}}); !isKind(err, fault.FetchRefused) {
		t.Errorf("plain to a host the options do not name: %v", err)
	}
	plain := httptest.NewServer(body("over loopback"))
	defer plain.Close()
	f, err := ForSource(plain.URL, Options{})
	if err != nil {
		t.Fatalf("plain on the loopback address was refused: %v", err)
	}
	if got, err := get(t, f, plain.URL+"/x"); err != nil || string(got) != "over loopback" {
		t.Errorf("%q, %v", got, err)
	}
	// A request to another scheme or host than the source's is refused too.
	if _, err := get(t, f, "https://elsewhere.example/x"); !isKind(err, fault.FetchRefused) {
		t.Errorf("a request outside the source: %v", err)
	}
}

// TestRedirectLimit, TestNoSecureToPlain and
// TestCrossHostRefusedUnlessAllowed are plan task 05.2.
func TestRedirectLimit(t *testing.T) {
	var hops atomic.Int32
	var limit int32
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hops.Add(1) <= limit {
			http.Redirect(w, r, fmt.Sprintf("/hop%d", hops.Load()), http.StatusFound)
			return
		}
		fmt.Fprint(w, "arrived")
	}))
	f, err := ForSource(srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	limit = 3
	if got, err := get(t, f, srv.URL+"/start"); err != nil || string(got) != "arrived" {
		t.Errorf("three redirects: %q, %v", got, err)
	}
	hops.Store(0)
	limit = 4
	if _, err := get(t, f, srv.URL+"/start"); !isKind(err, fault.FetchRefused) {
		t.Errorf("a fourth redirect: %v", err)
	}
}

func TestNoSecureToPlain(t *testing.T) {
	plain := httptest.NewServer(body("downgraded"))
	defer plain.Close()
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL+"/x", http.StatusFound)
	}))
	f, err := ForSource(srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := get(t, f, srv.URL+"/x"); !isKind(err, fault.FetchRefused) {
		t.Errorf("secure to plain: %v", err)
	}
}

func mustHost(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u.Host
}

// TestCrossHostRefused (L-10.1): a redirect to another host is refused, and
// no option allows it - a source's tiles come only from that source's host.
func TestCrossHostRefused(t *testing.T) {
	other, _ := secure(t, body("from the other host"))
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/x", http.StatusFound)
	}))
	opts.RootCAs.AddCert(other.Certificate())
	f, err := ForSource(srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := get(t, f, srv.URL+"/x"); !isKind(err, fault.FetchRefused) {
		t.Errorf("a redirect to another host: %v", err)
	}
}

// TestPrivateAddressRefusedAtDial is plan task 05.3. "localhost" stands for
// any public-looking name: what matters is where the connection lands.
func TestPrivateAddressRefusedAtDial(t *testing.T) {
	srv, opts := secure(t, body("should never be read"))
	named := strings.Replace(srv.URL, "127.0.0.1", "localhost", 1)
	f, err := ForSource(named, opts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := get(t, f, named+"/x"); !isKind(err, fault.FetchRefused) {
		t.Errorf("a name that resolves to the loopback address: %v", err)
	}
	for addr, allowed := range map[string]bool{
		"93.184.216.34:443": true, "[2606:2800:220:1::1]:443": true,
		"127.0.0.1:443": false, "[::1]:443": false, "10.1.2.3:443": false, "172.16.0.9:443": false, "192.168.1.1:443": false,
		"169.254.169.254:80": false, "[fe80::1]:443": false, "[fc00::1]:443": false, "0.0.0.0:443": false, "100.64.0.1:443": false,
		"not-an-address": false,
	} {
		if err := checkDial(addr, false); (err == nil) != allowed {
			t.Errorf("checkDial(%s) = %v; allowed should be %v", addr, err, allowed)
		}
	}
	if err := checkDial("127.0.0.1:8443", true); err != nil {
		t.Errorf("a source that is itself on the loopback address may be reached there: %v", err)
	}
}

// TestNoReferer and TestUserAgent are plan task 05.4.
func TestNoReferer(t *testing.T) {
	var referers []string
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		referers = append(referers, r.Header.Get("Referer"))
		if r.URL.Path == "/a" {
			http.Redirect(w, r, "/b", http.StatusFound)
		}
	}))
	f, _ := ForSource(srv.URL, opts)
	if _, err := get(t, f, srv.URL+"/a"); err != nil {
		t.Fatal(err)
	}
	for _, r := range referers {
		if r != "" {
			t.Errorf("a Referer was sent: %q", r)
		}
	}
}

func TestUserAgent(t *testing.T) {
	var agent string
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { agent = r.Header.Get("User-Agent") }))
	f, _ := ForSource(srv.URL, opts)
	if _, err := get(t, f, srv.URL+"/a"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(agent, "go-tuimaps/") || !strings.Contains(agent, "watchpost-test") {
		t.Errorf("User-Agent %q must name the library, its version and the host's token", agent)
	}
	host, _ := os.Hostname()
	for _, leak := range []string{"Go-http-client", "darwin", "linux", "windows", "arm64", "amd64", host, os.Getenv("USER")} {
		if leak != "" && strings.Contains(strings.ToLower(agent), strings.ToLower(leak)) {
			t.Errorf("User-Agent %q says something about the machine: %q", agent, leak)
		}
	}
	if _, err := ForSource(srv.URL, Options{Token: "bad\r\nX-Injected: 1"}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a token that would inject a header: %v", err)
	}
}

// TestBodyReadThroughLimit is plan task 05.5.
func TestBodyReadThroughLimit(t *testing.T) {
	srv, opts := secure(t, body(strings.Repeat("x", 5000)))
	f, _ := ForSource(srv.URL, opts)
	ctx := context.Background()
	if _, err := f.Fetch(ctx, Request{URL: srv.URL + "/big", MaxBytes: 4999}); !isKind(err, fault.OverLimit) {
		t.Errorf("5000 bytes against a limit of 4999: %v", err)
	}
	if got, err := f.Fetch(ctx, Request{URL: srv.URL + "/big", MaxBytes: 5000}); err != nil || len(got) != 5000 {
		t.Errorf("5000 bytes against a limit of 5000: %d, %v", len(got), err)
	}
	if _, err := f.Fetch(ctx, Request{URL: srv.URL + "/big"}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a request with no limit: %v", err)
	}
}

// TestRangeMustBe206Exact is plan task 05.6.
func TestRangeMustBe206Exact(t *testing.T) {
	var mode string
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch mode {
		case "honest":
			if r.Header.Get("Range") != "bytes=10-19" {
				t.Errorf("Range header %q", r.Header.Get("Range"))
			}
			w.Header().Set("Content-Range", "bytes 10-19/100")
			w.WriteHeader(http.StatusPartialContent)
			fmt.Fprint(w, "0123456789")
		case "whole":
			fmt.Fprint(w, strings.Repeat("y", 100))
		case "wrong range":
			w.Header().Set("Content-Range", "bytes 0-9/100")
			w.WriteHeader(http.StatusPartialContent)
			fmt.Fprint(w, "0123456789")
		case "no content-range":
			w.WriteHeader(http.StatusPartialContent)
			fmt.Fprint(w, "0123456789")
		}
	}))
	f, _ := ForSource(srv.URL, opts)
	req := Request{URL: srv.URL + "/archive", RangeStart: 10, RangeLen: 10, MaxBytes: 10}
	mode = "honest"
	if got, err := f.Fetch(context.Background(), req); err != nil || string(got) != "0123456789" {
		t.Errorf("an honest range reply: %q, %v", got, err)
	}
	for _, mode = range []string{"whole", "wrong range", "no content-range"} {
		if _, err := f.Fetch(context.Background(), req); !isKind(err, fault.FetchFailed) {
			t.Errorf("%s: %v; a range reply that is not a 206 with exactly the range asked for is refused", mode, err)
		}
	}
}

// TestErrorNeverCarriesAddress is plan task 05.7: a tile address may hold a key.
func TestErrorNeverCarriesAddress(t *testing.T) {
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no", http.StatusForbidden)
	}))
	f, _ := ForSource(srv.URL, opts)
	secret := "/6/16/26.pbf?key=SECRET-KEY-123"
	_, err := get(t, f, srv.URL+secret)
	if !isKind(err, fault.FetchFailed) {
		t.Fatalf("%v", err)
	}
	host := mustHost(t, srv.URL)
	if !strings.Contains(err.Error(), "https://"+host) {
		t.Errorf("%q does not name the scheme and host", err)
	}
	for e := err; e != nil; e = errors.Unwrap(e) {
		for _, leak := range []string{"SECRET", "key=", "26.pbf", "/6/16"} {
			if strings.Contains(e.Error(), leak) {
				t.Errorf("%q carries %q", e.Error(), leak)
			}
		}
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		t.Error("the transport's own error, which holds the address, can be reached with errors.As")
	}
	// A transport failure, not a status: the same holds.
	srv.Close()
	_, err = get(t, f, srv.URL+secret)
	if !isKind(err, fault.FetchFailed) || strings.Contains(err.Error(), "SECRET") || errors.As(err, &urlErr) {
		t.Errorf("a failed connection: %v", err)
	}
}

// TestStatusChecked is plan task 05.11.
func TestStatusChecked(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError, http.StatusNoContent, http.StatusTooManyRequests} {
		srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
		f, _ := ForSource(srv.URL, opts)
		_, err := get(t, f, srv.URL+"/x")
		if !isKind(err, fault.FetchFailed) {
			t.Errorf("status %d: %v", status, err)
			continue
		}
		if !strings.Contains(err.Error(), fmt.Sprint(status)) {
			t.Errorf("status %d: %q does not say the status", status, err)
		}
	}
}

// TestTimeout and TestContextCancel are plan task 05.10.
func TestTimeout(t *testing.T) {
	release := make(chan struct{})
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-release }))
	defer close(release)
	opts.Timeout = 150 * time.Millisecond
	f, _ := ForSource(srv.URL, opts)
	start := time.Now()
	_, err := f.Fetch(context.Background(), Request{URL: srv.URL + "/slow", MaxBytes: 10})
	if !isKind(err, fault.FetchFailed) || time.Since(start) > 3*time.Second {
		t.Errorf("a server that never answers: %v after %v", err, time.Since(start))
	}
	plain, err := ForSource(srv.URL, Options{RootCAs: opts.RootCAs})
	if err != nil || plain.client.Timeout != DefaultTimeout {
		t.Errorf("the default timeout is %v, %v; want %v", plain.client.Timeout, err, DefaultTimeout)
	}
}

func TestContextCancel(t *testing.T) {
	release := make(chan struct{})
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-release }))
	defer close(release)
	f, _ := ForSource(srv.URL, opts)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := f.Fetch(ctx, Request{URL: srv.URL + "/slow", MaxBytes: 10})
		done <- err
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !isKind(err, fault.Cancelled) || !errors.Is(err, context.Canceled) {
			t.Errorf("%v; a cancelled fetch is the cancelled kind, and answers errors.Is", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Get did not return after its context was cancelled")
	}
}

// TestProxyFromEnvironment is plan task 05.8.
func TestProxyFromEnvironment(t *testing.T) {
	f, err := ForSource("https://tiles.example", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if f.transport.Proxy == nil {
		t.Fatal("no proxy function: the environment's proxy settings are not honoured")
	}
	t.Setenv("HTTPS_PROXY", "http://proxy.example:3128")
	req, _ := http.NewRequest(http.MethodGet, "https://tiles.example/x", nil)
	u, err := f.transport.Proxy(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = u // the standard library reads the environment once a process; that it is wired is what is tested
}

// transport is a host's own transport in a function.
type transport func(*http.Request) (*http.Response, error)

func (t transport) RoundTrip(r *http.Request) (*http.Response, error) { return t(r) }

func reply(r *http.Request, status int, body io.Reader) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(body), Header: http.Header{}, Request: r}
}

// endless is a body that never ends, counting what is read of it.
type endless struct{ read int64 }

func (e *endless) Read(p []byte) (int, error) {
	e.read += int64(len(p))
	return len(p), nil
}

// TestAHostTransport is v0.2.0 L8.1 and L8.3 (L-7.1, L-7.3, D-55): a host's
// transport under the library's client. The library still shapes the
// request; bytes are bounded whatever the transport sends; an answer the
// transport brings back after the request's end is never used, though a
// transport that ignores its context holds its own call until it returns;
// and the transport's error, which may hold the address, is never passed on.
func TestAHostTransport(t *testing.T) {
	var seen *http.Request
	honest := transport(func(r *http.Request) (*http.Response, error) {
		seen = r
		resp := reply(r, http.StatusPartialContent, strings.NewReader(strings.Repeat("x", 16)))
		resp.Header.Set("Content-Range", "bytes 100-115/1000")
		return resp, nil
	})
	f, err := ForSource("https://tiles.example/", Options{Transport: honest, Token: "host"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := f.Fetch(context.Background(), Request{URL: "https://tiles.example/archive", RangeStart: 100, RangeLen: 16, MaxBytes: 16})
	if err != nil || len(got) != 16 {
		t.Fatalf("an honest transport: %d bytes, %v", len(got), err)
	}
	if seen.Header.Get("Range") != "bytes=100-115" || !strings.Contains(seen.Header.Get("User-Agent"), "(host)") {
		t.Errorf("the library did not shape the request: %v", seen.Header)
	}

	body := &endless{}
	flood := transport(func(r *http.Request) (*http.Response, error) { return reply(r, http.StatusOK, body), nil })
	f, _ = ForSource("https://tiles.example/", Options{Transport: flood})
	if _, err := f.Fetch(context.Background(), Request{URL: "https://tiles.example/x", MaxBytes: 1 << 16}); !isKind(err, fault.OverLimit) {
		t.Errorf("an endless body: %v; want it cut at the limit", err)
	}
	if body.read > 1<<20 {
		t.Errorf("an endless body was read to %d bytes; the limit is 64 KiB", body.read)
	}

	stubborn := transport(func(r *http.Request) (*http.Response, error) {
		time.Sleep(150 * time.Millisecond) // ignores its context
		return reply(r, http.StatusOK, strings.NewReader("too late")), nil
	})
	f, _ = ForSource("https://tiles.example/", Options{Transport: stubborn})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	data, err := f.Fetch(ctx, Request{URL: "https://tiles.example/x", MaxBytes: 1 << 10})
	if err == nil || data != nil {
		t.Errorf("a late answer was used: %q, %v", data, err)
	}
	if took := time.Since(start); took < 100*time.Millisecond {
		t.Errorf("the call returned after %v, before the transport did: that cannot be, and the contract says it holds its own call", took)
	}

	leaky := transport(func(r *http.Request) (*http.Response, error) { return nil, fmt.Errorf("GET %s failed", r.URL) })
	f, _ = ForSource("https://tiles.example/", Options{Transport: leaky})
	_, err = f.Fetch(context.Background(), Request{URL: "https://tiles.example/x?key=SECRET", MaxBytes: 10})
	if !isKind(err, fault.FetchFailed) || strings.Contains(err.Error(), "SECRET") || errors.Unwrap(err) != nil {
		t.Errorf("a transport's error must not be passed on: %v", err)
	}
}

// TestReservedRangesRefused is v0.2.0 L8.6's second half (L-10.3): each of the
// ranges added in v0.2.0 is refused at the moment of connection.
func TestReservedRangesRefused(t *testing.T) {
	for _, address := range []string{"0.0.0.1:443", "[64:ff9b::a00:1]:443", "[2002:a00:1::1]:443", "198.18.0.1:443", "198.19.255.254:443", "240.0.0.1:443", "255.255.255.254:443"} {
		if err := checkDial(address, false); !errors.Is(err, errPolicy) {
			t.Errorf("%s: %v; want it refused", address, err)
		}
	}
	for _, address := range []string{"8.8.8.8:443", "[2606:4700::1111]:443", "198.20.0.1:443"} {
		if err := checkDial(address, false); err != nil {
			t.Errorf("%s, a public address: %v", address, err)
		}
	}
}

// TestTheProxyDecisionIsPerConnection is L8.6 (L-10.3): the address check is
// skipped only for a connection to the proxy itself. A redirect to a host the
// proxy does not carry is dialled directly, and meets the check.
func TestTheProxyDecisionIsPerConnection(t *testing.T) {
	server := httptest.NewServer(body("the proxy"))
	defer server.Close()
	proxyAddress := strings.TrimPrefix(server.URL, "http://")
	f, err := ForSource("https://tiles.example/", Options{})
	if err != nil {
		t.Fatal(err)
	}
	f.proxyOf = func(r *http.Request) (*url.URL, error) {
		if r.URL.Hostname() == "tiles.example" {
			return url.Parse("http://" + proxyAddress)
		}
		return nil, nil // a host the proxy settings except
	}
	// Before any request has gone through the proxy, a loopback address is
	// refused like any private one.
	if conn, err := f.dial(context.Background(), "tcp", proxyAddress); err == nil {
		conn.Close()
		t.Fatal("a loopback address was dialled with no proxy in use")
	}
	if u, _ := f.proxy(httptest.NewRequest(http.MethodGet, "https://tiles.example/x", nil)); u == nil {
		t.Fatal("the source's request was not proxied")
	}
	conn, err := f.dial(context.Background(), "tcp", proxyAddress)
	if err != nil {
		t.Errorf("the connection to the proxy itself was refused: %v", err)
	} else {
		conn.Close()
	}
	if u, _ := f.proxy(httptest.NewRequest(http.MethodGet, "https://internal.example/x", nil)); u != nil {
		t.Fatal("a host the settings except was proxied")
	}
	if conn, err := f.dial(context.Background(), "tcp", "10.0.0.1:443"); err == nil {
		conn.Close()
		t.Error("a redirect's direct connection to a private address skipped the check")
	}
}
