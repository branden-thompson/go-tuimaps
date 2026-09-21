package testkit

import (
	"crypto/x509"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Link describes the network a ShapedServer stands for. Nothing waits: the
// cost of each request is added to a virtual clock, so a timing measured
// against it is the same on every machine and every run (NFR-5).
type Link struct {
	Setup         time.Duration // making the connection, paid once
	RoundTrip     time.Duration // paid once a round of requests
	BitsPerSecond int64         // what the body's bytes are charged at
	Width         int           // how many requests travel together in one round: the pump's width
}

// ShapedServer is a loopback server with secure transport, its own trust
// anchor, and a virtual clock charged by a Link.
type ShapedServer struct {
	// URL is the server's address, on the loopback interface.
	URL string
	// RootCAs is the server's own trust anchor; a client without it is refused.
	RootCAs *x509.CertPool

	srv  *httptest.Server
	link Link

	mu       sync.Mutex
	requests int
	bytes    int64
}

// ShapedStats is what a ShapedServer has served, and what its link would
// have charged for it.
type ShapedStats struct {
	Requests int
	Bytes    int64
	// VirtualElapsed is the connection once, a round trip for each round of
	// Width requests, and every body byte at the link's bandwidth. It is
	// arithmetic, not a measurement.
	VirtualElapsed time.Duration
}

// NewShapedServer starts a server for handler over link.
func NewShapedServer(handler http.Handler, link Link) (*ShapedServer, error) {
	if handler == nil {
		return nil, errors.New("testkit: a shaped server needs a handler")
	}
	if link.BitsPerSecond <= 0 || link.Width <= 0 {
		return nil, errors.New("testkit: a link needs a bandwidth and a width")
	}
	if link.Setup < 0 || link.RoundTrip < 0 {
		return nil, errors.New("testkit: a link's delays cannot be negative")
	}
	s := &ShapedServer{link: link}
	s.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The handler writes to a recorder, so its body can be counted
		// before it is passed on.
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		n, _ := w.Write(rec.Body.Bytes())
		s.mu.Lock()
		s.requests++
		s.bytes += int64(n)
		s.mu.Unlock()
	}))
	s.URL = s.srv.URL
	s.RootCAs = x509.NewCertPool()
	s.RootCAs.AddCert(s.srv.Certificate())
	return s, nil
}

// Shutdown stops the server.
func (s *ShapedServer) Shutdown() {
	if s == nil || s.srv == nil {
		return
	}
	s.srv.Close()
}

// Stats reports what has been served and what the link would have charged.
func (s *ShapedServer) Stats() ShapedStats {
	if s == nil {
		return ShapedStats{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.requests == 0 {
		return ShapedStats{}
	}
	rounds := (s.requests + s.link.Width - 1) / s.link.Width
	transfer := time.Duration(s.bytes*8) * time.Second / time.Duration(s.link.BitsPerSecond)
	return ShapedStats{Requests: s.requests, Bytes: s.bytes, VirtualElapsed: s.link.Setup + time.Duration(rounds)*s.link.RoundTrip + transfer}
}
