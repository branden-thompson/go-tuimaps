package fetch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// TestDescribe: a file's length and entity tag come from a small first range
// request, held to the same rule as every range: a 206 with exactly the
// range asked for. The tag is outside text and is cleaned.
func TestDescribe(t *testing.T) {
	var mode string
	srv, opts := secure(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Range") != "bytes=0-126" {
			t.Errorf("%s with Range %q; want a range request for the first 127 bytes", r.Method, r.Header.Get("Range"))
		}
		total := "86516374189"
		switch mode {
		case "whole":
			fmt.Fprint(w, "the whole file")
			return
		case "no total":
			total = "*"
		case "not a number":
			total = "12x"
		case "zero":
			total = "0"
		case "hostile tag":
			w.Header().Set("ETag", "\"abc\xe2\x80\xaedcba\"") // a direction override; a control byte never gets past the client
		}
		if mode != "hostile tag" {
			w.Header().Set("ETag", `"76dd-9168"`)
		}
		w.Header().Set("Content-Range", "bytes 0-126/"+total)
		w.WriteHeader(http.StatusPartialContent)
		fmt.Fprint(w, strings.Repeat("P", 127))
	}))
	f, err := ForSource(srv.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	mode = "honest"
	info, err := f.Describe(context.Background(), srv.URL+"/planet", 127)
	if err != nil || info.Length != 86516374189 || info.EntityTag.String() != `"76dd-9168"` {
		t.Errorf("an honest reply: %+v, %v", info, err)
	}
	mode = "hostile tag"
	info, err = f.Describe(context.Background(), srv.URL+"/planet", 127)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.EntityTag.String(); got != `"abcdcba"` {
		t.Errorf("the tag was not cleaned: %q", got)
	}
	for _, mode = range []string{"whole", "no total", "not a number", "zero"} {
		if _, err := f.Describe(context.Background(), srv.URL+"/planet", 127); !isKind(err, fault.FetchFailed) {
			t.Errorf("%s: %v; want the fetch-failed kind", mode, err)
		}
	}
	for _, probe := range []int64{0, -1, 1 << 20} {
		if _, err := f.Describe(context.Background(), srv.URL+"/planet", probe); !isKind(err, fault.FetchRefused) {
			t.Errorf("a probe of %d bytes: %v; want the fetch-refused kind", probe, err)
		}
	}
	if _, err := f.Describe(context.Background(), "https://elsewhere.invalid/planet", 127); !isKind(err, fault.FetchRefused) {
		t.Errorf("another host: %v; want the fetch-refused kind", err)
	}
}
