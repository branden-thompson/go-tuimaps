package fetch

import (
	"context"
	"strconv"
	"strings"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// maxEntityTag bounds the entity tag kept from a reply.
	maxEntityTag = 256
	// maxProbe bounds the first range Describe asks for.
	maxProbe = 16384
)

// Info is what a source says about one file: how long it is, and the tag
// that changes when the file does. Both are the source's word, not a proof.
type Info struct {
	Length    int64
	EntityTag textsafe.Text // outside text, cleaned; empty if the source sent none
}

// Describe asks for a file's first bytes and reads its length and entity tag
// from the reply. It is held to the rule every range is: a 206 with exactly
// the range asked for, or the reply is closed unread. The caller chooses how
// many bytes, up to maxProbe, and the file must be at least that long: one
// public host was seen to ignore a range of a single byte and answer with the
// whole file, while honouring a range of 127.
func (f *Fetcher) Describe(ctx context.Context, rawURL string, probe int64) (Info, error) {
	if f == nil || ctx == nil {
		return Info{}, refusedSource()
	}
	if rawURL == "" || probe <= 0 || probe > maxProbe {
		return Info{}, refusedSource()
	}
	first := Request{URL: rawURL, RangeStart: 0, RangeLen: probe, MaxBytes: probe}
	resp, err := f.send(ctx, first)
	if err != nil {
		return Info{}, err
	}
	_, err = f.readReply(ctx, resp, first)
	_ = resp.Body.Close() // the probe was read or the reply is being abandoned; a close error changes nothing
	if err != nil {
		return Info{}, err
	}
	_, total, _ := strings.Cut(resp.Header.Get("Content-Range"), "/")
	length, err := strconv.ParseInt(total, 10, 64)
	if err != nil {
		return Info{}, f.problem(fault.FetchFailed, textsafe.Const("the source did not say how long the file is"))
	}
	if length <= 0 {
		return Info{}, f.problem(fault.FetchFailed, textsafe.Const("the source did not say how long the file is"))
	}
	tag := resp.Header.Get("ETag")
	if len(tag) > maxEntityTag {
		tag = ""
	}
	return Info{Length: length, EntityTag: textsafe.Clean(tag)}, nil
}
