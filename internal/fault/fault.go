// Package fault holds the library's one error type and both closed lists of
// kinds: what a refused call or hand-in is, and what a warning is about.
// Every string in an error or a warning is textsafe.Text, so it arrived
// cleaned or was written as a constant in the library's source; no foreign
// error is ever wrapped, and no tile address is ever carried (FR-22b,
// FR-34). The public package re-exports what is here.
package fault

import (
	"context"

	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// Kind says what kind of refusal an Error is. The list is closed: adding a
// kind is a minor version (NFR-22).
type Kind uint8

// The kinds of error, in the contract's order (section 7).
const (
	InvalidCoordinates Kind = iota + 1
	SizeMismatch
	UnsortedBreaks
	MalformedRamp
	MissingTable
	MalformedTable
	UnknownPreset
	MalformedStyle
	InvalidID
	OverVertexCap
	OverImageCap
	ImageRefused
	RingTooShort
	BadCurrency
	UnsupportedSchema
	UnsupportedTile
	OverLimit
	FetchRefused
	FetchFailed
	CacheRefused
	NoSize
	ReentrantCall
	Cancelled
	Closed
	Internal

	lastKind = Internal
)

// String returns the kind's name as the contract writes it.
func (k Kind) String() string {
	if k == 0 || k > lastKind {
		return "unknown"
	}
	return [...]string{
		"invalid-coordinates", "size-mismatch", "unsorted-breaks", "malformed-ramp", "missing-table",
		"malformed-table", "unknown-preset", "malformed-style", "invalid-id", "over-vertex-cap", "over-image-cap",
		"image-refused", "ring-too-short", "bad-currency", "unsupported-schema", "unsupported-tile",
		"over-limit", "fetch-refused", "fetch-failed", "cache-refused", "no-size", "reentrant-call",
		"cancelled", "closed", "internal",
	}[k-1]
}

// WarningKind says what a warning is about. The list is closed.
type WarningKind uint8

// The kinds of warning, in the contract's order (section 7).
const (
	RampRuleBroken WarningKind = iota + 1
	UnmatchedImageColours
	StaleOverlay
	FutureValidTime
	ImplausibleUnit
	NearDuplicateID
	UnknownToken
	SetRefused
	BorrowChanged
	NoWorkCalled
	TileFailed
	CacheWriteFailed
	RenderFailed
	CacheUnderNeed

	lastWarningKind = CacheUnderNeed
)

// String returns the warning kind's name as the contract writes it.
func (k WarningKind) String() string {
	if k == 0 || k > lastWarningKind {
		return "unknown"
	}
	return [...]string{
		"ramp-rule-broken", "unmatched-image-colours", "stale-overlay", "future-valid-time",
		"implausible-unit", "near-duplicate-id", "unknown-token", "set-refused", "borrow-changed", "no-work-called",
		"tile-failed", "cache-write-failed", "render-failed", "cache-under-need",
	}[k-1]
}

// Error is the library's one error type: a kind from the closed list, and
// three cleaned sentences - what happened, why, and what to do next.
type Error struct {
	kind  Kind
	meant Kind // set only when an incomplete error was turned into Internal: the kind that was asked for
	what  textsafe.Text
	why   textsafe.Text
	todo  textsafe.Text
}

// Make makes an Error. Each sentence is Text, so it was cleaned or is a
// constant of the library's own. A kind outside the list is a fault of the
// library's, and becomes Internal.
func Make(kind Kind, what, why, todo textsafe.Text) *Error {
	if kind == 0 || kind > lastKind {
		kind = Internal
	}
	if what.String() == "" || why.String() == "" || todo.String() == "" {
		// Every error says what happened, why, and what to do next. One that
		// cannot is a defect here, and says so rather than printing a gap.
		return &Error{
			kind:  Internal,
			meant: kind,
			what:  textsafe.Const("the library raised an error without saying what happened, why, or what to do"),
			why:   textsafe.Const("that is a defect in the library"),
			todo:  textsafe.Const("report it, with the kind named here"),
		}
	}
	return &Error{kind: kind, what: what, why: why, todo: todo}
}

// Kind returns the error's kind.
func (e *Error) Kind() Kind {
	if e == nil {
		return Internal
	}
	return e.kind
}

// Error gives the kind and the three sentences.
func (e *Error) Error() string {
	if e == nil {
		return "tuimaps: internal: a nil error was reported"
	}
	name := e.kind.String()
	if e.meant != 0 {
		name += " (raised as " + e.meant.String() + ")"
	}
	return "tuimaps: " + name + ": " + e.what.String() + ": " + e.why.String() + ". " + e.todo.String() + "."
}

// Is lets a host's usual check work: an error of the Cancelled kind answers
// errors.Is for the context package's two errors. It is the one stated
// exemption to "no foreign error": nothing of theirs is wrapped or quoted.
func (e *Error) Is(target error) bool {
	if e == nil || e.kind != Cancelled {
		return false
	}
	return target == context.Canceled || target == context.DeadlineExceeded
}

// Warning is one thing that was accepted but is off, or that happened
// during Work: its kind, the overlay or tile it concerns, and how often.
type Warning struct {
	Kind    WarningKind
	Subject textsafe.Text
	Count   int
}
