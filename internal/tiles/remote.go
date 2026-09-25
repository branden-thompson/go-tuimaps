package tiles

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/fetch"
	"github.com/branden-thompson/go-tuimaps/internal/jsonsafe"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// maxTileJSONBytes and maxTileJSONDepth bound a TileJSON document (NFR-10).
	maxTileJSONBytes = 1 << 20
	maxTileJSONDepth = 64
	// maxTileBytes bounds a tile as received, as the decoder's own limit does.
	maxTileBytes = 2 << 20
)

// Template builds a tile's address from a pattern in which only {z}, {x}
// and {y} are filled.
type Template struct {
	parts []string // literal text, alternating with one of "z", "x", "y"
}

// URL is the address of one tile.
func (t Template) URL(tile scene.TileID) string {
	if len(t.parts) == 0 {
		return ""
	}
	var b strings.Builder
	for i, part := range t.parts {
		switch {
		case i%2 == 0:
			b.WriteString(part)
		case part == "z":
			b.WriteString(strconv.Itoa(int(tile.Z)))
		case part == "x":
			b.WriteString(strconv.FormatUint(uint64(tile.X), 10))
		default:
			b.WriteString(strconv.FormatUint(uint64(tile.Y), 10))
		}
	}
	return b.String()
}

// refusedAddress never repeats the address: it may hold a key.
func refusedAddress(why textsafe.Text) error {
	return fault.Make(fault.FetchRefused, textsafe.Const("the tile source was refused"), why,
		textsafe.Const("name a secure source whose tiles come from its own host, or allow the other host explicitly"))
}

// parseTemplate splits a pattern at its tokens. Anything in braces other
// than z, x and y refuses the source, and so does a brace left open.
func parseTemplate(raw string) (Template, error) {
	if raw == "" || len(raw) > 2048 {
		return Template{}, refusedAddress(textsafe.Const("its tile address is empty or longer than 2,048 bytes"))
	}
	var parts []string
	seen := map[string]bool{}
	rest := raw
	for range len(raw) {
		open := strings.IndexAny(rest, "{}")
		if open < 0 {
			break
		}
		end := strings.IndexByte(rest, '}')
		if rest[open] == '}' || end < 0 {
			return Template{}, refusedAddress(textsafe.Const("its tile address has a brace that is not part of a token"))
		}
		token := rest[open+1 : end]
		if token != "z" && token != "x" && token != "y" {
			return Template{}, refusedAddress(textsafe.Const("its tile address has a token other than {z}, {x} and {y}"))
		}
		seen[token] = true
		parts = append(parts, rest[:open], token)
		rest = rest[end+1:]
	}
	if len(seen) != 3 {
		return Template{}, refusedAddress(textsafe.Const("its tile address does not have all of {z}, {x} and {y}"))
	}
	return Template{parts: append(parts, rest)}, nil
}

// checkAddress holds a tile address to the fetch rules (FR-22b): a secure
// scheme, or the source's own if the host allowed that; no user name; the
// source's own host unless the application allowed another; no token in
// the scheme or host.
func checkAddress(raw string, source *url.URL, allowHosts []string) error {
	if source == nil {
		return refusedAddress(textsafe.Const("there is no source to hold its tile address to"))
	}
	u, err := url.Parse(strings.NewReplacer("{z}", "0", "{x}", "0", "{y}", "0").Replace(raw))
	if err != nil {
		return refusedAddress(textsafe.Const("its tile address is not an address"))
	}
	if u.Scheme != "https" && u.Scheme != source.Scheme {
		return refusedAddress(textsafe.Const("its tile address is not secure"))
	}
	if u.User != nil || u.Host == "" {
		return refusedAddress(textsafe.Const("its tile address carries a user name, or has no host"))
	}
	// The address is checked as it is parsed but used as it was written, so
	// the two must be the same text: a scheme written "Https" parses as
	// https and passes every check above, while every tile the library then
	// asks for carries the odd spelling.
	if !strings.HasPrefix(raw, u.Scheme+"://") {
		return refusedAddress(textsafe.Const("its tile address is not written as it parses: the scheme must be lower case"))
	}
	if !strings.HasPrefix(u.EscapedPath(), "/") {
		// "https://host?/{z}{x}{y}" parses with the tile numbers in its query
		// and no path at all: found by FuzzTileJSON.
		return refusedAddress(textsafe.Const("its tile address has no path after its host"))
	}
	if strings.Contains(raw, "#") {
		return refusedAddress(textsafe.Const("its tile address has a fragment, which is never sent: every tile would be the same request"))
	}
	if at := strings.Index(raw, "://"); at < 0 || strings.ContainsAny(strings.SplitN(raw[at+3:], "/", 2)[0], "{}") {
		return refusedAddress(textsafe.Const("its tile address has a token in its host"))
	}
	if u.Host == source.Host {
		return nil
	}
	for _, h := range allowHosts {
		if h == u.Host {
			return nil
		}
	}
	return refusedAddress(textsafe.Const("its tiles come from a host other than its own, which the application has not allowed"))
}

// TileJSON is what the library reads of a TileJSON document.
type TileJSON struct {
	Template         Template
	MinZoom, MaxZoom uint8
	Layers           []string      // the layer ids it declares
	Attribution      textsafe.Text // outside text, cleaned
}

func overLimit() error {
	return fault.Make(fault.OverLimit, textsafe.Const("the tile source was refused"),
		textsafe.Const("its TileJSON is larger than 1 MiB or nested deeper than 64"),
		textsafe.Const("check the source; the limits protect the host's memory"))
}

// zoom reads an optional zoom from a TileJSON number.
func zoom(n *float64, otherwise uint8) (uint8, bool) {
	if n == nil {
		return otherwise, true
	}
	if !(*n >= 0 && *n <= scene.MaxTileZoom) {
		return 0, false
	}
	return uint8(*n), true
}

// ParseTileJSON reads a TileJSON document fetched from source. Its first
// tile address is used, as upstream does (P-49), held to the fetch rules.
func ParseTileJSON(body []byte, source string, allowHosts []string) (TileJSON, error) {
	if len(body) == 0 || source == "" {
		return TileJSON{}, refusedAddress(textsafe.Const("it sent no TileJSON, or no source was named"))
	}
	if !jsonsafe.Within(body, maxTileJSONBytes, maxTileJSONDepth) {
		return TileJSON{}, overLimit()
	}
	from, err := url.Parse(source)
	if err != nil {
		return TileJSON{}, refusedAddress(textsafe.Const("the source is not an address"))
	}
	var doc struct {
		Tiles        []string `json:"tiles"`
		MinZoom      *float64 `json:"minzoom"`
		MaxZoom      *float64 `json:"maxzoom"`
		Attribution  string   `json:"attribution"`
		VectorLayers []struct {
			ID string `json:"id"`
		} `json:"vector_layers"`
	}
	err = json.Unmarshal(body, &doc)
	if err != nil || len(doc.Tiles) == 0 {
		return TileJSON{}, refusedAddress(textsafe.Const("what it sent is not TileJSON with a tile address"))
	}
	err = checkAddress(doc.Tiles[0], from, allowHosts)
	if err != nil {
		return TileJSON{}, err
	}
	tmpl, err := parseTemplate(doc.Tiles[0])
	if err != nil {
		return TileJSON{}, err
	}
	lo, okLo := zoom(doc.MinZoom, 0)
	hi, okHi := zoom(doc.MaxZoom, defaultMaxZoom)
	if !okLo || !okHi || lo > hi {
		return TileJSON{}, refusedAddress(textsafe.Const("its TileJSON gives zooms that are out of order or deeper than any tile"))
	}
	info := TileJSON{Template: tmpl, MinZoom: lo, MaxZoom: hi, Attribution: textsafe.Clean(doc.Attribution)}
	for _, l := range doc.VectorLayers {
		info.Layers = append(info.Layers, l.ID)
	}
	_, err = ResolveSchema(info.Layers)
	if err != nil {
		return TileJSON{}, err
	}
	return info, nil
}

// Remote is a named network source in either of its two modes (P-49): a
// source ending in a slash is a prefix, and any other is a TileJSON
// document, read once, inside the first tile job that needs it.
type Remote struct {
	source string
	get    fetch.Func
	allow  []string

	mu      sync.Mutex // never held across a fetch
	ready   bool
	info    TileJSON
	refusal error
}

// NewRemote names a source. It reaches nothing.
func NewRemote(source string, get fetch.Func, allowHosts []string) (*Remote, error) {
	if get == nil {
		return nil, refusedOptions(textsafe.Const("a named source was given nothing to fetch with"))
	}
	if source == "" {
		return nil, refusedAddress(textsafe.Const("no source was named"))
	}
	u, err := url.Parse(source)
	if err != nil {
		return nil, refusedAddress(textsafe.Const("the source is not an address"))
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil {
		return nil, refusedAddress(textsafe.Const("the source is not an http address with a host, or it carries a user name"))
	}
	r := &Remote{source: source, get: get, allow: append([]string(nil), allowHosts...)}
	if strings.HasSuffix(source, "/") {
		tmpl, err := parseTemplate(source + "{z}/{x}/{y}.pbf")
		if err != nil {
			return nil, err
		}
		r.ready, r.info = true, TileJSON{Template: tmpl, MaxZoom: defaultMaxZoom}
	}
	return r, nil
}

// Network is the source as the pipeline takes it.
func (r *Remote) Network() *Network {
	if r == nil {
		return nil
	}
	return &Network{Identity: r.source, Get: r.Get, Zooms: r.zooms}
}

func (r *Remote) zooms() (lo, hi uint8) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.ready {
		return 0, defaultMaxZoom
	}
	return r.info.MinZoom, r.info.MaxZoom
}

// Attribution is the credit the source asks for, once its TileJSON is read.
func (r *Remote) Attribution() textsafe.Text {
	if r == nil {
		return textsafe.Text{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.info.Attribution
}

// own turns a fetcher's error into the library's own. A replacement
// fetcher's error is never passed on: it may hold the address.
func own(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return cancelled()
	}
	var mine *fault.Error
	if errors.As(err, &mine) {
		return mine
	}
	return sourceFailed()
}

// describe reads the TileJSON if it has not been read. Two jobs may both
// read it the first time; the first answer is kept.
func (r *Remote) describe(ctx context.Context) (TileJSON, error) {
	r.mu.Lock()
	ready, info, refusal := r.ready, r.info, r.refusal
	r.mu.Unlock()
	if ready || refusal != nil {
		return info, refusal
	}
	body, err := r.get(ctx, fetch.Request{URL: r.source, MaxBytes: maxTileJSONBytes})
	if err != nil {
		return TileJSON{}, own(ctx, err) // not kept: the network may come back
	}
	info, err = ParseTileJSON(body, r.source, r.allow)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ready || r.refusal != nil {
		return r.info, r.refusal
	}
	r.ready, r.info, r.refusal = err == nil, info, err
	return info, err
}

// Get fetches one tile's bytes as the source stores them.
func (r *Remote) Get(ctx context.Context, tile scene.TileID) ([]byte, error) {
	if r == nil || ctx == nil {
		return nil, refusedOptions(textsafe.Const("a tile was asked of no source, or with no context"))
	}
	if tile.Validate() != nil {
		return nil, refusedAddress(textsafe.Const("the tile asked for is not a tile"))
	}
	info, err := r.describe(ctx)
	if err != nil {
		return nil, err
	}
	body, err := r.get(ctx, fetch.Request{URL: info.Template.URL(tile), MaxBytes: maxTileBytes})
	if err != nil {
		return nil, own(ctx, err)
	}
	return body, nil
}
