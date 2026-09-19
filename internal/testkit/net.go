package testkit

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
)

// ErrNotLoopback is the refusal every test dial to a non-loopback address
// receives. Tests reach loopback only (NFR-11).
var ErrNotLoopback = errors.New("testkit: tests may connect to loopback only")

// CheckLoopback passes a literal loopback address with a port and refuses
// everything else: names, because a name can resolve anywhere, and every
// other address.
func CheckLoopback(address string) error {
	if address == "" {
		return fmt.Errorf("%w: empty address", ErrNotLoopback)
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%w: %q is not host:port", ErrNotLoopback, address)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("%w: %q is a name, not a literal address", ErrNotLoopback, host)
	}
	if !ip.IsLoopback() {
		return fmt.Errorf("%w: %s", ErrNotLoopback, host)
	}
	return nil
}

// Dialer returns a dialer that refuses any address but loopback. The check
// runs on the address actually being connected to, after name resolution
// and before any packet is sent.
func Dialer() *net.Dialer {
	return &net.Dialer{
		Control: func(_, address string, _ syscall.RawConn) error {
			return CheckLoopback(address)
		},
	}
}

// InstallLoopbackOnly puts the loopback-only dialer under the process's
// default HTTP transport and returns the call that undoes it.
func InstallLoopbackOnly() (restore func(), err error) {
	tr, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("testkit: the default transport has been replaced; cannot guard it")
	}
	if tr == nil {
		return nil, errors.New("testkit: the default transport is nil")
	}
	old := tr.DialContext
	tr.DialContext = Dialer().DialContext
	tr.CloseIdleConnections()
	return func() { tr.DialContext = old; tr.CloseIdleConnections() }, nil
}

// BlockingTransport is an HTTP transport whose requests never complete: a
// request returns only when its context ends. It proves that a caller —
// Render above all — never waits on the network (FR-23).
type BlockingTransport struct{}

// RoundTrip blocks until the request's context ends and returns the
// context's own error.
func (BlockingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("testkit: nil request")
	}
	if req.URL == nil {
		return nil, errors.New("testkit: request with no address")
	}
	ctx := req.Context()
	<-ctx.Done()
	return nil, context.Cause(ctx)
}
