// Package fetch is the one door to the network (FR-22b). It opens a
// connection only for a source a host has named; it insists on secure
// transport, follows at most three redirects and never to a plain or a
// stranger's address, refuses to land in private address space, sends no
// Referer and says nothing about the machine, reads every body through a
// limit, and returns errors that name a scheme and a host and never an
// address - a tile address may hold a key.
package fetch

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"
)

const (
	// Version names the library in the User-Agent.
	Version = "0.1.0-dev"
	// DefaultTimeout bounds a whole request (constants, section 5).
	DefaultTimeout = 20 * time.Second
	// firstByteTimeout bounds the wait for a connection and for the first
	// byte of a reply.
	firstByteTimeout = 10 * time.Second
	// maxRedirects is how many redirects one request may follow.
	maxRedirects = 3
	// maxToken bounds the host's token in the User-Agent.
	maxToken = 64
)

// HostOptions are how a host shapes the library's fetching, as the root
// package's FetchOptions (D-55).
type HostOptions struct {
	// Transport is the host's own: its proxy, its trust roots, its pool, or
	// a store it serves tiles from. Nil is the library's, which refuses to
	// connect to a private address.
	Transport http.RoundTripper
	// UserAgent is the host's name for itself, added to the library's own
	// (L-7.2): printable ASCII, no spaces or brackets, at most 64 bytes.
	UserAgent string
	// Timeout bounds a whole request; zero is the library's default.
	Timeout time.Duration
	// AllowHTTP names hosts, as host or host:port, that may be fetched from
	// over plain http; without them only a literal loopback address may
	// (L-10.2).
	AllowHTTP []string
}

// Options are a host's choices for one source.
type Options struct {
	// AllowHTTP names hosts, as host or host:port, that may be fetched from
	// over plain http. Without them plain http is permitted only to a literal
	// loopback address (L-10.2).
	AllowHTTP []string
	// Token is the host's name for itself in the User-Agent (L-7.2).
	Token string
	// Timeout bounds a whole request; zero means DefaultTimeout.
	Timeout time.Duration
	// RootCAs are the trust anchors; nil means the system's.
	RootCAs *x509.CertPool
	// Transport is a host's own, under the library's client (D-55): the
	// library still shapes each request, confines redirects, sets the
	// headers and reads the body through its limit. Nil is the library's
	// own, with its private-address check; a host transport dials as it
	// likes, and keeps that check only by dialling through CheckedDialer.
	Transport http.RoundTripper
}

// Request asks for one address: all of it, or a range of it. MaxBytes is
// required: no body is read without a limit.
type Request struct {
	URL        string
	RangeStart int64
	RangeLen   int64 // zero means the whole body
	MaxBytes   int64
}

// Func fetches one request. It is the shape of the library's own fetcher
// and of a host's replacement.
type Func func(ctx context.Context, r Request) ([]byte, error)

// Fetcher fetches from one named source.
type Fetcher struct {
	scheme    string
	host      string
	private   bool // the source is itself in private or loopback space, so connections may land there
	agent     string
	client    *http.Client
	transport *http.Transport
}

// viaProxy marks a request that goes through a proxy from the environment.
// The connection then lands on the proxy, so the check on the connected
// address cannot be made; FR-22b documents this.
type viaProxy struct{}

// errPolicy marks a refusal made inside the HTTP client - a redirect or a
// dial - so that Get can tell it from a failure of the network.
var errPolicy = errors.New("fetch: refused by policy")

// ForSource makes a fetcher for one source, named by its address. Nothing
// is connected to until Fetch is called.
func ForSource(source string, opts Options) (*Fetcher, error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, refusedSource()
	}
	if u.Host == "" || u.User != nil {
		return nil, refusedSource()
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, refusedSource()
	}
	ip := net.ParseIP(u.Hostname())
	if u.Scheme == "http" && !ip.IsLoopback() && !listed(u, opts.AllowHTTP) {
		return nil, refusedSource()
	}
	agent, err := userAgent(opts.Token)
	if err != nil {
		return nil, err
	}
	f := &Fetcher{scheme: u.Scheme, host: u.Host, private: ip != nil && !isPublic(ip), agent: agent}
	f.transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           f.dial,
		TLSClientConfig:       &tls.Config{RootCAs: opts.RootCAs, MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout:   firstByteTimeout,
		ResponseHeaderTimeout: firstByteTimeout,
		ForceAttemptHTTP2:     true,
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	var through http.RoundTripper = f.transport
	if opts.Transport != nil {
		through = opts.Transport
	}
	f.client = &http.Client{Transport: through, Timeout: timeout, CheckRedirect: f.checkRedirect}
	return f, nil
}

// listed reports whether an address's host is one the options name, as
// host or host:port.
func listed(u *url.URL, hosts []string) bool {
	for _, h := range hosts {
		if h == u.Host || h == u.Hostname() {
			return true
		}
	}
	return false
}

// CheckToken reports whether a host's name for itself may go in the
// User-Agent, before any source is named.
func CheckToken(token string) error {
	_, err := userAgent(token)
	return err
}

// userAgent names the library, its version and the host's token, and
// nothing about the machine. A token that could break out of the header is
// refused.
func userAgent(token string) (string, error) {
	if len(token) > maxToken {
		return "", refusedSource()
	}
	for _, r := range token {
		if r <= ' ' || r > '~' || r == '(' || r == ')' {
			return "", refusedSource()
		}
	}
	if token == "" {
		return "go-tuimaps/" + Version, nil
	}
	return "go-tuimaps/" + Version + " (" + token + ")", nil
}

// isPublic reports whether ip is an address a named public source may be
// reached at: not loopback, not link-local, not private, not unspecified,
// not the shared address space of carrier-grade translation.
func isPublic(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
		return false
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return false
	}
	_, shared, err := net.ParseCIDR("100.64.0.0/10")
	if err != nil {
		return false
	}
	return !shared.Contains(ip)
}

// checkDial is the check on the address actually being connected to, after
// name resolution and before any packet: a name that resolves into private
// space is refused here, whatever it looked like.
func checkDial(address string, sourceIsPrivate bool) error {
	if sourceIsPrivate {
		return nil
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return errPolicy
	}
	if !isPublic(net.ParseIP(host)) {
		return errPolicy
	}
	return nil
}

// dial connects for the transport, with the address check unless the
// request goes through a proxy.
func (f *Fetcher) dial(ctx context.Context, network, address string) (net.Conn, error) {
	if network == "" || address == "" {
		return nil, errPolicy
	}
	d := net.Dialer{Timeout: firstByteTimeout}
	if ctx.Value(viaProxy{}) == nil {
		d.Control = func(_, connected string, _ syscall.RawConn) error { return checkDial(connected, f.private) }
	}
	return d.DialContext(ctx, network, address)
}

// allowed reports whether a request or a redirect may go to this address:
// the source's own scheme, host and port, and nothing else (L-10.1).
func (f *Fetcher) allowed(u *url.URL) bool {
	return u != nil && u.User == nil && u.Scheme == f.scheme && u.Host == f.host
}

// checkRedirect is the redirect policy: at most three, never from secure to
// plain, never to a host that is neither the source's nor allowed. The
// Referer the standard client adds is removed.
func (f *Fetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	if req == nil || req.URL == nil {
		return errPolicy
	}
	if len(via) > maxRedirects {
		return errPolicy
	}
	if !f.allowed(req.URL) {
		return errPolicy
	}
	req.Header.Del("Referer")
	return nil
}
