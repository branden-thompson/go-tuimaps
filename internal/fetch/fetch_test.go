package fetch

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
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
	if _, err := ForSource("ftp://tiles.example/planet", Options{AllowPlainHTTP: true}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a scheme that is neither: %v", err)
	}
	if _, err := ForSource("https://user:secret@tiles.example/planet", Options{}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a source with a user-info part: %v", err)
	}
	if _, err := ForSource("http://tiles.example/planet", Options{AllowPlainHTTP: true}); err != nil {
		t.Errorf("plain by explicit option was refused: %v", err)
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
	opts.AllowHosts = []string{mustHost(t, plain.URL)}
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

func TestCrossHostRefusedUnlessAllowed(t *testing.T) {
	other, otherOpts := secure(t, body("from the other host"))
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/x", http.StatusFound)
	}))
	opts.RootCAs.AddCert(other.Certificate())
	_ = otherOpts
	f, err := ForSource(srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := get(t, f, srv.URL+"/x"); !isKind(err, fault.FetchRefused) {
		t.Errorf("a redirect to another host: %v", err)
	}
	opts.AllowHosts = []string{mustHost(t, other.URL)}
	f, err = ForSource(srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := get(t, f, srv.URL+"/x"); err != nil || string(got) != "from the other host" {
		t.Errorf("a redirect to an allowed host: %q, %v", got, err)
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

// TestReplacementFetcherContract is plan task 05.9.
func TestReplacementFetcherContract(t *testing.T) {
	var seen Request
	honest := func(ctx context.Context, r Request) ([]byte, error) { seen = r; return make([]byte, r.RangeLen), nil }
	req := Request{URL: "https://tiles.example/archive", RangeStart: 100, RangeLen: 16, MaxBytes: 16}
	got, err := Checked(honest)(context.Background(), req)
	if err != nil || len(got) != 16 || seen != req {
		t.Errorf("the range and the maximum length must reach the replacement: %+v, %d, %v", seen, len(got), err)
	}
	tooMuch := func(ctx context.Context, r Request) ([]byte, error) { return make([]byte, r.MaxBytes+1), nil }
	if _, err := Checked(tooMuch)(context.Background(), req); !isKind(err, fault.OverLimit) {
		t.Errorf("a replacement that returns more than the maximum: %v", err)
	}
	short := func(ctx context.Context, r Request) ([]byte, error) { return make([]byte, 5), nil }
	if _, err := Checked(short)(context.Background(), req); !isKind(err, fault.FetchFailed) {
		t.Errorf("a replacement that returns a different length than the range: %v", err)
	}
	leaky := func(ctx context.Context, r Request) ([]byte, error) { return nil, fmt.Errorf("GET %s failed", r.URL) }
	_, err = Checked(leaky)(context.Background(), Request{URL: "https://tiles.example/x?key=SECRET", MaxBytes: 10})
	if !isKind(err, fault.FetchFailed) || strings.Contains(err.Error(), "SECRET") || errors.Unwrap(err) != nil {
		t.Errorf("a replacement's error must not be passed on: %v", err)
	}
	if Checked(nil) != nil {
		t.Error("no replacement gives no fetcher")
	}
}
