package tiles

// State is where a tile is in its life (L3 States, machine 1).
type State uint8

// The states of a tile.
const (
	Wanted      State = iota + 1 // a view wants it and nothing has been done yet
	Queued                       // a job for it is pending
	Loading                      // a Work call is fetching or decoding it
	OnHand                       // decoded and in the memory cache
	Unavailable                  // no source is named, and the embedded tiles do not hold it
	Waiting                      // a named source failed; it may be tried again at its not-before time
	Evicted                      // it was on hand, nothing needed it, and the cache was over its cap
	Dropped                      // no longer wanted
)

// String names the state.
func (s State) String() string {
	if s < Wanted || s > Dropped {
		return "unknown"
	}
	switch s {
	case Wanted:
		return "wanted"
	case Queued:
		return "queued"
	case Loading:
		return "loading"
	case OnHand:
		return "on-hand"
	case Unavailable:
		return "unavailable"
	case Waiting:
		return "waiting"
	case Evicted:
		return "evicted"
	case Dropped:
		return "dropped"
	}
	return "unknown"
}

// Event is something that happens to a tile.
type Event uint8

// The events of a tile's life.
const (
	Enqueue        Event = iota + 1 // it joined the pending work
	LeftView                        // no view wants it any more
	CapDropped                      // the queue's cap dropped it
	PickedUp                        // a Work call picked it up
	Cancelled                       // its context ended: it left the view while loading
	PassedGate                      // its bytes passed the untrusted-input gate
	NoSource                        // no source is named and the embedded tiles do not hold it
	SourceFailed                    // a named source refused, failed, or had no such tile
	RetryDue                        // its not-before time has passed and it is still wanted
	SourcesChanged                  // a source was named, or embedded tiles were passed
	Evict                           // nothing needs it and the cache is over its cap
	WantedAgain                     // a view wants it again
)

// String names the event.
func (e Event) String() string {
	if e < Enqueue || e > WantedAgain {
		return "unknown"
	}
	switch e {
	case Enqueue:
		return "enqueue"
	case LeftView:
		return "left-view"
	case CapDropped:
		return "cap-dropped"
	case PickedUp:
		return "picked-up"
	case Cancelled:
		return "cancelled"
	case PassedGate:
		return "passed-gate"
	case NoSource:
		return "no-source"
	case SourceFailed:
		return "source-failed"
	case RetryDue:
		return "retry-due"
	case SourcesChanged:
		return "sources-changed"
	case Evict:
		return "evict"
	case WantedAgain:
		return "wanted-again"
	}
	return "unknown"
}

// Next is the state machine: the state an event leads to, or the same state
// and false if the event does not apply there.
func Next(from State, ev Event) (State, bool) {
	if from < Wanted || from > Dropped {
		return from, false
	}
	if ev < Enqueue || ev > WantedAgain {
		return from, false
	}
	switch from {
	case Queued:
		return nextQueued(ev)
	case Loading:
		return nextLoading(ev)
	case Waiting:
		return nextWaiting(ev)
	}
	return nextSingle(from, ev)
}

// nextSingle is the states that one event alone can leave.
func nextSingle(from State, ev Event) (State, bool) {
	switch {
	case from == Wanted && ev == Enqueue:
		return Queued, true
	case from == Unavailable && ev == SourcesChanged:
		return Wanted, true
	case from == OnHand && ev == Evict:
		return Evicted, true
	case from == Evicted && ev == WantedAgain:
		return Wanted, true
	}
	return from, false
}

func nextWaiting(ev Event) (State, bool) {
	if ev == RetryDue {
		return Queued, true
	}
	if ev == LeftView {
		return Dropped, true
	}
	return Waiting, false
}

func nextQueued(ev Event) (State, bool) {
	switch ev {
	case LeftView, CapDropped:
		return Dropped, true
	case PickedUp:
		return Loading, true
	}
	return Queued, false
}

func nextLoading(ev Event) (State, bool) {
	switch ev {
	case Cancelled:
		return Dropped, true
	case PassedGate:
		return OnHand, true
	case NoSource:
		return Unavailable, true
	case SourceFailed:
		return Waiting, true
	}
	return Loading, false
}
