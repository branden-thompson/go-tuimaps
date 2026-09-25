package fetch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// refusedSource is the error for a source, a token or a request the policy
// does not permit. It names nothing, because nothing has been validated.
func refusedSource() error {
	return fault.Make(fault.FetchRefused,
		textsafe.Const("the tile source was refused"),
		textsafe.Const("it is not a secure address with a host, it carries a user name, its token holds characters a header cannot, or the request gave no size limit"),
		textsafe.Const("name an https source, or allow plain http explicitly; give every request a maximum size"))
}

// problem makes the fetcher's own error of the given kind. It names the
// source's scheme and host - all of an address an error may name - and keeps
// nothing of the transport's error, which holds the whole address.
func (f *Fetcher) problem(kind fault.Kind, why textsafe.Text) error {
	about := textsafe.Join(textsafe.Const("a request to "), textsafe.Quote(f.scheme+"://"+f.host))
	switch kind {
	case fault.FetchRefused:
		return fault.Make(kind, textsafe.Join(about, textsafe.Const(" was refused")),
			textsafe.Const("it would have left the named source: another scheme or host, a fourth redirect, a redirect from secure to plain, or a connection landing in private address space"),
			textsafe.Const("check the source; a host a redirect may go to can be allowed explicitly"))
	case fault.Cancelled:
		return fault.Make(kind, textsafe.Join(about, textsafe.Const(" was abandoned")),
			textsafe.Const("the work it was part of was cancelled or ran out of time"),
			textsafe.Const("nothing; it is asked for again if it is still wanted"))
	}
	return fault.Make(fault.FetchFailed, textsafe.Join(about, textsafe.Const(" failed")), why,
		textsafe.Const("the tile will be tried again later; check the network and the source"))
}

func tooLarge() error {
	return fault.Make(fault.OverLimit,
		textsafe.Const("a reply was refused"),
		textsafe.Const("its body is larger than the request allowed"),
		textsafe.Const("check the source; the limits protect the host's memory"))
}

// Fetch fetches one request from the source. The transport's own error is
// never returned or wrapped: it holds the address.
func (f *Fetcher) Fetch(ctx context.Context, r Request) ([]byte, error) {
	if f == nil || ctx == nil {
		return nil, refusedSource()
	}
	resp, err := f.send(ctx, r)
	if err != nil {
		return nil, err
	}
	data, err := f.readReply(ctx, resp, r)
	_ = resp.Body.Close() // the body was read through its limit or is being abandoned; a close error changes nothing
	if err == nil && ctx.Err() != nil {
		// A late answer is never used (L-7.3): a transport that ignored its
		// context has held this call past its end, and what it brought back
		// is thrown away.
		return nil, f.problem(fault.Cancelled, textsafe.Text{})
	}
	return data, err
}

// send checks a request against the source's policy and makes it. The caller
// closes the reply's body.
func (f *Fetcher) send(ctx context.Context, r Request) (*http.Response, error) {
	if r.MaxBytes <= 0 || r.RangeStart < 0 || r.RangeLen < 0 {
		return nil, refusedSource()
	}
	u, err := url.Parse(r.URL)
	if err != nil {
		return nil, f.problem(fault.FetchRefused, textsafe.Text{})
	}
	if !f.allowed(u) {
		return nil, f.problem(fault.FetchRefused, textsafe.Text{})
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, f.problem(fault.FetchRefused, textsafe.Text{})
	}
	req.Header.Set("User-Agent", f.agent)
	if r.RangeLen > 0 {
		req.Header.Set("Range", "bytes="+strconv.FormatInt(r.RangeStart, 10)+"-"+strconv.FormatInt(r.RangeStart+r.RangeLen-1, 10))
	}
	proxy, err := f.transport.Proxy(req)
	if err != nil {
		return nil, f.problem(fault.FetchFailed, textsafe.Const("the proxy settings in the environment could not be read"))
	}
	if proxy != nil {
		req = req.WithContext(context.WithValue(ctx, viaProxy{}, true))
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, f.transportError(ctx, err)
	}
	return resp, nil
}

// transportError turns the client's error into one of the library's own,
// keeping nothing of it.
func (f *Fetcher) transportError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return f.problem(fault.Cancelled, textsafe.Text{})
	}
	if errors.Is(err, errPolicy) {
		return f.problem(fault.FetchRefused, textsafe.Text{})
	}
	return f.problem(fault.FetchFailed, textsafe.Const("the connection could not be made, or the source did not answer in time"))
}

// readReply checks the status, and for a range the range, then reads the body
// through the limit. A range reply that is not a 206 with exactly the range
// asked for is closed unread.
func (f *Fetcher) readReply(ctx context.Context, resp *http.Response, r Request) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, f.problem(fault.FetchFailed, textsafe.Const("the source sent no reply to read"))
	}
	want := http.StatusOK
	if r.RangeLen > 0 {
		want = http.StatusPartialContent
	}
	if resp.StatusCode != want {
		return nil, f.problem(fault.FetchFailed, textsafe.Join(textsafe.Const("the source answered with status "), textsafe.Clean(strconv.Itoa(resp.StatusCode))))
	}
	if r.RangeLen > 0 {
		exact := "bytes " + strconv.FormatInt(r.RangeStart, 10) + "-" + strconv.FormatInt(r.RangeStart+r.RangeLen-1, 10) + "/"
		got := resp.Header.Get("Content-Range")
		if len(got) <= len(exact) || got[:len(exact)] != exact {
			return nil, f.problem(fault.FetchFailed, textsafe.Const("the source did not answer with exactly the range that was asked for"))
		}
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, r.MaxBytes+1))
	if err != nil {
		return nil, f.transportError(ctx, err)
	}
	if int64(len(data)) > r.MaxBytes {
		return nil, tooLarge()
	}
	if r.RangeLen > 0 && int64(len(data)) != r.RangeLen {
		return nil, f.problem(fault.FetchFailed, textsafe.Const("the source sent a different number of bytes than the range that was asked for"))
	}
	return data, nil
}
