package work

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

func isKind(err error, k fault.Kind) bool {
	var f *fault.Error
	return errors.As(err, &f) && f.Kind() == k
}

// job is a test's slow work: a kind, a key, and what running it does.
type job struct {
	kind scene.JobKind
	key  string
	run  func(ctx context.Context) error
}

func (j job) Kind() scene.JobKind { return j.kind }
func (j job) Key() string         { return j.key }
func (j job) Run(ctx context.Context) error {
	if j.run == nil {
		return nil
	}
	return j.run(ctx)
}

func tile(key string) job { return job{kind: scene.KindTile, key: key} }

func member(t *testing.T) *Member {
	t.Helper()
	m, err := NewQueue(DefaultCap).Join()
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func add(t *testing.T, m *Member, jobs ...scene.Job) {
	t.Helper()
	for _, j := range jobs {
		if err := m.Add(j); err != nil {
			t.Fatal(err)
		}
	}
}

func drain(t *testing.T, m *Member) (keys int) {
	t.Helper()
	for range 10000 {
		did, err := m.Work(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !did {
			return keys
		}
		keys++
	}
	t.Fatal("Work never said it had nothing to do")
	return 0
}

// TestPendingCounts is plan task 07.1.
func TestPendingCounts(t *testing.T) {
	m := member(t)
	if m.Backlog() != 0 {
		t.Fatal("a new queue has work pending")
	}
	add(t, m, tile("t/1"), tile("t/2"), job{kind: scene.KindOverlayPrepare, key: "o/alerts"}, job{kind: scene.KindDescribe, key: "d/1"})
	add(t, m, tile("t/1")) // the same work is never queued twice
	if got := m.Backlog(); got != 4 {
		t.Errorf("Pending() = %d, want 4", got)
	}
	if err := m.Add(job{kind: 0, key: "x"}); !isKind(err, fault.Internal) {
		t.Errorf("a job with no kind: %v", err)
	}
	if err := m.Add(job{kind: scene.KindTile, key: ""}); !isKind(err, fault.Internal) {
		t.Errorf("a job with no key: %v", err)
	}
	if err := m.Add(nil); !isKind(err, fault.Internal) {
		t.Errorf("no job: %v", err)
	}
	if drain(t, m) != 4 || m.Backlog() != 0 {
		t.Errorf("after draining, %d pending", m.Backlog())
	}
}

// TestCapDropsOldest and TestNewestViewFirst are plan task 07.2.
func TestCapDropsOldest(t *testing.T) {
	q := NewQueue(3)
	m, _ := q.Join()
	add(t, m, tile("old/1"), tile("old/2"))
	m.NewView()
	add(t, m, tile("new/1"), tile("new/2"))
	if got := m.Backlog(); got != 3 {
		t.Fatalf("Pending() = %d with a cap of 3", got)
	}
	var ran []string
	var mu sync.Mutex
	record := func(key string) scene.Job {
		return job{kind: scene.KindTile, key: key, run: func(context.Context) error { mu.Lock(); ran = append(ran, key); mu.Unlock(); return nil }}
	}
	q2 := NewQueue(3)
	m2, _ := q2.Join()
	add(t, m2, record("old/1"), record("old/2"))
	m2.NewView()
	add(t, m2, record("new/1"), record("new/2"))
	drain(t, m2)
	if strings.Join(ran, " ") != "new/1 new/2 old/2" {
		t.Errorf("ran %v; the oldest job of a view no longer shown is dropped first, and the newest view runs first", ran)
	}
}

func TestNewestViewFirst(t *testing.T) {
	m := member(t)
	var ran []string
	record := func(key string, kind scene.JobKind) scene.Job {
		return job{kind: kind, key: key, run: func(context.Context) error { ran = append(ran, key); return nil }}
	}
	add(t, m, record("a", scene.KindTile), record("b", scene.KindTile))
	m.NewView()
	add(t, m, record("c", scene.KindTile), record("d", scene.KindDescribe), record("e", scene.KindTile))
	drain(t, m)
	if strings.Join(ran, " ") != "c d e a b" {
		t.Errorf("ran %v; want the newest view's jobs first, each view's in the order asked for", ran)
	}
}

// TestWorkDoesOneJob is plan task 07.3.
func TestWorkDoesOneJob(t *testing.T) {
	m := member(t)
	var stacks []string
	for _, key := range []string{"a", "b"} {
		add(t, m, job{kind: scene.KindTile, key: key, run: func(context.Context) error {
			buf := make([]byte, 4096)
			stacks = append(stacks, string(buf[:runtime.Stack(buf, false)]))
			return nil
		}})
	}
	did, err := m.Work(context.Background())
	if !did || err != nil || len(stacks) != 1 || m.Backlog() != 1 {
		t.Fatalf("one Work call: did=%v err=%v ran=%d pending=%d", did, err, len(stacks), m.Backlog())
	}
	if !strings.Contains(stacks[0], "TestWorkDoesOneJob") {
		t.Errorf("the job did not run on the caller's goroutine:\n%s", stacks[0])
	}
	did, _ = m.Work(context.Background())
	again, _ := m.Work(context.Background())
	if !did || again {
		t.Errorf("second call did=%v, third did=%v; want true, then false", did, again)
	}
}

// TestWorkCancel is plan task 07.4.
func TestWorkCancel(t *testing.T) {
	m := member(t)
	started := make(chan struct{})
	add(t, m, job{kind: scene.KindTile, key: "slow", run: func(ctx context.Context) error { close(started); <-ctx.Done(); return ctx.Err() }})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := m.Work(ctx); done <- err }()
	<-started
	if m.Backlog() != 0 || m.InFlight() != 1 {
		t.Errorf("while it runs: %d pending, %d in flight", m.Backlog(), m.InFlight())
	}
	cancel()
	err := <-done
	if !isKind(err, fault.Cancelled) || !errors.Is(err, context.Canceled) {
		t.Errorf("%v; want the cancelled kind", err)
	}
	if m.Backlog() != 0 || m.InFlight() != 0 {
		t.Errorf("after a cancel: %d pending, %d in flight; an abandoned job that no other map wants is gone", m.Backlog(), m.InFlight())
	}
	add(t, m, tile("slow")) // and can be asked for again
	if m.Backlog() != 1 {
		t.Error("the abandoned job's key is still held")
	}
	already, stop := context.WithCancel(context.Background())
	stop()
	if did, err := m.Work(already); did || !isKind(err, fault.Cancelled) {
		t.Errorf("Work with a context that has already ended: did=%v, %v", did, err)
	}
	if m.Backlog() != 1 {
		t.Error("a Work call that never started took a job with it")
	}
}

// TestWorkNeverWaitsOnWork is plan task 07.5: there is no limiter (D-84).
func TestWorkNeverWaitsOnWork(t *testing.T) {
	const wide = 8
	m := member(t)
	var inside atomic.Int32
	all := make(chan struct{})
	for i := range wide {
		add(t, m, job{kind: scene.KindTile, key: string(rune('a' + i)), run: func(context.Context) error {
			if inside.Add(1) == wide {
				close(all)
			}
			<-all
			return nil
		}})
	}
	var wg sync.WaitGroup
	for range wide {
		wg.Add(1)
		go func() { defer wg.Done(); m.Work(context.Background()) }()
	}
	select {
	case <-all:
	case <-time.After(5 * time.Second):
		t.Fatalf("only %d of %d Work calls were inside a job at once; something made one wait on another", inside.Load(), wide)
	}
	wg.Wait()
	src, err := os.ReadFile("work.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "about 1 MB") {
		t.Error("the note on what each further concurrent Work can cost is missing from the package's documentation (D-84)")
	}
}

// TestLeftViewCancelsJob is plan task 07.6.
func TestLeftViewCancelsJob(t *testing.T) {
	m := member(t)
	started := make(chan struct{})
	add(t, m, job{kind: scene.KindTile, key: "t/left", run: func(ctx context.Context) error { close(started); <-ctx.Done(); return ctx.Err() }}, tile("t/left-too"), tile("t/stays"))
	done := make(chan error, 1)
	go func() { _, err := m.Work(context.Background()); done <- err }()
	<-started
	m.NewView()
	if err := m.Keep(func(key string) bool { return key == "t/stays" }); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("a job cancelled because its view left is not the Work call's failure: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the in-flight job was not cancelled when its tile left the view")
	}
	if m.Backlog() != 1 {
		t.Errorf("%d pending; only the tile still in view should wait", m.Backlog())
	}
}

// TestChangeCounterMovesOnCompletion is plan task 07.7.
func TestChangeCounterMovesOnCompletion(t *testing.T) {
	m := member(t)
	before := m.Changed()
	add(t, m, tile("ok"), job{kind: scene.KindTile, key: "bad", run: func(context.Context) error { return errors.New("no such tile") }})
	if m.Changed() != before {
		t.Error("asking for work moved the counter; only finished work changes what a redraw would show")
	}
	m.Work(context.Background())
	first := m.Changed()
	_, err := m.Work(context.Background())
	if first == before || m.Changed() == first {
		t.Errorf("counter %d, %d, %d; it moves when a job finishes, and when one fails: the frame's status changes either way", before, first, m.Changed())
	}
	if err == nil {
		t.Error("a failed job's error was swallowed")
	}
}

// TestNextCallIsEarliest is plan task 07.8, and TestNothingDueWhenOffline 07.12.
func TestNextCallIsEarliest(t *testing.T) {
	m := member(t)
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	if _, due := m.NextCall(now); due {
		t.Fatal("nothing is set, yet something is due")
	}
	m.SetDeadline("blink", now.Add(800*time.Millisecond))
	m.SetDeadline("stale/radar", now.Add(20*time.Minute))
	if err := m.Defer(tile("t/failed"), now.Add(30*time.Second)); err != nil {
		t.Fatal(err)
	}
	if at, due := m.NextCall(now); !due || !at.Equal(now.Add(800*time.Millisecond)) {
		t.Errorf("NextCall = %v, %v; want the blink", at, due)
	}
	m.SetDeadline("blink", time.Time{}) // cleared: reduce-motion, say
	if at, due := m.NextCall(now); !due || !at.Equal(now.Add(30*time.Second)) {
		t.Errorf("NextCall = %v, %v; want the retry time", at, due)
	}
	if at, _ := m.NextCall(now.Add(time.Hour)); !at.Equal(now.Add(30 * time.Second)) {
		t.Errorf("a deadline already past is still the earliest: %v", at)
	}
}

func TestNothingDueWhenOffline(t *testing.T) {
	m := member(t)
	add(t, m, tile("t/1"))
	if _, due := m.NextCall(time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)); due {
		t.Error("waiting work is not a deadline: with nothing set and nothing deferred, nothing is due")
	}
}

// TestPendingCountsWaitingOnly is plan task 07.10.
func TestPendingCountsWaitingOnly(t *testing.T) {
	m := member(t)
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	if err := m.Defer(tile("t/failed"), now.Add(30*time.Second)); err != nil {
		t.Fatal(err)
	}
	if m.Backlog() != 0 {
		t.Error("a failed job waiting for its retry time is pending")
	}
	if n, _ := m.Promote(now.Add(29 * time.Second)); n != 0 || m.Backlog() != 0 {
		t.Error("promoted before its time")
	}
	if n, _ := m.Promote(now.Add(30 * time.Second)); n != 1 || m.Backlog() != 1 {
		t.Errorf("at its time: promoted %d, pending %d", n, m.Backlog())
	}
	if _, due := m.NextCall(now); due {
		t.Error("a promoted retry still counts as a deadline")
	}
}

// TestSettle is plan task 07.9.
func TestSettleEndsWhenIdle(t *testing.T) {
	m := member(t)
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	m.Defer(tile("t/backoff"), now.Add(time.Minute))
	add(t, m, tile("ok/1"), tile("ok/2"), job{kind: scene.KindTile, key: "bad", run: func(context.Context) error { return errors.New("refused") }})
	res, err := m.Drain(context.Background())
	if err != nil || res.Ran != 3 || res.Failed != 1 || m.Backlog() != 0 {
		t.Errorf("%+v, %v; work waiting for its retry time is not pending, so Settle ends", res, err)
	}
}

func TestSettleEndsOnContext(t *testing.T) {
	m := member(t)
	add(t, m, job{kind: scene.KindTile, key: "blocked", run: func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := m.Drain(ctx)
	if !isKind(err, fault.Cancelled) || time.Since(start) > 3*time.Second {
		t.Errorf("%v after %v", err, time.Since(start))
	}
}

func TestSettleSaysNoSource(t *testing.T) {
	m := member(t)
	noSource := fault.Make(fault.FetchRefused, textConst("no tile could be fetched"), textConst("no source is named and no embedded tiles are imported"), textConst("name a source, or import the assets package"))
	add(t, m, job{kind: scene.KindTile, key: "t/1", run: func(context.Context) error { return noSource }}, job{kind: scene.KindTile, key: "t/2", run: func(context.Context) error { return noSource }})
	res, err := m.Drain(context.Background())
	if err != nil || res.Failed != 2 || res.Why == nil || !strings.Contains(res.Why.Error(), "no source is named") {
		t.Errorf("%+v, %v; the result must say why nothing was fetched", res, err)
	}
}

func TestSettleReportsWorkInFlightElsewhere(t *testing.T) {
	m := member(t)
	started, release := make(chan struct{}), make(chan struct{})
	add(t, m, job{kind: scene.KindTile, key: "slow", run: func(context.Context) error { close(started); <-release; return nil }})
	go m.Work(context.Background())
	<-started
	res, err := m.Drain(context.Background())
	close(release)
	if err != nil || res.InFlight != 1 {
		t.Errorf("%+v, %v; Settle never waits on a Work running elsewhere: it returns and says how many", res, err)
	}
}

// TestNoPanicEscapes is plan task 07.11.
func TestNoPanicEscapes(t *testing.T) {
	m := member(t)
	add(t, m, job{kind: scene.KindTile, key: "boom", run: func(context.Context) error { panic("index out of range") }}, tile("after"))
	did, err := m.Work(context.Background())
	if !did || !isKind(err, fault.Internal) {
		t.Errorf("did=%v, %v; a panicking job is an internal error, not a crash", did, err)
	}
	if strings.Contains(err.Error(), "index out of range") {
		t.Errorf("%q repeats the panic's own text, which may hold outside data", err)
	}
	if did, err := m.Work(context.Background()); !did || err != nil || m.InFlight() != 0 {
		t.Errorf("the queue is not usable after a panic: did=%v, %v", did, err)
	}
}

// TestNoWorkCalledWarning is plan task 07.15.
func TestNoWorkCalledWarning(t *testing.T) {
	m := member(t)
	add(t, m, tile("t/1"))
	warned := 0
	for range 60 {
		if m.NoteRender() {
			warned++
		}
	}
	if warned != 1 {
		t.Errorf("warned %d times in 60 renders with work pending and no Work; want once, after 20", warned)
	}
	m2 := member(t)
	add(t, m2, tile("t/1"), tile("t/2"))
	for i := range 60 {
		if i == 10 {
			m2.Work(context.Background())
		}
		if m2.NoteRender() && i < 29 {
			t.Errorf("warned at render %d; a Work call restarts the count", i)
		}
	}
	m3 := member(t)
	for range 60 {
		if m3.NoteRender() {
			t.Error("warned with nothing pending")
		}
	}
}
