---
title: "v0.2.0 PLAN — Approach 2: the host fetcher's shape"
date: 2026-09-23
phase: PLAN
sev: SEV-0
authority: HUM LEAD
status: "RULED — D-55: A, the host supplies a transport; L-7.3 reworded (a recorded deviation from DISCOVER). Signatures and shape only."
---

# Approach 2 — the host fetcher's shape

**The question.** How does a host take over how tiles are fetched (L-7), for its own proxy, its own
user-agent, or its own TLS roots, while the library still bounds what comes back (L-7.3), stays
confined to the source's host (L-10), and says what the host takes on?

**Today.** `Fetcher = fetch.Func`, which is `func(ctx, fetch.Request) ([]byte, error)`. `fetch.Request`
is internal, so no host can write one without reflection (L-7.1). A host fetcher returns a whole body
the library can only measure after the fact. `fetch.Checked` exists to wrap it, is used nowhere, and
has no deadline (round 2, S-2). `Fetcher` takes effect only at the next `Source` call (F4).

## A deviation from DISCOVER that PLAN must record

**L-7.3 as ruled, "the library bounds the bytes and the time of every host fetch itself, whatever
the host fetcher does", cannot be met for time.** Since v0.1.0 the library never starts a goroutine:
the host's own goroutine drives all work through `Work`, and the gate enforces it statically
(`internal/rules`, `RuleGoroutine`). Without a goroutine of its own, no library can abandon a host
function that blocks and ignores its context. That call runs on the host's goroutine, inside the
host's `Work`, until the host's code returns.

**What can be promised, whatever the host code does:**
- **Bytes are bounded.** The library reads the body itself, through a limit, so it never holds more
  than the tile cap.
- **A late answer is never used.** After the host returns, the library checks the deadline and throws
  away a result that arrived too late.
- **The time is bounded only while the host's code honours its context.** A host that blocks and
  ignores the context blocks its own `Work` call, never anything else of the library's, and the
  contract says so.

**The alternative to the deviation** is to let the library start one goroutine per host fetch, so it
can walk away from a blocked call. That reverses a v0.1.0 design decision the whole work model rests
on. It is set out as option C below so it is a real choice, not a buried one.

## The options

### A — The host supplies a transport; the library keeps the client

```go
// FetchOptions is how a host shapes the library's own fetching (L-7.1, L-7.2).
type FetchOptions struct {
	Transport http.RoundTripper // nil: the library's own, with its private-address check
	UserAgent string            // the host's name for itself, added to the library's (L-7.2)
	Timeout   time.Duration     // per request; zero means the library's default
	AllowHTTP []string          // hosts the host lets be fetched over plain http (L-10.2)
}

func (m *Map) SetFetchOptions(o FetchOptions) error // takes effect at once (L-7.4)

// CheckedDialer is the library's dialer, with its private-address refusal
// (L-10.3), for a host transport that wants to keep it.
func CheckedDialer() func(ctx context.Context, network, addr string) (net.Conn, error)
```

The library keeps its own `http.Client`. It builds each request, confines redirects to the one
allow-list (L-10.2), sets the user-agent and reads the body through its limit. The host swaps only the
transport underneath: its proxy, TLS roots or connection pool, or a transport that serves from its own
store. A host that brings its own transport also brings its own dialing, so it loses the
private-address refusal unless it uses `CheckedDialer`, and the contract says so.

- **For:** it is Go's own extension point, so there is no new function type to learn. The library keeps
  everything it can keep: request shape, redirects, headers, the body limit, the late-answer rule.
  The host gets what hosts actually ask for.
- **Against:** a host that wants tiles from somewhere that is not HTTP (a file, a bundle) has to wrap
  it as a transport, which is awkward. `net/http` also waits on the transport, so the time caveat above
  applies here too.

### B — The host supplies a streaming fetch function

```go
type FetchRequest struct{ URL string; RangeStart, RangeLen, MaxBytes int64 }

// Fetch returns the body as a stream the library reads itself, through its
// limit, checking the context between reads.
type Fetch func(ctx context.Context, r FetchRequest) (io.ReadCloser, error)

func (m *Map) SetFetch(f Fetch) error // takes effect at once (L-7.4)
```

It is today's shape, exported and turned into a stream. The host owns the whole request, and the
library owns the reading.

- **For:** the most freedom. Any source works (HTTP, files, a cache, a test double), and the reading
  (bytes, late answers) stays the library's.
- **Against:** the host takes on more, and the library can promise less. Redirects, the host allow-list
  and the private-address check are all the host's once it builds the request, and L-10.2 then holds
  only for the TileJSON document, not for the connection. There is also more to state in the contract.

### C — Either A or B, plus one library goroutine per host fetch

```go
// the same API as A or B; inside, each host call runs in a goroutine the
// library abandons at the deadline
```

This is the only option that makes L-7.3's time promise hold whatever the host code does.

- **For:** the promise as ruled.
- **Against:** it reverses v0.1.0's "the library starts no goroutine" (a gate rule, and the base of the
  host-driven `Work` model), for the benefit of a host that ignores its own context. An abandoned
  call keeps running, holding whatever it holds, so the leak moves somewhere else rather than going
  away.

## Recommendation

**A, with the deviation recorded:** L-7.3 is reworded to what the library can promise — bytes bounded
whatever the host does, a late answer never used, and time bounded while the host's code honours its
context — and the contract states the rest.

A keeps more of the library's guarantees than B: redirects, the allow-list and the user-agent stay
the library's. It is also the shape Go programmers already know. The time caveat is the same in A and
B, and only C removes it, at a price v0.1.0 decided not to pay.

**The strongest argument against A.** B is more general. A host that serves tiles from a file or a
bundle has to write an `http.RoundTripper` to do it, which is more ceremony than a function. And a
future host may want exactly that for an offline mode.

## Cross-reference with watchpost 0.18.0

| Watchpost | Here |
|---|---|
| HR-6 (L-7): a host can write a fetcher | A: `SetFetchOptions` with its own `Transport` |
| NFR-4: tiles under the host's own user-agent | `FetchOptions.UserAgent` (L-7.2) |
| FR-5.7: radar bodies capped as they are read | Watchpost's own `httpx`. The library never fetches radar (L-1.4); this approach is for tiles only |
| HR-10 (L-10): tile-host confinement | Kept by the library's client under A; the host's under B |
