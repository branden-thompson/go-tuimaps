---
title: "v0.2.0 PLAN — Approach 4: the description's additions"
date: 2026-09-23
phase: PLAN
sev: SEV-0
authority: HUM LEAD
status: "RULED — D-57: B, a structured Report beside Describe (the HUM LEAD's reason: no consumers of v0.1.0 yet). Signatures and shape only."
---

# Approach 4 — the description's additions

**The question.** Where do the description's new answers live in the API? They are: each alert
answered on its own, with name and severity (L-13.6); "nearby" (L-13.6); a list of the alerts shown,
with no place needed (L-13.5); severity and valid time as data on each alert (L-13.9); and motion,
relative to a named place or to the view (L-1.12). And all of it must stay additive to v0.1.0's
`Describe` (NFR-1).

**Today.** `Describe(places) ([]Description, error)` returns one `Description` per place, each with one
flat `Answer` per overlay (`describe.go:19-28`, `internal/describe/answer.go:52-79`). An area answer
merges every feature of the overlay and carries no alert name. A feature knows its colour `Role` and
`Label`, and nothing else about the alert (`internal/overlay/store.go:50-57`). With no place, `Describe`
returns nothing. The answer is data, never a sentence (v0.1.0 D-52).

## Shared by every option — what an alert carries

```go
type Severity uint8 // SeverityUnknown, Minor, Moderate, Severe, Extreme

type Feature struct {
	// … v0.1.0's fields, unchanged …
	Severity Severity  // new; zero = worked out from Role, as it is drawn today (L-13.9)
	Valid    time.Time // new; zero = the overlay's Valid
	Expires  time.Time // new; zero = the overlay's Valid + Keeps
}

type Where uint8 // v0.1.0: Outside, Inside; new: Nearby (L-13.6)

// SetNearby sets how close counts as nearby (L-13.6): a host setting, per
// the standing rule that the listener chooses; default 10 km until PLAN's
// number is set.
func (m *Map) SetNearby(km float64) error

// Motion is two timed positions and a relation, never a forecast (L-1.12).
type Motion struct {
	Threshold          int       // the class "heavier" means, from the table (D-42)
	From, To           Sighting  // the oldest usable and the newest non-gap frame
	Relation           string    // "closer", "farther", "held" — words said aloud, as Form is
	Span               time.Duration
}
type Sighting struct {
	Valid   time.Time
	Km      float64 // from the named place, or from the view's centre
	Bearing float64
	Compass string
	InView  bool    // relative to the view: within it, or beyond an edge
}
```

## The choice: where the new answers go

### A — Extend the answers v0.1.0 already gives

```go
type Answer struct {
	// … v0.1.0's flat fields, unchanged, still one per overlay …
	Alerts []AlertAnswer // new, for an area overlay: each alert, on its own
	Motion *Motion       // new, for a looped image
}
type AlertAnswer struct {
	Label    string
	Severity Severity
	Relation Where // inside, outside, nearby
	Km       float64
	Bearing  float64
	Valid, Expires time.Time
	Stale    bool
}

// Alerts is the list of alerts shown, needing no place (L-13.5).
func (m *Map) Alerts() []AlertAnswer // Relation zero, Km zero: no place asked

// Describe with no places, where none are registered, now returns one
// Description for the view itself: Place is empty, and it carries the motion
// relative to the view (L-1.12, D-42). v0.1.0 returned nothing there.
```

A v0.1.0 host that reads the flat fields sees exactly what it saw before. A v0.2.0 host reads
`Alerts` for the per-alert detail, and `Motion` for movement.

- **For:** one surface. `Describe` stays the one call a host makes. Old fields keep their meaning, and
  new detail is additive.
- **Against:** an area `Answer` then says the same thing twice, merged in its flat fields and per
  alert in `Alerts`, and a host must know which to read. And `Describe` with no places changes from
  "nothing" to "the view", which a v0.1.0 host that treated nothing as "no description" would notice.

### B — A new, structured report beside `Describe`

```go
type Report struct {
	Alerts []AlertShown   // the list, no place needed
	Places []PlaceReport  // each named place: each alert's relation, each image's answer
	Motion []MotionReport // per looped image: relative to each place, and to the view
}
func (m *Map) Report(places []Place) (Report, error)
```

`Describe` stays exactly v0.1.0. `Report` is the new, complete answer, organised the way a listener
asks: what is out there, where am I relative to it, and what is moving.

- **For:** a clean model with no duplicated fields. `Describe`'s behaviour doesn't change at all,
  the new shape is organised for how a screen-reader host asks, and it could supersede `Describe`
  later.
- **Against:** two description calls, one of which a host must choose. Two surfaces to keep
  consistent and tested, both kept for good. And watchpost's M1b, built on `Describe`, would move to
  `Report` to get the new answers.

## Recommendation

**A.** It stays on the one surface hosts already use. A v0.1.0 host sees no difference, apart from
a view description where there used to be nothing, and that is the point of D-42. The duplication in
an area answer is managed by documenting the flat fields as "the merged summary" and `Alerts` as
"each one".

**The strongest argument against A.** B is the better long-term model. A grows a flat struct that
already carries some fifteen optional fields, and a later clean-up would have to break it, which
NFR-1 makes a ruling of its own. B pays two surfaces now to avoid that.

## Cross-reference with watchpost 0.18.0

| Watchpost | Here |
|---|---|
| FR-7.4: the text description ships in phase 1 | `Describe`; `Alerts` for the list |
| M1b: where is it, in words | Per-alert `Relation` and distance, from `Answer.Alerts` |
| D-29, D-43: never "you"; inside / outside / nearby | `Where.Nearby` via `SetNearby` (the host's Setting); watchpost words it |
| D-42: motion in a view with no names | The view `Description` when no place is asked |
