// Package work is the queue of slow jobs a host drains. The library never
// runs anything by itself (D-73): a job runs inside a host's Work call, on
// the host's goroutine, and nothing here starts a goroutine or a timer.
//
// There is no limiter (D-84). Work never waits on another Work: how many run
// at once is the host's pump's width, and so is the memory that costs - each
// further concurrent Work can add about 1 MB while a tile decodes, more at
// the input limits. The memory line is stated for a pump two wide.
//
// A queue may be shared by several maps. Each is a Member, with its own wake
// hook, its own change counter and its own say in what is still wanted. The
// queue knows jobs only as scene.Job: it imports none of the packages that
// supply them.
package work

import (
	"context"
	"sync"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

const (
	// DefaultCap is how many jobs may wait (constants, section 2).
	DefaultCap = 256
	// noWorkRenders is how many renders in a row may find work pending, with
	// no Work called, before the host is warned (constants, section 5).
	noWorkRenders = 20
	// maxSettle bounds one Settle.
	maxSettle = 1 << 20
)

// entry is one job and who wants it.
type entry struct {
	job     scene.Job
	wanters map[*Member]bool
	view    uint64 // the newest view, of any wanter, that asked for it
	seq     uint64 // the order it was first asked for
}

// flight is a job inside a Work call.
type flight struct {
	entry   *entry
	cancel  context.CancelFunc
	dropped bool // nobody wants it any longer; its outcome is ignored
}

// Queue holds the waiting jobs of every map that joined it.
type Queue struct {
	mu      sync.Mutex
	limit   int
	waiting []*entry
	flying  map[string]*flight
	seq     uint64
}

// NewQueue makes a queue that lets limit jobs wait; fewer than one means
// DefaultCap.
func NewQueue(limit int) *Queue {
	if limit < 1 {
		limit = DefaultCap
	}
	return &Queue{limit: limit, flying: map[string]*flight{}}
}

// Waiting is how many jobs wait, for every member together.
func (q *Queue) Waiting() int {
	if q == nil {
		return 0
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.waiting)
}

// deferred is a failed job waiting for the time it may be tried again.
type deferred struct {
	job scene.Job
	at  time.Time
}

// Member is one map's share of a queue.
type Member struct {
	q         *Queue
	closed    bool
	view      uint64
	hook      func()
	inHook    bool // a hook fired by an owner call is running: owner calls are refused
	changed   uint64
	later     []deferred
	deadlines map[string]time.Time
	idle      int  // renders in a row that found work pending with no Work called
	warned    bool // the no-work warning has been given since the last Work
}

// Join adds a map to the queue.
func (q *Queue) Join() (*Member, error) {
	if q == nil {
		return nil, internal()
	}
	return &Member{q: q, deadlines: map[string]time.Time{}}, nil
}

// internal is the error for a defect inside the library. It repeats nothing
// of what caused it, which may hold outside data.
func internal() error {
	return fault.New(fault.Internal,
		textsafe.Const("a piece of background work could not be queued or failed unexpectedly"),
		textsafe.Const("the library handed itself a job it cannot run, or a job panicked"),
		textsafe.Const("report this as a defect in the library; the map stays usable"))
}

// problem is the error for a call the queue will not serve.
func problem(kind fault.Kind) error {
	switch kind {
	case fault.Closed:
		return fault.New(kind, textsafe.Const("the map is closed"), textsafe.Const("Close was called"), textsafe.Const("make a new map"))
	case fault.ReentrantCall:
		return fault.New(kind, textsafe.Const("the call was refused"), textsafe.Const("it was made from inside the pending-work hook, which must not call the map"), textsafe.Const("have the hook signal the pump and do nothing else"))
	}
	return fault.New(fault.Cancelled, textsafe.Const("the work was abandoned"), textsafe.Const("its context was cancelled or ran out of time"), textsafe.Const("nothing; it is asked for again if it is still wanted"))
}

// owner is the check every owner call makes first. Owner calls are one at a
// time, so one arriving while an owner-fired hook runs can only be re-entry.
func (m *Member) owner() error {
	if m == nil || m.q == nil {
		return internal()
	}
	if m.closed {
		return problem(fault.Closed)
	}
	if m.inHook {
		return problem(fault.ReentrantCall)
	}
	return nil
}

// pendingLocked counts the waiting jobs m wants. A job inside a Work call
// and a failed job waiting for its retry time are not pending.
func (m *Member) pendingLocked() int {
	n := 0
	for _, e := range m.q.waiting {
		if e.wanters[m] {
			n++
		}
	}
	return n
}

// Pending is how many of this map's jobs wait to be picked up. It may be
// called from any goroutine.
func (m *Member) Pending() int {
	if m == nil || m.q == nil {
		return 0
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	return m.pendingLocked()
}

// InFlight is how many of this map's jobs are inside a Work call.
func (m *Member) InFlight() int {
	if m == nil || m.q == nil {
		return 0
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	n := 0
	for _, f := range m.q.flying {
		if f.entry.wanters[m] {
			n++
		}
	}
	return n
}

// Changed moves whenever a job this map wanted finishes or fails: either
// changes what a redraw would show. It may be called from any goroutine.
func (m *Member) Changed() uint64 {
	if m == nil || m.q == nil {
		return 0
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	return m.changed
}

// OnPending sets the hook called when this map's pending work goes from
// none to some. It is called with no lock held and must not call the map.
func (m *Member) OnPending(hook func()) error {
	err := m.owner()
	if err != nil {
		return err
	}
	m.q.mu.Lock()
	m.hook = hook
	m.q.mu.Unlock()
	return nil
}

// NewView says the map's view has moved: jobs asked for from now on run
// before those asked for earlier.
func (m *Member) NewView() {
	if m.owner() != nil {
		return
	}
	m.q.mu.Lock()
	m.view++
	m.q.mu.Unlock()
}

// fire calls the hooks of members whose pending work went from none to
// some, with no lock held. armed is true when an owner call fired them: an
// owner call made from inside such a hook is then refused. A hook fired from
// inside a Work call arms nothing, because the owner may legally be mid-call.
func fire(woken []*Member, armed bool) {
	for _, m := range woken {
		m.q.mu.Lock()
		hook := m.hook
		if armed {
			m.inHook = true
		}
		m.q.mu.Unlock()
		if hook != nil {
			hook()
		}
		if armed {
			m.q.mu.Lock()
			m.inHook = false
			m.q.mu.Unlock()
		}
	}
}

// Add asks for a job on this map's behalf. The same work is never queued
// twice: a second asker joins the first.
func (m *Member) Add(job scene.Job) error {
	err := m.owner()
	if err != nil {
		return err
	}
	err = valid(job)
	if err != nil {
		return err
	}
	m.q.mu.Lock()
	before := m.pendingLocked()
	m.q.addLocked(m, job)
	woke := before == 0 && m.pendingLocked() > 0
	m.q.mu.Unlock()
	if woke {
		fire([]*Member{m}, true)
	}
	return nil
}

// valid reports whether a job can be queued.
func valid(job scene.Job) error {
	if job == nil {
		return internal()
	}
	if job.Key() == "" {
		return internal()
	}
	if job.Kind().Validate() != nil {
		return internal()
	}
	return nil
}

// addLocked queues job for m, or joins m to the job if it already waits or
// flies, and then holds the queue to its limit.
func (q *Queue) addLocked(m *Member, job scene.Job) {
	key := job.Key()
	if f, ok := q.flying[key]; ok {
		f.entry.wanters[m] = true
		return
	}
	for _, e := range q.waiting {
		if e.job.Key() != key {
			continue
		}
		e.wanters[m] = true
		if m.view > e.view {
			e.view = m.view
		}
		return
	}
	q.seq++
	q.waiting = append(q.waiting, &entry{job: job, wanters: map[*Member]bool{m: true}, view: m.view, seq: q.seq})
	if len(q.waiting) > q.limit {
		q.dropOneLocked()
	}
}

// dropOneLocked drops the oldest job of a view no map is showing any more,
// or failing that the oldest job.
func (q *Queue) dropOneLocked() {
	victim, stale := -1, false
	for i, e := range q.waiting {
		old := true
		for w := range e.wanters {
			if e.view >= w.view {
				old = false
			}
		}
		better := victim < 0 || (old && !stale) || (old == stale && e.seq < q.waiting[victim].seq)
		if better {
			victim, stale = i, old
		}
	}
	if victim >= 0 {
		q.waiting = append(q.waiting[:victim], q.waiting[victim+1:]...)
	}
}

// Keep tells the queue which of this map's jobs are still wanted. A waiting
// job nobody wants is dropped; one in flight is cancelled and its outcome
// ignored. wanted is called with the queue's lock held and must not call
// the queue.
func (m *Member) Keep(wanted func(key string) bool) error {
	err := m.owner()
	if err != nil {
		return err
	}
	if wanted == nil {
		return internal()
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	m.q.releaseLocked(m, wanted)
	kept := m.later[:0]
	for _, d := range m.later {
		if wanted(d.job.Key()) {
			kept = append(kept, d)
		}
	}
	m.later = kept
	return nil
}

// releaseLocked withdraws m from every job it no longer wants.
func (q *Queue) releaseLocked(m *Member, wanted func(key string) bool) {
	kept := q.waiting[:0]
	for _, e := range q.waiting {
		if e.wanters[m] && !wanted(e.job.Key()) {
			delete(e.wanters, m)
		}
		if len(e.wanters) > 0 {
			kept = append(kept, e)
		}
	}
	q.waiting = kept
	for key, f := range q.flying {
		if !f.entry.wanters[m] || wanted(key) {
			continue
		}
		delete(f.entry.wanters, m)
		if len(f.entry.wanters) == 0 {
			f.dropped = true
			f.cancel()
		}
	}
}

// Leave takes the map out of the queue, and with it the work nobody else
// wants. Every later call returns the closed kind.
func (m *Member) Leave() {
	if m == nil || m.q == nil || m.closed {
		return
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	m.q.releaseLocked(m, func(string) bool { return false })
	m.later, m.closed = nil, true
}
