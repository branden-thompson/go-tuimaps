package tuimaps

import (
	"context"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// Work does one unit of the map's slow work - a tile fetched or decoded, an
// overlay prepared - on the goroutine that calls it, and returns when that
// unit is done or ctx ends. It reports whether there was anything to do.
//
// **The library starts no goroutine of its own** (D-73): a host that wants
// work done in the background calls this from goroutines of its own, as many
// as it likes. It is safe beside Render and beside every other call.
func (m *Map) Work(ctx context.Context) (did bool, err error) {
	defer guard("Work", &err)
	m.plant("Work")

	if m == nil {
		return false, closed()
	}
	if ctx == nil {
		return false, fault.Make(fault.Internal, textsafe.Const("the map could not work"),
			textsafe.Const("it was given no context"), textsafe.Const("pass a context; context.Background() if there is no other"))
	}
	m.inside.Add(1)
	defer m.inside.Add(-1)
	m.mu.Lock()
	shut := m.shut
	m.mu.Unlock()
	if shut {
		return false, closed()
	}
	before := m.landed()
	did, err = m.member.RunOne(ctx)
	m.noteLanded(before)
	return did, err
}

// landed counts what work has landed: tiles into the cache and prepared
// overlays into the store. Each has its own lock; the map's is not needed.
func (m *Map) landed() uint64 {
	return m.pipe.Landed() + m.store.Landed()
}

// noteLanded raises Changed when work landed something since before: a host
// that renders when Changed moves then draws it, with no Render needed to
// find out (D-66). Anything landed counts, since the view asked for it.
func (m *Map) noteLanded(before uint64) {
	if m.landed() == before {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.shut {
		m.changed++
	}
}

// OnPending sets a hook called when this map's pending work goes from none
// to some. It is how a host is told rather than having to poll, and it is
// called on whichever goroutine made the work pending, so it should do no
// more than wake a pump.
func (m *Map) OnPending(hook func()) (err error) {
	defer guard("OnPending", &err)
	m.plant("OnPending")

	if m == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	return m.member.WhenPending(hook)
}

// InFlight is how many units of work are inside a Work call somewhere else.
// Settle never waits for them (D-86).
func (m *Map) InFlight() int {
	defer m.guardQuiet("InFlight")
	m.plant("InFlight")

	if m == nil {
		return 0
	}
	return m.member.Flying()
}
