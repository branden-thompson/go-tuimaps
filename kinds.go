package tuimaps

import (
	"errors"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
)

// Kind says what went wrong. The list is closed: a host may switch on it and
// know that a later version adds nothing to it without saying so (contract,
// section 7).
type Kind = fault.Kind

// The kinds of error, in the contract's order.
const (
	InvalidCoordinates = fault.InvalidCoordinates
	SizeMismatch       = fault.SizeMismatch
	UnsortedBreaks     = fault.UnsortedBreaks
	MalformedRamp      = fault.MalformedRamp
	MissingTable       = fault.MissingTable
	MalformedTable     = fault.MalformedTable
	UnknownPreset      = fault.UnknownPreset
	MalformedStyle     = fault.MalformedStyle
	InvalidID          = fault.InvalidID
	OverVertexCap      = fault.OverVertexCap
	OverImageCap       = fault.OverImageCap
	ImageRefused       = fault.ImageRefused
	RingTooShort       = fault.RingTooShort
	BadCurrency        = fault.BadCurrency
	UnsupportedSchema  = fault.UnsupportedSchema
	UnsupportedTile    = fault.UnsupportedTile
	OverLimit          = fault.OverLimit
	FetchRefused       = fault.FetchRefused
	FetchFailed        = fault.FetchFailed
	CacheRefused       = fault.CacheRefused
	NoSize             = fault.NoSize
	ReentrantCall      = fault.ReentrantCall
	Cancelled          = fault.Cancelled
	Closed             = fault.Closed
	Internal           = fault.Internal
)

// WarningKind says what a warning is about. This list is closed too.
type WarningKind = fault.WarningKind

// The kinds of warning, in the contract's order.
const (
	RampRuleBroken        = fault.RampRuleBroken
	UnmatchedImageColours = fault.UnmatchedImageColours
	StaleOverlay          = fault.StaleOverlay
	FutureValidTime       = fault.FutureValidTime
	ImplausibleUnit       = fault.ImplausibleUnit
	NearDuplicateID       = fault.NearDuplicateID
	UnknownToken          = fault.UnknownToken
	SetRefused            = fault.SetRefused
	NoWorkCalled          = fault.NoWorkCalled
	TileFailed            = fault.TileFailed
	CacheWriteFailed      = fault.CacheWriteFailed
	RenderFailed          = fault.RenderFailed
	CacheUnderNeed        = fault.CacheUnderNeed
)

// KindOf is the kind of an error the library raised, and false for an error
// that came from somewhere else. Every error a public call returns carries
// one: it is how a host tells a mistake of its own from a failure it can do
// nothing about.
func KindOf(err error) (Kind, bool) {
	var mine *fault.Error
	if !errors.As(err, &mine) {
		return 0, false
	}
	return mine.Kind(), true
}

// Kinds are every error kind there is, in the contract's order: what a host
// can switch on, and what a test can walk.
func Kinds() []Kind {
	out := make([]Kind, 0, int(Internal))
	for k := InvalidCoordinates; k <= Internal; k++ {
		out = append(out, k)
	}
	return out
}

// WarningKinds are every warning kind there is, in the contract's order.
func WarningKinds() []WarningKind {
	out := make([]WarningKind, 0, int(CacheUnderNeed))
	for k := RampRuleBroken; k <= CacheUnderNeed; k++ {
		out = append(out, k)
	}
	return out
}

// Rule is one of the rules a colour ramp is held to (FR-16, D-88).
type Rule = colour.Rule

// The rules. A host that themes a ramp can hold its own to the same ones.
const (
	Ordered    = colour.Ordered    // luminance moves one way, or one way each side of a midpoint
	Distinct   = colour.Distinct   // every class is a different colour at the depth in use
	Readable   = colour.Readable   // line work 3:1 against the ground; text 4.5:1 over a class
	VisionSafe = colour.VisionSafe // every pair, and every class against the ground, far enough apart
)

// Finding is one way a ramp breaks a rule: which rule, which classes, and
// by how much.
type Finding = colour.Finding

// RampCheck says how a ramp is used, which decides what is asked of it.
type RampCheck = colour.RampCheck

// CheckRamp holds a ramp of colours to the library's own rules and answers
// with what it found, empty for a ramp that passes (FR-16). It is the check
// the library runs on its own presets, exported so that a host can run it on
// a palette of its own, in its own tests, before anyone sees it.
func CheckRamp(ramp []RGB, how RampCheck) []Finding {
	return colour.Check(ramp, how)
}
