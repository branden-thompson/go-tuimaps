package work

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/scene"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

func textConst(s string) textsafe.Text { return textsafe.Clean(s) }

// TestOnPendingFiresOnceInsideOwnerCall is plan task 07.13, first part.
func TestOnPendingFiresOnceInsideOwnerCall(t *testing.T) {
	m := member(t)
	fired := 0
	var pendingSeen int
	if err := m.OnPending(func() { fired++; pendingSeen = m.Backlog() }); err != nil {
		t.Fatal(err)
	}
	add(t, m, tile("t/1"))
	if fired != 1 {
		t.Fatalf("the hook fired %d times before Add returned; want once", fired)
	}
	if pendingSeen != 1 {
		t.Errorf("the hook saw %d pending; it runs with no lock held, so Pending can be read from it", pendingSeen)
	}
	add(t, m, tile("t/2"), tile("t/3"))
	if fired != 1 {
		t.Errorf("the hook fired again while work was already pending: %d", fired)
	}
	drain(t, m)
	add(t, m, tile("t/4"))
	if fired != 2 {
		t.Errorf("none to some a second time: fired %d, want 2", fired)
	}
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	drain(t, m)
	m.Defer(tile("t/retry"), now)
	if fired != 2 {
		t.Error("a deferred job is not pending and must not wake the pump")
	}
	m.Promote(now)
	if fired != 3 {
		t.Errorf("a retry coming due at an owner call: fired %d, want 3", fired)
	}
}

// TestOwnerCallFromHookRefused is plan task 07.13, second part.
func TestOwnerCallFromHookRefused(t *testing.T) {
	m := member(t)
	var inHook []error
	m.OnPending(func() {
		inHook = append(inHook, m.Add(tile("from/hook")), m.Keep(func(string) bool { return true }))
		_, err := m.Promote(time.Time{})
		inHook = append(inHook, err, m.OnPending(nil))
	})
	add(t, m, tile("t/1"))
	if len(inHook) != 4 {
		t.Fatalf("the hook did not run: %v", inHook)
	}
	for i, err := range inHook {
		if !isKind(err, fault.ReentrantCall) {
			t.Errorf("owner call %d from inside the hook: %v; want the reentrant-call kind", i, err)
		}
	}
	if m.Backlog() != 1 {
		t.Errorf("%d pending; the refused Add must not have queued anything", m.Backlog())
	}
	if err := m.Add(tile("t/2")); err != nil {
		t.Errorf("after the hook returned, owner calls must work again: %v", err)
	}
}

// TestTwoWidePumpBothWake is plan task 07.13, third part: the pump as the
// contract draws it - a channel with one slot a goroutine, a wake that
// fills every free slot and never blocks.
func TestTwoWidePumpBothWake(t *testing.T) {
	m := member(t)
	const wide = 2
	wake := make(chan struct{}, wide)
	m.OnPending(func() {
		for range wide {
			select {
			case wake <- struct{}{}:
			default:
			}
		}
	})
	var mu sync.Mutex
	workers := map[int]int{}
	both := make(chan struct{})
	var once sync.Once
	var wg sync.WaitGroup
	ctx, stop := context.WithCancel(context.Background())
	for id := range wide {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case <-wake:
				}
				for {
					did, _ := m.Work(ctx)
					if !did {
						break
					}
					mu.Lock()
					workers[id]++
					if len(workers) == wide {
						once.Do(func() { close(both) })
					}
					mu.Unlock()
				}
			}
		}()
	}
	gate := make(chan struct{})
	var held atomic.Int32
	slow := func(key string) scene.Job {
		return job{kind: scene.KindTile, key: key, run: func(context.Context) error {
			if held.Add(1) == wide {
				close(gate)
			}
			<-gate // no job finishes until both goroutines hold one
			return nil
		}}
	}
	add(t, m, slow("a"), slow("b"), slow("c"), slow("d"))
	select {
	case <-both:
	case <-time.After(5 * time.Second):
		t.Fatalf("one burst of four jobs was worked by %d goroutine(s): %v", len(workers), workers)
	}
	stop()
	wg.Wait()
}

// TestSharedJobReturnsToQueue is plan task 07.14.
func TestSharedJobReturnsToQueue(t *testing.T) {
	q := NewQueue(DefaultCap)
	a, _ := q.Join()
	b, _ := q.Join()
	var bWoken atomic.Int32
	b.OnPending(func() { bWoken.Add(1) })
	started := make(chan struct{}, 2)
	shared := job{kind: scene.KindTile, key: "t/shared", run: func(ctx context.Context) error {
		started <- struct{}{}
		<-ctx.Done()
		return ctx.Err()
	}}
	add(t, a, shared)
	add(t, b, shared) // both maps want the same tile: one job
	if a.Backlog() != 1 || b.Backlog() != 1 || q.Waiting() != 1 {
		t.Fatalf("a=%d b=%d waiting=%d; two maps wanting one tile is one job, pending for both", a.Backlog(), b.Backlog(), q.Waiting())
	}
	woken := bWoken.Load()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := a.Work(ctx); done <- err }()
	<-started
	if b.Backlog() != 0 {
		t.Error("a job inside a Work call is not pending, for any map")
	}
	cancel()
	if err := <-done; !isKind(err, fault.Cancelled) {
		t.Errorf("%v", err)
	}
	if q.Waiting() != 1 || b.Backlog() != 1 {
		t.Errorf("waiting=%d, b pending=%d; the other map still wants the tile, so the job returns to the queue", q.Waiting(), b.Backlog())
	}
	if bWoken.Load() != woken+1 {
		t.Errorf("the other map's hook fired %d times; it must fire from inside the cancelled call, or its sleeping pump never wakes", bWoken.Load()-woken)
	}
	if a.Backlog() != 0 {
		t.Error("the map whose Work was cancelled gave the job up; it does not want it back")
	}
}

func TestLeavingAQueue(t *testing.T) {
	q := NewQueue(DefaultCap)
	a, _ := q.Join()
	b, _ := q.Join()
	add(t, a, tile("only/a"), tile("both"))
	add(t, b, tile("both"))
	a.Leave()
	if q.Waiting() != 1 || b.Backlog() != 1 {
		t.Errorf("waiting=%d; a map that closes takes with it only the work nobody else wants", q.Waiting())
	}
	if err := a.Add(tile("x")); !isKind(err, fault.Closed) {
		t.Errorf("Add after Leave: %v", err)
	}
	if did, err := a.Work(context.Background()); did || !isKind(err, fault.Closed) {
		t.Errorf("Work after Leave: did=%v, %v", did, err)
	}
}
