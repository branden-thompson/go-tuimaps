package fetch

import (
	"context"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// A LIMIT AT ITS EDGE NEVER WRAPS: a size limit at the largest int64 would
// make the read's limit of one more byte negative, and read nothing as a
// success; a range reaching past the largest int64 would ask for a range
// that is not the one meant. Both are refused before anything is sent.
func TestLimitsAtTheirEdgesNeverWrap(t *testing.T) {
	srv, opts := secure(t, body("hello"))
	f, err := ForSource(srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if got, err := f.Fetch(ctx, Request{URL: srv.URL + "/a", MaxBytes: math.MaxInt64}); err == nil && len(got) == 0 {
		t.Error("a limit of the largest int64 read nothing and called it the body")
	}
	if _, err := f.Fetch(ctx, Request{URL: srv.URL + "/a", MaxBytes: 16, RangeStart: math.MaxInt64 - 1, RangeLen: 5}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a range past the largest int64: %v; want refused", err)
	}
}

// A FETCHER WITH NOTHING BEHIND IT REFUSES: the zero Fetcher - the one the
// checked dialer is made from - has no client to send with, and a request
// through it is refused rather than a panic.
func TestAFetcherWithNoClientRefuses(t *testing.T) {
	var f Fetcher
	if _, err := f.Fetch(context.Background(), Request{URL: "https://example.com/a", MaxBytes: 16}); !isKind(err, fault.FetchRefused) {
		t.Errorf("the zero Fetcher: %v; want refused", err)
	}
}

// NO ADDRESS IS NO ANSWER: the helpers that read an address answer a missing
// one plainly rather than dereferencing it.
func TestAMissingAddressIsAnsweredPlainly(t *testing.T) {
	if listed(nil, []string{"example.com"}) {
		t.Error("no address is listed")
	}
	if got := proxyAddress(nil); got != "" {
		t.Errorf("no proxy has the address %q", got)
	}
	var f *Fetcher
	if f.allowed(&url.URL{Scheme: "https", Host: "example.com"}) {
		t.Error("no fetcher allows an address")
	}
}

// NO FETCHER, NO PANIC: every method the client and the dialer call answers
// a nil receiver with a refusal, and an error made without a context or a
// fetcher is still the library's own.
func TestANilFetcherRefusesEverywhere(t *testing.T) {
	var f *Fetcher
	ctx := context.Background()
	if err := f.problem(fault.FetchFailed, textsafe.Text{}); err == nil {
		t.Error("no fetcher made no error")
	}
	if err := f.checkRedirect(&http.Request{URL: &url.URL{Scheme: "https", Host: "example.com"}}, nil); err == nil {
		t.Error("no fetcher allowed a redirect")
	}
	if _, err := f.proxy(&http.Request{}); err == nil {
		t.Error("no fetcher named a proxy")
	}
	if _, err := f.dial(ctx, "tcp", "example.com:443"); err == nil {
		t.Error("no fetcher dialled")
	}
	if _, err := f.send(ctx, Request{URL: "https://example.com/a", MaxBytes: 16}); err == nil {
		t.Error("no fetcher sent")
	}
	g := &Fetcher{scheme: "https", host: "example.com"}
	var none context.Context
	if err := g.transportError(none, errors.New("x")); !isKind(err, fault.FetchFailed) {
		t.Errorf("an error with no context: %v", err)
	}
}

// A MALFORMED ADDRESS IS NOT PUBLIC: an IP of neither four nor sixteen bytes
// answers none of the private-space questions, and must not fall through to
// "public".
func TestAMalformedAddressIsNotPublic(t *testing.T) {
	if isPublic(net.IP{1, 2, 3}) {
		t.Error("a three-byte address is public")
	}
}

// EACH LIMIT IS HELD WHERE IT IS USED: send refuses a limit that would wrap
// before anything goes out, and readReply refuses it before it reads, so
// neither depends on the other having run.
func TestEachEdgeIsHeldWhereItIsUsed(t *testing.T) {
	g := &Fetcher{scheme: "https", host: "example.com", client: &http.Client{}}
	ctx := context.Background()
	if _, err := g.send(ctx, Request{URL: "https://example.com/a", MaxBytes: math.MaxInt64}); !isKind(err, fault.FetchRefused) {
		t.Errorf("send with a limit of the largest int64: %v; want refused", err)
	}
	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("hello"))}
	if _, err := g.readReply(ctx, resp, Request{MaxBytes: math.MaxInt64}); !isKind(err, fault.FetchRefused) {
		t.Errorf("readReply with a limit of the largest int64: %v; want refused", err)
	}
	noClient := &Fetcher{scheme: "https", host: "example.com"}
	if _, err := noClient.send(ctx, Request{URL: "https://example.com/a", MaxBytes: 16}); !isKind(err, fault.FetchRefused) {
		t.Errorf("a fetcher with no client sent: %v; want refused", err)
	}
	if err := g.transportError(ctx, nil); err != nil {
		t.Errorf("no error was translated into %v", err)
	}
}

// TestTheLesserReservedRangesAreRefused (L-10.3; BUILD-exit red team, InfoSec
// S-5): local-use NAT64, which can carry a private IPv4 address inside it,
// deprecated site-local, Teredo, and the ranges set aside for documentation
// and protocol use are no public source's address.
func TestTheLesserReservedRangesAreRefused(t *testing.T) {
	for _, address := range []string{"[64:ff9b:1::a00:1]:443", "[fec0::1]:443", "[2001::1]:443", "[2001:db8::1]:443",
		"192.0.0.8:443", "192.0.2.1:443", "198.51.100.1:443", "203.0.113.1:443",
		"[::7f00:1]:443", "[::a00:1]:443", "[::ffff:0:a00:1]:443", "[100::1]:443"} { // IPv4-compatible, IPv4-translated, discard-only (REVIEW, InfoSec F6)
		if err := checkDial(address, false); !errors.Is(err, errPolicy) {
			t.Errorf("%s: %v; want it refused", address, err)
		}
	}
	for _, address := range []string{"192.0.3.1:443", "[2001:4860::8888]:443", "203.0.114.1:443"} {
		if err := checkDial(address, false); err != nil {
			t.Errorf("%s, a public address: %v", address, err)
		}
	}
}
