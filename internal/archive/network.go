package archive

import (
	"context"

	"github.com/branden-thompson/go-tuimaps/internal/fetch"
)

// OverNetwork returns a reader that gets each range of the archive at rawURL
// as a range request through get. The fetcher refuses any reply but a 206
// with exactly the range asked for, so a source that answers with the whole
// file - 86 GB of it, for the planet - is closed unread.
func OverNetwork(get fetch.Func, rawURL string) (RangeReader, error) {
	if get == nil {
		return nil, bad()
	}
	if rawURL == "" {
		return nil, bad()
	}
	return func(ctx context.Context, offset, length int64) ([]byte, error) {
		return get(ctx, fetch.Request{URL: rawURL, RangeStart: offset, RangeLen: length, MaxBytes: length})
	}, nil
}
