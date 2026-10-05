# Reviewer report — PLAN red team, a11y and infosec

Filed verbatim; machine paths redacted.

## PLAN red team: Accessibility (A11y) and InfoSec

Short names used below:
- **LP** = go-tuimaps `06_docs/02_features/radar-loops/04-development/implementation-plan.md`
- **WP** = watchpost `06_docs/02_features/observer-maps/04-development/implementation-plan.md`
- **LR** = the library's `01-objectives/requirements.md`
- **WR** = watchpost's `01-objectives/requirements.md`

Work was read-only. Nothing was written to either tree.

---

## Staff Accessibility Advocate

### A1: Is every control keyboard-operable and discoverable?

**A1-1. The watchpost window has no keys for playback. The only control is a Settings row.** *A11y.* The library adds step, seek and newest (D-25 item 2) so that a listener with motion off can still look through the frames. Its measure L-1.13 names watchpost as the host that proves the API. WP W8.9 wires only `SetPlayback` from a Settings row. So pausing a loop that is already playing means opening the Settings modal (WCAG 2.2.2), and step, seek and newest are never reachable.
- Severity: Important
- Evidence: WP:194, LP:148, LR:59, library rulings D-25/D-26
- Simplify/Delete? N
- Action: add a W8 task giving the map window its own keymap actions: play/pause, previous/next frame, newest. Each is listed in Help and can be overridden in `[keys]`. A PTY test drives them.

**A1-2. The description cannot be scrolled.** *A11y.* See A5-1.

### A2: Is information conveyed by more than colour, at every depth?

**A2-1. The fallback release (P1-a) carries severity by colour alone at colour-on depths, and W7.1 claims otherwise.** *A11y.*
- W7.1 asserts "every alert area present has a hatch or a label" at every depth the hint produces. But:
  - On v0.1.0 the hatch runs only at NoColour.
  - Labels are dropped whole when they do not fit (library D-28).
  - W9.9 admits that colour-on depths join the invariant only in P1-b.
- FR-7.1 names the Light theme, yet W7.1 sends only Monochrome to "none".
- RK-4 says only that the gap is recorded.

The cheap fix is already in watchpost's hands: the host writes `Feature.Label`, so P1-a can put the severity word in the label today.
- Severity: Important
- Evidence: WP:176, WP:214, WR:108, WR:153
- Simplify/Delete? Y (use the label, not new machinery)
- Action: W5.2 puts the severity word into every alert label. W7.1 states the depths its invariant covers in P1-a, and the Light theme's handling.

**A2-2. L3.9–L3.11 are PENDING, and the requirement they serve has no fallback if the specimen fails.** *A11y.* At Colours16 and NoColour, L-8.8 rests severity on the word and the dash. If OW-2 shows five dashes cannot be read, only the hatch is left, and it shares strokes (Extreme with Severe, Minor with Unknown).
- Severity: Minor
- Evidence: LP:162-164, LR:121, LR:128
- Simplify/Delete? N
- Action: name the ruled fallback now, for example "the word is mandatory, the dash is best effort".

### A3: Is there a speakable text alternative that answers the same questions?

**A3-1. P1-a's description is built on a `Describe` that cannot give per-alert answers.** *A11y.*
- In v0.1.0, `Describe` returns one `Answer` per place per overlay: `Relation`, `Distance`, and no severity.
- Library D-43 records that it "merge[s] every feature of an overlay".
- W1.4 says the description lists "the alerts shown, from the library's `Describe(places)`".
- W5.2 does not say whether alerts share an overlay. If they do, "inside" names no alert.

M1b, the only screen-reader metric in the ship-without-radar release, would then be scored on merged answers.
- Severity: Important
- Evidence: WP:106, WP:160; go-tuimaps `describe.go:19`; `internal/describe/answer.go:53-78`
- Simplify/Delete? N
- Action: W5.2 commits to one overlay per alert, so each answer names its alert. W1.4 takes the alert list, severity and times from watchpost's own snapshot, not from `Describe`.

**A3-2. The reading order and the "voice" claim have no owner.** *A11y.*
- In "with the picture" mode, a terminal screen reader reads 12–38 rows of braille before reaching any words.
- WR FR-7.4 says "the existing voice can speak it". The app's voice is the radio synth (`app/voices.go`). No task wires the description to it, and there is no key to speak it.
- Severity: Important
- Evidence: WP:106, WP:112, WR:111
- Simplify/Delete? Y (drop the voice claim if it is not built)
- Action: put the description first (above the map) whenever its mode is on, and test that order. Either add a "speak description" action or strike the voice sentence by ruling.

**A3-3. Missing zones are named by code.** *A11y.* "TXZ277" is spoken as letters and digits and means nothing to a listener.
- Severity: Minor
- Evidence: WP:162, WP:178
- Simplify/Delete? N
- Action: name missing zones from the alert's own area description, and extend W7.3's speech guard to cover zone codes.

### A4: Can motion and timing be avoided or adjusted? Any WCAG 2.3.1 risk?

**A4-1. Watchpost never states the default for map motion.** *A11y.* The library defaults to off (D-26). Watchpost's row is "builder-chosen" (FR-9.1), and W8.9 names no default. A default of "normal" auto-plays for more than five seconds (WCAG 2.2.2). D-27's link to `ReduceMotion` has also disappeared from W8.9.
- Severity: Important
- Evidence: WP:194, WR:130
- Simplify/Delete? N
- Action: state default = off (or slow), with a test. Say whether any app-wide reduce-motion reaches `ReduceMotion`.

**A4-2. P1-a has no motion control at all, and the change ceiling covers one map, not the screen.** *A11y.*
- If watchpost sets `Place.Blink`, the blink runs uncontrolled in P1-a (go-tuimaps `clock.go:118`).
- The 2.5-changes-a-second ceiling is per map, while the Observer ticks beside it every 300 ms and 50 ms (`modes/tty/dashboard.go:835,897`). This is deferred to F-180.
- The 2.3.1 risk is low: 2.5 is under 3, and the loop time grid is shared (L4.2).
- Severity: Minor
- Evidence: WP:194, LP:139
- Simplify/Delete? N
- Action: P1-a sets no blink, or passes `ReduceMotion` through. Record that the ceiling is per map only.

### A5: Does the design hold at small sizes and large fonts?

**A5-1. The degradation order leaves out the partial-area line, and the description cannot scroll.** *A11y.*
- W1.6's order is picture, then description, then notice. W5.4's line naming the missing zones (the M4 honesty line) is not in it.
- In an outbreak (specimen 30, 20 warnings), the description cannot fit at 69×12. There is no scroll or page key.
- W1.5's sweep passes on "the notice **with words**", which a single word satisfies.
- Severity: Important
- Evidence: WP:107-108, WP:162
- Simplify/Delete? N
- Action: rank the partial-area line inside the order, above the picture. Add scroll or paging keys for the description. Make the sweep assert that every alert's name and relation is reachable at every size.

### A6: Can the planned tests catch an accessibility regression, or only golden drift?

**A6-1. Switching the description to `Report` has no automated oracle for M1b.** *A11y.* M1b is human-graded once, in P1-a (W7.2). W9.2 swaps the description's source and asserts "M1b's recorded scenarios green", but nothing machine-checkable exists to be green. The library already has the pattern in `go-tuimaps/m1b_test.go`.
- Severity: Important
- Evidence: WP:177, WP:207
- Simplify/Delete? N
- Action: add an answer-key test per scenario (the place, each alert, the expected relation word present in the description). Run it on both sources.

**A6-2. The library's non-visual M1 arm grades data structures, not words.** *A11y.* The grader is the author, reading `Report` structs rather than the sentences a listener hears.
- Severity: Minor
- Evidence: LP:48, LP:238
- Simplify/Delete? N
- Action: score the arm on watchpost's worded output (W9.7).

---

## Principal InfoSec Engineer

### S1: Is every external input validated and bounded before use?

The core controls are sound: frames are copied at hand-in, the PNG header is checked before decode, file size is capped per pixel, `MaxFrames` exists, vertex caps exist (`internal/overlay/store.go:254-267`), and TileJSON credits are neutralised (W3.6).

**S1-1. Frame valid times are never bounded against the clock.** *InfoSec.* A source advertising future times, or a skewed clock, keeps the newest frame "current" forever. This defeats `stale` (L4.9) and M3, which is RK-3's exact harm.
- Severity: Minor
- Evidence: LP:122, LP:146, WP:189
- Simplify/Delete? N
- Action: refuse a valid time later than now plus a small skew, both at hand-in (L2.2) and in W8.4.

**S1-2. The advertised time list is unbounded against `MaxFrames = 36`.** *InfoSec.* If W8.6 hands in whatever the source lists, `Set` refuses the whole loop. A provider change then blanks radar.
- Severity: Minor
- Evidence: LP:122, WP:191
- Simplify/Delete? N
- Action: W8.6 trims to the newest frames within the cap, with a test.

### S2: Can input reach a path, cache, terminal escape or decoder unbounded?

**S2-1. Strings returned by `Report` are never cleaned before watchpost prints them.** *InfoSec.* The library cleans labels only when it draws them (`internal/render/basemap.go:200`). `AlertShown.Label`, `Drop.Shown`, alert and sender text, and zone names all return raw and are printed by W1.4 and W9.2, with no cleaning stated.
- Severity: Minor
- Evidence: WP:106, WP:207, LP:164, LP:174
- Simplify/Delete? N
- Action: route the description's lines through the existing text cleaner, and add W3.6's hostile-bytes test to the description.

### S3: What leaves the machine, to whom, and is the listener told?

**S3-1. Every radar request sends the exact on-screen rectangle, centred on the selected location.** *InfoSec.* The brief says so: "the exact rectangle on screen". The map follows the selection, so every refresh sends IEM or NCEP a timestamped box centred on the listener's station, usually a home. W1.12's test checks only that the right sources are named.
- Severity: Important
- Evidence: WP:188 (`Frame(ctx, t, box)`), WP:114; brief `project-brief.md:180`
- Simplify/Delete? Y
- Action: request radar per region (W4.1's regions) or snap the box to a coarse fixed grid, and crop locally. This reveals only the region and also makes FR-5.5's "fetch only frames not held" survive a selection change. The disclosure text should say what is sent.

**S3-2. Tiles go out under the library's `0.1.0-dev` agent until W9.3.** *InfoSec.* This is recorded under NFR-4 and accepted.
- Severity: Minor
- Evidence: WR:142
- Simplify/Delete? N
- Action: none.

### S4: Are network destinations confined to a closed list at every hop?

**S4-1. Radar's HTTP client does not do what RK-11 says it does.** *InfoSec.* RK-11 states the policy as "https only … private or link-local addresses refused after resolution". Radar goes through `platform/httpx`, which:
- builds a plain `net.Dialer` with `Proxy: http.ProxyFromEnvironment` (`platform/httpx/httpx.go:96-100`);
- refuses no private address anywhere in the tree;
- does not require https (redirects are only same-origin, `memo.go:194-203`).

No WP task closes this. TLS limits the harm, but the record asserts a control that does not exist.
- Severity: Important
- Evidence: WR:157, WP:190
- Simplify/Delete? Y (reuse the library's `CheckedDialer`, L8.4)
- Action: the radar client dials through `CheckedDialer` (or an equivalent control in httpx) and pins https. Test it with a resolver that answers 127.0.0.1.

**S4-2. The library's environment-read test cannot see the proxy variables.** *InfoSec.* L10.1 lists `os.Getenv` calls only. `HTTPS_PROXY` and `NO_PROXY`, read by net/http, can reroute every tile silently.
- Severity: Minor
- Evidence: LP:230
- Simplify/Delete? N
- Action: name the proxy variables in the list, or state whether the proxy is honoured.

### S5: Do errors, logs or cache names leak location or structure?

**S5-1. Radar errors carry the viewed rectangle.** *InfoSec.* httpx builds its error text with `RedactURL`, which masks keys but not `BBOX` (`platform/httpx/httpx.go:285-305`, `:648`). Any radar error that is shown or logged therefore holds the viewed rectangle. Tile file names and times are bounded by `MaxAge`, which is acceptable.
- Severity: Minor
- Evidence: WP:190
- Simplify/Delete? N (disappears with S3-1)
- Action: redact `BBOX` for radar hosts.

### S6: Is cached data kept no longer than stated, and can the listener clear it all?

**S6-1. "Clear map data" misses the zone record and races in-flight fetches.** *InfoSec.*
- **Zone geometry survives the clear.** Zones are fetched from `api.weather.gov` by `GetJSON` with no TTL (`domains/weather/nws/zones/zones.go:286`), so the server's max-age decides whether they are written to disk. W3.8 purges only "map hosts" and the library's `Purge` (W9.5) never reaches httpx. The record of which zones the listener's locations sit in survives.
- **Fetches in flight can rewrite the cache.** In P1-a, v0.1.0 has no cache generation, so a tile fetch already in flight can write back after the directory delete.
- **W9.5 overstates.** "Purge empties every map cache" is false without keeping `PurgeHost`.
- Severity: Important
- Evidence: WP:143, WP:210
- Simplify/Delete? N
- Action: give zone fetches an explicit TTL/Persist and add them to the clear's scope. Close the map before the P1-a delete, and test the clear with a fetch in flight. W9.5 keeps `PurgeHost` for radar and zone entries.

### S7: Are dependencies and versions pinned and scanned?

**S7-1. The library's new CI workflow has no pinning stated.** *InfoSec.* L10.5 creates `.github/workflows/gate.yml` but does not say actions are pinned by SHA or permissions are least-privilege. Watchpost's own workflows pin by SHA (`.github/workflows/ci.yml:27`).
- Severity: Minor
- Evidence: LP:82, LP:234
- Simplify/Delete? N
- Action: pin actions by SHA and set `permissions: contents: read`.

Everything else here is sound: the module is required at its tag with no local replace (W0.1, W8.1); the library pins `govulncheck`; and watchpost's `@latest` for `govulncheck` is deliberate (`Makefile:94`, `scripts/lint.sh:28`).

---

## Verdicts

- **Accessibility: not ready for BUILD.** The P1-a fallback cannot meet its own screen-reader metric as planned (A3-1). Keyboard playback (A1-1), the reading order (A3-2), the motion default (A4-1) and the missing M1b oracle (A6-1) are all task-level additions of about a day.
  - **Fix first: A3-1.** Commit to one overlay per alert, and build the alert list from watchpost's own data, so that M1b in the ship-without-radar release is scored on per-alert answers.
- **InfoSec: ready for BUILD once S3-1, S4-1 and S6-1 are folded in as tasks.** Nothing is Critical, and the library's input bounds are solid.
  - **Fix first: S3-1.** Request radar per region or on a coarse grid, not the exact on-screen rectangle. This removes a continuous, timestamped home-location feed to two third parties and simplifies caching at the same time.
