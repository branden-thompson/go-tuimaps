package work

import (
	"context"
	"errors"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
)

// SettleResult is what one Settle did.
type SettleResult struct {
	Ran      int   // jobs run
	Failed   int   // of those, how many failed
	InFlight int   // this map's jobs still inside a Work call elsewhere; Settle never waits for them
	Why      error // the first failure, so a result with nothing fetched says why
}

// Work runs at most one job, on the caller's goroutine, and says whether it
// ran one. It may be called from any goroutine, any number at once: no Work
// waits on another. It must not be called from the interface goroutine - a
// fetch can take seconds.
func (m *Member) Work(ctx context.Context) (did bool, err error) {
	if m == nil || m.q == nil || ctx == nil {
		return false, internal()
	}
	if ctx.Err() != nil {
		return false, problem(fault.Cancelled)
	}
	e, run, err := m.takeLocked(ctx)
	if err != nil {
		return false, err
	}
	if e == nil {
		return false, nil
	}
	failure := safely(run, e.job)
	return true, m.finish(ctx, e, failure)
}

// takeLocked picks the next job - the newest view's first, then the order
// asked for - and marks it in flight under a context of its own.
func (m *Member) takeLocked(ctx context.Context) (*entry, context.Context, error) {
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	if m.closed {
		return nil, nil, problem(fault.Closed)
	}
	m.idle, m.warned = 0, false
	best := -1
	for i, e := range m.q.waiting {
		if best < 0 || e.view > m.q.waiting[best].view || (e.view == m.q.waiting[best].view && e.seq < m.q.waiting[best].seq) {
			best = i
		}
	}
	if best < 0 {
		return nil, nil, nil
	}
	e := m.q.waiting[best]
	if m.q.flying[e.job.Key()] != nil {
		return nil, nil, internal() // the same work never both waits and flies
	}
	m.q.waiting = append(m.q.waiting[:best], m.q.waiting[best+1:]...)
	run, cancel := context.WithCancel(ctx)
	m.q.flying[e.job.Key()] = &flight{entry: e, cancel: cancel}
	return e, run, nil
}

// safely runs a job with no lock held and turns a panic into an error.
func safely(ctx context.Context, job scene.Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = internal() // the panic's own text may hold outside data and is not repeated
		}
	}()
	return job.Run(ctx)
}

// finish records a job's outcome. A job nobody wants any longer is
// forgotten. A job given up because this call was cancelled goes back to
// the queue if another map still wants it, and that map's hook fires from
// here - inside the cancelled call - or its sleeping pump would never wake.
func (m *Member) finish(ctx context.Context, e *entry, failure error) error {
	m.q.mu.Lock()
	f := m.q.flying[e.job.Key()]
	delete(m.q.flying, e.job.Key())
	if f != nil {
		f.cancel()
	}
	if f == nil {
		m.q.mu.Unlock()
		return internal() // a job that ran was not recorded as in flight
	}
	if f.dropped {
		m.q.mu.Unlock()
		return nil
	}
	if ctx.Err() != nil {
		woken := m.q.giveBackLocked(m, e)
		m.q.mu.Unlock()
		fire(woken, false)
		return problem(fault.Cancelled)
	}
	for w := range e.wanters {
		w.changed++
	}
	m.q.mu.Unlock()
	return own(failure)
}

// giveBackLocked returns a job to the queue for the maps that still want it,
// without the one that gave it up, and lists those it wakes.
func (q *Queue) giveBackLocked(quitter *Member, e *entry) []*Member {
	delete(e.wanters, quitter)
	if len(e.wanters) == 0 {
		return nil
	}
	var woken []*Member
	for w := range e.wanters {
		if w.pendingLocked() == 0 {
			woken = append(woken, w)
		}
	}
	q.waiting = append(q.waiting, e)
	return woken
}

// own makes a job's failure one of the library's own errors. The packages
// that supply jobs return those already; anything else is a defect, and its
// text is not repeated.
func own(failure error) error {
	if failure == nil {
		return nil
	}
	var f *fault.Error
	if errors.As(failure, &f) {
		return f
	}
	return internal()
}

// Settle is the pump's loop run on the caller's goroutine: Work until
// nothing is pending or the context ends. Work that failed and waits for
// its retry time is not pending, so Settle always ends; and it never waits
// on a Work running elsewhere - it returns and says how many there are.
func (m *Member) Drain(ctx context.Context) (SettleResult, error) {
	var res SettleResult
	if m == nil || m.q == nil || ctx == nil {
		return res, internal()
	}
	for range maxSettle {
		did, err := m.Work(ctx)
		var f *fault.Error
		if errors.As(err, &f) && (f.Kind() == fault.Cancelled || f.Kind() == fault.Closed) {
			return res, err
		}
		if !did {
			break
		}
		res.Ran++
		if err != nil {
			res.Failed++
			if res.Why == nil {
				res.Why = err
			}
		}
	}
	res.InFlight = m.InFlight()
	return res, nil
}

// Defer holds a failed job until the time it may be tried again. It is not
// pending, and wakes nobody, until Promote finds its time has come.
func (m *Member) Defer(job scene.Job, notBefore time.Time) error {
	err := m.owner()
	if err != nil {
		return err
	}
	err = valid(job)
	if err != nil {
		return err
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	for i, d := range m.later {
		if d.job.Key() == job.Key() {
			m.later[i].at = notBefore
			return nil
		}
	}
	m.later = append(m.later, deferred{job: job, at: notBefore})
	return nil
}

// Promote queues every deferred job whose time has come, and says how many.
// The map calls it at each owner call, with the wall clock the host passed
// in: nothing in the library can make a retry come due by itself.
func (m *Member) Promote(now time.Time) (int, error) {
	err := m.owner()
	if err != nil {
		return 0, err
	}
	m.q.mu.Lock()
	before, promoted := m.pendingLocked(), 0
	kept := m.later[:0]
	for _, d := range m.later {
		if d.at.After(now) {
			kept = append(kept, d)
			continue
		}
		m.q.addLocked(m, d.job)
		promoted++
	}
	m.later = kept
	woke := before == 0 && m.pendingLocked() > 0
	m.q.mu.Unlock()
	if woke {
		fire([]*Member{m}, true)
	}
	return promoted, nil
}

// SetDeadline names a moment the host should call again by: a marker's next
// phase, an overlay going stale. The zero time clears it.
func (m *Member) SetDeadline(name string, at time.Time) {
	if m.owner() != nil || name == "" {
		return
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	if at.IsZero() {
		delete(m.deadlines, name)
		return
	}
	m.deadlines[name] = at
}

// NextCall is the earliest moment anything is due: a named deadline, or a
// deferred job's retry time. Waiting work is not a deadline, so an offline
// map with nothing set reports nothing due.
func (m *Member) NextCall(now time.Time) (time.Time, bool) {
	if m == nil || m.q == nil {
		return time.Time{}, false
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	var earliest time.Time
	for _, at := range m.deadlines {
		if earliest.IsZero() || at.Before(earliest) {
			earliest = at
		}
	}
	for _, d := range m.later {
		if earliest.IsZero() || d.at.Before(earliest) {
			earliest = d.at
		}
	}
	return earliest, !earliest.IsZero()
}

// NoteRender is called by each render. It reports true once, when twenty
// renders in a row have found work pending and no Work has been called: the
// silent failure a host meets when it forgets the pump.
func (m *Member) NoteRender() bool {
	if m.owner() != nil {
		return false
	}
	m.q.mu.Lock()
	defer m.q.mu.Unlock()
	if m.pendingLocked() == 0 {
		m.idle = 0
		return false
	}
	m.idle++
	if m.idle < noWorkRenders || m.warned {
		return false
	}
	m.warned = true
	return true
}
