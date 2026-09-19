package textsafe

import (
	"errors"
	"fmt"
	"strings"

	"github.com/clipperhouse/uax29/v2/graphemes"
)

const (
	// QuoteClusters is how much untrusted text an error or a warning may
	// quote: 64 grapheme clusters (FR-34).
	QuoteClusters = 64
	// MaxIDBytes is the longest id a host may give an overlay or a place.
	MaxIDBytes = 256
)

// Width is the number of cells t occupies: the sum of its clusters, each
// one or two cells by the pinned width table (NFR-8).
func Width(t Text) int {
	cells := 0
	clusters := graphemes.FromString(t.s)
	for range len(t.s) { // a cluster is at least one byte
		if !clusters.Next() {
			break
		}
		cells += clusterWidth(clusters.Value())
	}
	return cells
}

// Fit returns the longest run of whole clusters from the start of t that
// occupies no more than cells. A cluster that does not fit is dropped
// whole, never split: not half a wide character, not a letter without its
// accent.
func Fit(t Text, cells int) Text {
	if cells <= 0 {
		return Text{}
	}
	used, end := 0, 0
	clusters := graphemes.FromString(t.s)
	for range len(t.s) {
		if !clusters.Next() {
			break
		}
		w := clusterWidth(clusters.Value())
		if used+w > cells {
			break
		}
		used, end = used+w, clusters.End()
	}
	return Text{s: t.s[:end]}
}

// Quote cleans untrusted text for quoting inside an error or a warning and
// cuts it to QuoteClusters clusters, saying so when it cuts.
func Quote(s string) Text {
	cleaned := Clean(s)
	count, end := 0, 0
	clusters := graphemes.FromString(cleaned.s)
	for range len(cleaned.s) {
		if !clusters.Next() {
			return cleaned // the whole text fits
		}
		if count == QuoteClusters {
			return Text{s: cleaned.s[:end] + "..."}
		}
		count, end = count+1, clusters.End()
	}
	return cleaned
}

// ID validates an id a host handed in and returns it unchanged. Ids are
// never cleaned: cleaning could make two different ids the same. An id that
// would need cleaning is refused instead (FR-34, NFR-20).
func ID(id string) (Text, error) {
	if id == "" {
		return Text{}, errors.New("textsafe: the id is empty")
	}
	if len(id) > MaxIDBytes {
		return Text{}, fmt.Errorf("textsafe: the id is %d bytes; at most %d are allowed", len(id), MaxIDBytes)
	}
	if cleaned := Clean(id); cleaned.s != id {
		return Text{}, fmt.Errorf("textsafe: the id %q holds characters that cannot be shown safely; ids are never altered, so it is refused", Quote(strings.ToValidUTF8(id, "?")))
	}
	return Text{s: id}, nil
}
