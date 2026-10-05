# Reviewer report — PLAN red team, architecture and code

Filed verbatim; machine paths redacted.

**PLAN-exit red team: go-tuiMaps v0.2.0 and watchpost 0.18.0**

The trees were not edited. One throwaway probe test ran in a scratch copy at `…/scratchpad/rt-plan-arch/lib/zz_changed_probe_test.go`.

Short names used below:
- **LP** = `<projects>/go-tuimaps/06_docs/02_features/radar-loops/04-development/implementation-plan.md`
- **WP** = `<projects>/watchpost/06_docs/02_features/observer-maps/04-development/implementation-plan.md`
- **IM** = `<projects>/go-tuimaps/06_docs/02_features/radar-loops/03-architecture-design/integration-map.md`
- **LIB** = `<projects>/go-tuimaps/`

---

### 1. Understandability

**F1. Neither plan says when playback plays.** Four questions are left open:
- Where the loop's phase starts. The shown frame is "a function of the animation time" (LP:139).
- What `Step` does while playback is on. L4.4 only tests it (LP:141).
- Whether a stepped position survives a refresh `Set`. `Seek` takes an index (0 = oldest, LP:141), so a refresh that drops the oldest frame silently moves the host's position.
- Whether stepping while Off contradicts L-1.10f. The approach's state diagram allows `Off --> Off: Step/Seek` (approach-1-loop-model.md:174), but L-1.10f says Off shows the newest frame (requirements.md:56).

A new engineer cannot build L4 from this without asking the author.
- Severity: Important. Simplify/Delete: N.
- Action: state the phase origin, the effect of step/seek while playing, and the rule that a held position follows its valid time across a re-`Set`. Resolve the conflict between Off and step/seek.

**F2. The integration map's diagram node IDs are shifted from the WP names** (IM:32-41). Node `L1` is WP-L2, `L3` is WP-L4, and so on.
- Severity: Minor. Simplify/Delete: Y.
- Action: make the node IDs match the WP names.

### 2. Maintainability

**F3 (Critical). Both plans misread what `Changed()` means.**
- In v0.1.0, `Render` itself raises the counter whenever it redraws (LIB `map.go:395-397`). `Work` never raises it (`work.go:17-40`).
- The probe confirmed this: after `Set` the counter read 5; after a `Render` 6; after `Work` still 6; after another `Render` 7, although the lines were identical to the previous frame.
- So the contract's claim that "the counter moves whenever a redraw would differ" (`contract.md:26`) is false. It is a sixth false claim that L1.3 does not list.

What this breaks:
- **L4.7** requires that a frame advance leave `Changed()` alone (LP:144). But an advance forces a redraw, which bumps the counter at `map.go:396`. No task changes that line, and no changelog row (L1.5) redefines `Changed`. The approach quietly redefines it as "data and look changes" (approach-1-loop-model.md:67-69).
- **`FrameTicks()` takes no time argument**, so it too can only move inside `Render`.
- **Watchpost's design** is "Redraws happen when `Changed()` or `FrameTicks()` move" (observer-maps approach-1-render-placement.md:48; WP:122-123). Neither counter can be a reason to call `Render`, because each moves only inside `Render`. A tile that lands in `Work` never reaches the screen. A frame advance never draws.
- Severity: Critical. Simplify/Delete: Y.
- Action: in L1/L4, define `Changed` as counting inputs only (data, look, view, tiles landed), never raised by `Render` itself, with `Work` raising it when it lands something visible. Add a changelog row. `Frame.Changed` and `Frame.Ticks` then carry the values the frame was drawn at, unambiguously.

**F4. Nothing says when the light-ground fallback is chosen** (LP:161). If it is decided from the library's own presets at build time, a host palette set through `SetPalette` (watchpost W7.5, WP:180) can make the blend fail with no warning.
- Severity: Important. Simplify/Delete: N.
- Action: re-run the check whenever the palette or ground changes, and surface the result in `Legend` or `Warnings`.

**F5. The integration map is shared across two repositories but is kept in step only by a rule** (IM:7: "same commit as either plan"). A watchpost commit cannot touch a go-tuiMaps file.
- Severity: Important. Simplify/Delete: N.
- Action: each plan records the IM commit it was reconciled against, and a docs-lane check refuses a mismatch.

### 3. Complexity

**F6. Playback per overlay id pushes state onto the host.**
- `Set` replaces the overlay (v0.1.0 D-74). The plan does not say whether playback survives the refresh `Set`, nor what `SetPlayback` does before the loop exists (LP:138).
- If playback resets on refresh, watchpost must re-apply its Setting after every refresh. That contradicts W8.9's "one `SetPlayback` call per Setting change" (WP:194) and L-1.13's "no playback state of its own".
- Watchpost has one Setting, and the ceiling is per map anyway.
- Severity: Important. Simplify/Delete: Y.
- Action: make `SetPlayback(p)` map-wide and persistent across `Set`. Keep `Step`, `Seek` and `Loop` per id.

**F7. Watchpost's `mapKey` memo duplicates the library's own frame reuse.** Once F3 is fixed, the simpler design is: call `Render` on every host mutation, on every `Work` that returns `did`, and on every `mapTickMsg`, and let the renderer's reuse make unchanged frames cheap.
- Severity: Important. Simplify/Delete: Y.
- Action: drop W2.1's key and the memo guard extended to it (WP:122). Keep guard 2 (WP:124) and guard 3 (WP:125).

### 4. Necessity: removable elements

**F8. Guard 4 cannot fire** (W2.5 at WP:126, W9.6 at WP:211). Under D-41, one goroutine calls `Render` and every mutator, so nothing can land between `Render` and storing the lines. The only concurrent caller is `Work`, which moves neither counter.
- On v0.1.0, W2.5's "store only if `Changed` did not move" refuses every frame that actually redrew, because of `map.go:396`.
- Severity: Important. Simplify/Delete: Y.
- Action: delete both tasks as guards. Keep L3.2's `Frame.Changed` and `Frame.Ticks` (LP:155) as the values a host stores beside the frame.

**F9. Other removable elements:**
- **`Describe` beside `Report`** (LP:174). D-58 allows the break. Two description surfaces are kept, and the plan never states `Describe`'s fate. Delete it in v0.2.0 once F10 is fixed.
- **`PurgeWithReport` beside `Purge`** (LP:222). Fold into `Purge() (PurgeReport, error)`.
- **Paired option and setter**: `WithImageBudget` with `SetImageBudget` (LP:126), and `WithBound` with `SetBound` (LP:199). Keep the setter only.
- **`CacheOption`**, a variadic type for a single option (LP:220). Use a plain `maxAge` parameter.
- **Public `Tables()` and `ProviderTable`** (LP:185). L-2.5's seam is internal. Export provider constants only.
- **`ProviderTable.Unverified string`** (LP:191). A legend flag is enough.
- **`OffBecause.NotOff`** (approach-1-loop-model.md:45). A double negative; the zero value can serve.
- **L3.9's interim severity derived from the role** (LP:162), which L5.1 replaces (LP:173). Move L5.1 ahead of L3 instead.
- **W1.13's self-registration** on top of W3.7's closed table (WP:115, WP:142). Two mechanisms for about six members. One literal table meets FR-3.8, and FR-9.3's "plug in" can be a single slice.

Severity: Important as a set. Simplify/Delete: Y.

### 5. Name semanticism

**F10. `Report`'s shapes contradict v0.1.0's and don't cover what it replaces** (LP:174-177).
- `PlaceAlert{DistanceKm float64; Bearing string}` ignores `Units()` (`describe.go:33`) and puts a compass word under a name that in v0.1.0 means degrees (`Answer.Bearing float64` beside `Compass string`, `internal/describe/answer.go:60-64`).
- `MotionReport.Relation` is a `string`, while `Answer.Relation` in the same package is the `Where` enum.
- `AlertShown.Feature` names a field that doesn't exist: `Feature` has no ID (`internal/overlay/store.go:50-57`). Watchpost cannot join an entry back to its alert (UGC codes, sender).
- `PlaceReport` has no image answers, although approach 4 promised them (approach-4-description.md:101). W9.2's "`Report` replaces `Describe`" (WP:207) therefore loses `Class` and `HeavierAt`.
- Severity: Important. Simplify/Delete: N.
- Action: add `Feature.ID`; reuse `Distance`, `Unit`, `Bearing`, `Compass` and `Where`; type the motion relation; add image answers to `PlaceReport`.

**F11. Smaller naming problems:**
- `Frame.Ticks` against `FrameTicks()` (LP:155, LP:144).
- `Drop`, `DropKind`, `DropAlertLabel` read as verbs (LP:164).
- `CheckBlend` takes positional arguments while `CheckRamp` takes a struct (LP:160; `kinds.go:118`).
- `Image.Provider` is a free string (LP:185).

Severity: Minor. Simplify/Delete: N.

### 6. Testability

**F12. Watchpost's `const mapMinBody = tuimaps.Size{…}` is not valid Go**: a struct cannot be a constant (WP:107). Untyped parameters appear in `refreshCost(layers, scope)` (WP:116) and `mapDepth(env, theme, profile)` (WP:176), so these shapes were never compiled.
- Severity: Minor. Simplify/Delete: N.
- Action: use `var`, and type the parameters.

**F13. W2.2's test, "no library call off the Bubble Tea goroutine", names no mechanism** (WP:123). `-race` cannot show which goroutine made a call.
- Severity: Minor. Simplify/Delete: N.
- Action: specify the wrapper or goroutine-identity shim the test uses.

### 7. Evidence soundness

**F14. W2.4's reference check disturbs what it checks** (WP:125). Its fresh `Render` bumps `Changed` (`map.go:396`) and resets the renderer's reuse state, which hides exactly the key bugs guard 3 exists to catch.
- Severity: Important. Simplify/Delete: N.
- Action: compare against a second `Map` built from the same inputs.

**F15. Tests that pin text instead of proving behaviour:**
- L1.3 closes RK-6 with "L1.2 green; a checklist row" (LP:112). L1.2 checks names only. That is the very failure RK-6 describes.
- L1.4 and L1.6 test that a sentence exists, not that it is true: "Change one sentence: the test fails" (LP:113, LP:115).
- Severity: Important. Simplify/Delete: N.
- Action: give each corrected claim a test of the behaviour it describes.

**F16. L6.9 proves the MRMS table on the wrong frames** (LP:193). It passes on "the OW-11 archive frames", which are IEM archive radar (D-51, rulings.md:68). Its threshold is not in the numbers table (LP:44-50).
- Severity: Important. Simplify/Delete: N.
- Action: set the threshold. State that the MRMS heavy end has no oracle before OW-12.

**F17. Checks that can report green without their proof point:**
- L10.5 runs the second architecture "on a runner that has it" (LP:234). A missing runner would pass as green.
- W9.2 declares the human-graded M1b "green" on the new source (WP:207), with no human re-score.
- Severity: Important. Simplify/Delete: N.
- Action: make the workflow fail when the architecture leg did not run, and require the HUM LEAD's M1b re-score.

### 8. Optimisation target

**F18. Both plans optimise for traceability and defensive evidence**: every row traced, a guard for every guard. The stated priority is D-58's simplicity and developer ergonomics, and D-48's matching of process to blast radius.
- The library plan adds about 45 exported names and 15 constants for one consumer.
- Watchpost builds four shims that exist only for P1-a (W2.5, W4.4, W3.8's `PurgeHost`, the path through `Describe`).
- Severity: Important. Simplify/Delete: Y.
- Action: apply F8–F9.

### 9. Long-term brittleness

**F19. Watchpost gates wall-clock timing inside `make verify`**: W2.9 and W8.12 assert p90 latencies, and W8.13 asserts a drift bound with live publishers (WP:130, WP:197-198). The library's plan refuses timed checks in its gate because they flake (LP:167).
- Severity: Important. Simplify/Delete: N.
- Action: gate structural proxies (request counts, "the first frame never waits for radar"), and record the times at SHIP.

**F20. W8.14 counts on the server's TTL to keep radar off disk** (WP:199). `httpx` writes whenever the server's expiry is past its floor (`platform/httpx/cache.go:256`).
- Severity: Minor. Simplify/Delete: N.
- Action: add an explicit per-request "memory only" option.

### 10. Evolution capacity

**F21. `LoopFrame` carries PNG bytes only** (LP:121). A forecast-grid loop, the likely next overlay kind, would need a second loop shape. The playback API itself generalises if it stays keyed by overlay id and independent of `Image`. Most of this is absorbed by the declarative model.
- Severity: Minor. Simplify/Delete: N.
- Action: record the limit in the contract.

### 11. Failure propagation

**F22. All of P1-b waits on the full v0.2.0 tag** (IM:55; W8 and W9 at WP:82-83). That tag needs WP-L10 (LP:99): hosted CI, the hour-long soak, five clean full runs (M6), OW-4, and the HUM LEAD's M1 grading.
- W9's non-radar items (bound, fetch options, purge, labels) are held hostage to work unrelated to radar.
- Integration is first attempted only after the tag, so defects like F3 surface after release.
- Separately, if the PENDING OW-2 specimen overturns D-17, HR-7 changes, yet W9.8 and W9.9 (WP:213-214) are not marked as dependent on it.
- Severity: Important. Simplify/Delete: N.
- Action: publish `v0.2.0-rc.N` tags per landed WP (they satisfy FR-5.6's "tagged"; scope L10.3's refusal of a release tag to final tags only), and mark W9.8 and W9.9 as dependent on OW-2.

### 12. Dependency completeness

**F23. Dependencies neither plan names:**
- The `Drop` type is defined in PENDING L3.11 but needed by L3.2 and L3.13, which are not pending (LP:155, LP:164, LP:166).
- L5.5's threshold comes from the provider table (L6), which is not in the dependency line (LP:245).
- L3.9 needs L5.1.
- W9.1 sets the bound "`WithBound` once" (WP:206), but the region follows the selection (watchpost requirements.md:40), so it needs `SetBound` whenever the selection changes region.
- W8 lists only W2 as a prerequisite, but W8.11 needs W5 (WP:82).
- External services: `govulncheck`'s vulnerability database at tag time, an arm64 hosted runner, and the SPC trigger for OW-12.
- Severity: Important. Simplify/Delete: N.
- Action: add these edges to both plans and to IM.

### 13. Scope calibration

**F24.** The set-up is over-built where F8–F9 point. It is under-specified on playback semantics (F1, F6), counter semantics (F3), `Report`'s identity and units (F10), what `SetImageBudget` does when lowered below the images already held (LP:126), and when the light-ground fallback is decided (F4).
- Severity: Important. Simplify/Delete: Y.
- Action: settle all of these as signatures and behaviour rules before BUILD.

---

**Verdict: not ready for BUILD.**

**Fix first: F3.** Define `Changed()` and `FrameTicks()` so that neither is raised inside `Render` and `Work` raises `Changed()` when it lands something. Put the change in the contract changelog, and rebuild watchpost's redraw triggers on it. As planned, the one signal the two releases share cannot tell watchpost when to draw.
