package overlay

import (
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// maxCurrency is the longest an overlay may say its data stays current.
	maxCurrency = 7 * 24 * time.Hour
	// aheadAllowed is how far ahead of the host's clock a valid time may be
	// before it is taken for a bad timestamp: clocks differ by minutes, not more.
	aheadAllowed = 5 * time.Minute
)

// Freshness is whether an overlay's data is still current (FR-32).
type Freshness uint8

// The states of freshness.
const (
	Current   Freshness = iota + 1
	Stale               // older than the host said it stays current
	Uncertain           // valid more than five minutes ahead of the clock: a bad timestamp
)

// String names the state.
func (f Freshness) String() string {
	switch f {
	case Current:
		return "current"
	case Stale:
		return "stale"
	case Uncertain:
		return "uncertain"
	}
	return "unknown"
}

// DrawnStale reports whether the frame marks the overlay stale. Uncertain
// data is, so that a bad timestamp cannot keep data looking fresh for ever.
func (f Freshness) DrawnStale() bool {
	return f != Current
}

// CheckCurrency holds a valid time and a currency to FR-32: every overlay
// carries a valid time, and stays current for more than zero and at most
// seven days.
func CheckCurrency(valid time.Time, keeps time.Duration) error {
	if valid.IsZero() {
		return fault.Make(fault.BadCurrency, textsafe.Const("the overlay was refused"),
			textsafe.Const("it does not say when its data was valid"),
			textsafe.Const("give the time the data was valid: the map shows when data is old"))
	}
	if keeps <= 0 || keeps > maxCurrency {
		return fault.Make(fault.BadCurrency, textsafe.Const("the overlay was refused"),
			textsafe.Const("how long its data stays current must be more than zero and at most seven days"),
			textsafe.Const("give the period after which this data should be shown as old"))
	}
	return nil
}

// FreshnessAt is an overlay's freshness on the host's clock.
func FreshnessAt(valid time.Time, keeps time.Duration, now time.Time) Freshness {
	if valid.After(now.Add(aheadAllowed)) {
		return Uncertain
	}
	if now.Sub(valid) > keeps {
		return Stale
	}
	return Current
}

// NextChange is when a current overlay goes stale, for the moment the host is
// told to call again by. An overlay that is not current has none.
func NextChange(valid time.Time, keeps time.Duration, now time.Time) (time.Time, bool) {
	if FreshnessAt(valid, keeps, now) != Current {
		return time.Time{}, false
	}
	return valid.Add(keeps).Add(time.Nanosecond), true
}
