---
title: "v0.2.0 — REVIEW report"
date: 2026-10-04
phase: REVIEW
sev: SEV-0
authority: HUM LEAD
status: "APPROVED by the HUM LEAD 2026-10-04 (D-139): every finding fixed, ruled (D-129..D-138), or recorded with its reason; VALIDATE opens."
---

# REVIEW report — go-tuiMaps v0.2.0

## Dispatch

Eight reviewers with no prior context, each briefed from watchpost's `06_docs/red-team-brief.md` template,
read the library whole at `72cf0e5` - everything v0.1.0 shipped and everything v0.2.0 adds, the nested
modules, the gate and hosted CI (`review-scope.md`) - with no builds or tests: the four axes (code quality,
project hygiene, docs quality, business quality), li-A2DH's REVIEW phase lens (Principal QA Engineer), and
three personas - the Staff Accessibility Advocate and the Principal InfoSec Engineer, confirmed at PLAN, and
the Principal Performance Engineer, run under the standing clearance of 2026-10-03 for the HUM LEAD to
confirm or set aside.

| Reviewer | Verdict | Fix first |
|---|---|---|
| Code quality | do not ship | Settle held the map's lock with no deferred unlock: a recovered panic hung the next call; the panic proof never panicked with the lock held and named 36 of about 65 calls |
| Project hygiene | do not ship | the hosted full gate cannot pass as configured, and D-128 said hosted CI was green |
| Docs quality | do not ship | the README's pump example stopped at any error - the bug that blanked M1's radar - and broke the contract's wake rule |
| Business quality | do not ship (ship after a docs pass) | RK-8 marked closed while no listener hears the motion words |
| REVIEW lens (QA) | do not ship the final tag | the hosted full gate has never passed, and a ruling said it had |
| Accessibility | do not ship | `--describe` told old radar as "over the place now" with no out-of-date mark (Critical) |
| InfoSec | ship | the vulnerability scan measured a newer standard library than the module's floor allows |
| Performance | do not ship | `Render` waited on the disk cache's lock while a worker walked the whole cache (Critical) |

## The correction the review forced

**D-128 overstated hosted CI** (D-129): only the quick gate had been green on hosted CI; the one hosted full
gate (run 37181469202) failed on both architectures, seven fuzz legs at their 600 s limit, and the nightly
run has never fired because a schedule runs only from the default branch. BUILD exit stands as approved;
the hosted full gate is fitted to the runner (`GATE_FUZZ_SCALE`) and must be dispatched green on both
architectures before SHIP (checklist row 0.5).

## Every finding, and what was done

**Key:** fixed - changed and held by a test or a document; ruled - brought to the HUM LEAD; recorded - left
as it is, with the reason; v0.3.0 - an Owed row or the quality pass (follow-ups F-1, F-2).

### Defects in the library

| Finding | Disposition |
|---|---|
| Settle's lock without a deferred unlock (code 9) | fixed: `settleNote` defers it; a panic planted inside the locked region (`Settle.locked`) and the map answers within 5 s (`TestPanicRecoveredAtEveryPublicCall`) |
| The panic proof named 36 of about 65 calls; `Changed` unguarded (code 2, QA 5) | fixed: every public call is planted; `TestEveryPublicCallIsPanicTested` fails on any call left out; `Changed` guarded |
| `Render` waits on the disk cache's lock during a whole-cache walk (perf F12) | fixed: the tiles in view under a lock of their own (`TestTheViewIsNotedWithoutWaitingOnTheDisk`) |
| A reused frame stays stale after new shapes land (perf F15) | fixed: landed work in `OverlaysVersion` (`TestAFrameIsNotReusedOverNewShapes`) |
| The bucket in view never prepared once another is cached (perf F16) | fixed (`TestTheBucketInViewIsPrepared`) |
| Tile records kept for every tile ever viewed (perf F6) | fixed: let go when no view needs them (`TestTilesTheViewLeftAreLetGo`) |
| 576 MiB copied under the store's lock before the 6 MiB budget check (perf F13, QA 4, InfoSec F3) | fixed: the budget is asked of the host's headers first (`TestAnOverBudgetLoopIsRefusedBeforeItIsCopied`) |
| Web Mercator sampling repeats its trigonometry 16 times a dot (perf F3) | fixed: rows and columns placed once; identical answers (`TestAFootprintIsSampledAsEachPointAlone`) |
| The motion search tries up to 129² shifts for hourly frames, under the lock (perf F9) | fixed: every second shift then each around the best, past a reach of 8 blocks (`TestHourlyFramesAreMeasuredWithinBounds`) |
| Idle connections and net/http goroutines outlive `Close` (code) | fixed: `Fetcher.CloseIdle` at a source change and at `Close`; an idle timeout (`TestAFetcherLetsGoOfItsConnections`); `TestNothingOfTheMapRunsAfterClose` |
| `Answer.Cleaned` left five fields uncleaned (code 6) | fixed (`TestEveryTextOfAnAnswerIsCleaned`) |
| `no-work-called` never sent (code 4) | ruled D-131: wired (`TestAHostThatNeverWorksIsWarned`) |
| The released-id list grows for ever, read by nothing; the contract contradicts itself (code 4, perf F7) | ruled D-132: deleted, the contract corrected |
| Old motion told as now on the listener's path (a11y F4) | ruled D-130: `MotionReport.Stale`; the demo renders once before `--describe` (`TestOldMotionIsSaidToBeOld`, `TestAnOldLoopIsSaidToBeOld`) |
| 28 reachable standard-library vulnerabilities at the floor, go1.25.0 (InfoSec F2) | ruled D-133: every module at go 1.25.13; the gate scans at the floor and the local toolchain, every module at release |
| A cache root's owner never checked; macOS ACLs unseen (InfoSec F1) | ruled D-134: refused (`TestARootAnotherUserOwnsIsRefused`, `TestARootWithAnACLIsRefused`); the ACL check's build tags ratified, D-138 |
| Missing IPv6 forms in the private-address check (InfoSec F6) | fixed: `::/96`, `::ffff:0:0:0/96`, `100::/64` (`TestTheLesserReservedRangesAreRefused`) |
| Failed deletes in the disk cache dropped silently (InfoSec F5) | fixed: counted with the failed writes (`TestAFailedDeleteIsTold`) |
| Place ids skip the id rule (InfoSec F7) | fixed: at most 256 bytes, plain text (`TestAPlaceIDIsHeldToTheIDRule`); names: recorded, a host's own text, cleaned once a frame |
| An unparsable tile address told as leaving the source (QA 13) | fixed (`TestAnUnreadableAddressSaysSo`) |
| The run length defined twice (code 2) | fixed: `scene.RunLength` |
| Dead work-queue API: `Promote`, `SetDeadline`, `Member.Changed`, `Queue.Waiting`'s use outside tests (code 4) | fixed: deleted; `Waiting` a test helper |

### The gate, CI and the evidence

| Finding | Disposition |
|---|---|
| The hosted full gate cannot pass; the nightly run never fires (hygiene, QA 1) | fixed: `GATE_FUZZ_SCALE` divides every fuzz budget, written on the log line (M6 does not count it); 8 in the hosted full job; the schedule's limit said in the workflow. Dispatch green on both architectures before SHIP (row 0.5) |
| The release check fails open: a missing Done table, a split table, the working tree judged (QA 7, 12; code 9) | fixed: Done tables and rows counted; a dirty tree refused; the tag must be HEAD; the tag on the log line |
| M6 drops failed runs from the streak (code 9, QA 8) | fixed: any failed full run resets it; a failed run names its legs |
| The licence check skips modules not on disk (code 9) | fixed |
| No P10 leg in the gate (hygiene, code 8) | ruled D-136: a gate leg where the harness is installed; the package rows stand |
| NFR-22's break check never built (hygiene, QA) | ruled D-137: `TestNothingOfTheLastReleaseBreaksUnlisted` |
| The decoder oracle accepts any lost attribute (code 9) | fixed: exact on the fixture tiles; the allowance for damaged input alone |
| The static-rules walk could check nothing; the clock rule missed `time.Now` (code 9) | fixed: an empty walk refused; `Now`, `Since`, `Until`; the public `assets` package walked |
| The soak replays one static loop (QA 3) | fixed: a refresh sharing eleven frames once a minute; goroutines held; a soak with no sample refused |
| The Critical's regression test checks only that the frame changed (QA 6) | fixed: no heavy cell, some light (a mutant restoring the Critical fails it) |
| No benchmark runs in the gate (QA 9) | fixed: a smoke leg runs every benchmark once |
| `tables_test` drops parse errors; the module list misses `tools/atlas` (QA 10, code) | fixed |
| `scripts/uat` cannot build from a fresh clone (hygiene) | fixed |
| "real amd64 hardware NOT RUN" on amd64 (hygiene, code) | fixed |
| Quick CI's `GATE_ARCH_LEG` and re-exported PATH (code 10) | fixed |
| `GATE_ARCH_LEG=sibling` honoured on any machine (code 10) | recorded: logged on the run's line and never counted by M6 |
| Tests run on the floor toolchain only; no Windows or macOS CI (QA 11) | recorded: the NOT RUN line says so; Windows is cross-compiled |
| A clone cannot reproduce "P10 clean" (hygiene, code 8) | recorded: the gate's leg runs where the harness is; the ledger's mirror is tracked (D-118, D-136) |

### The demo app (D-135) and the example

| Finding | Disposition |
|---|---|
| The pump example stops at any error; one wake slot for two goroutines (docs 2) | fixed: goes on past a failed job, one slot a goroutine, every free slot filled (`TestThePumpGoesOnPastAFailedJob`) |
| The `?` panel shows no keys on 24 rows; modified and SS3 arrows misparse; no confirmation of toggles; a focused place hides the loop state; failures unsaid; "class 6"; wind's direction; auto-play; the wording pass (a11y F1-F3, F5-F7, F9, F11-F14) | ruled D-135: every one fixed and tested (`cmd/tuimaps/access_test.go`) |
| The demo reads files whole; prints the cache path raw (InfoSec F8, F9) | fixed (`TestAFileIsReadNoFurtherThanItsLimit`) |

### Record and documents

| Finding | Disposition |
|---|---|
| RK-8 closed against the evidence; M1's failure missing from the release notes; "no layer depends on colour alone"; MRMS's heaviest cores can draw blank, unsaid; OW-12's due date (business F1-F3, F6, F7; docs) | fixed in the requirements and the release notes |
| Contract: section 2's released ids, section 9's wind row, the header, the frame cap's "plus", the tile-need figure, "the description", host-facing limits, an accessibility section (docs, a11y F15) | fixed |
| Diagrams before v0.2.0's gates (hygiene) | fixed: L2-gates, L2-tiles redrawn; the atlas regenerated |
| The package doc's list of read-only calls (docs) | fixed |
| Checklist row 0.10 ticked for a check not yet run; row 0.9 (hygiene, QA 2) | fixed |
| L-12.5 held by a benchmark that asserts nothing; speed unchecked on real loops (docs) | OW-22 (a dense measure owed to VALIDATE); the speed caveat in the contract |
| v0.1.0's deferred list not carried (business F8) | OW-23 |
| Watchpost says "playing" from `Playing` (business F4) | OW-21, watchpost's own exit chain, before row 3.4 |
| No host voices motion (business F5) | OW-21 |

### Recorded, with the reason

| Finding | Why it stays |
|---|---|
| Quake age, fire strength, buoy and tide by colour alone, colliding at 16 colours (a11y F8) | D-112; OW-17 extended |
| No shade key on `Class`; the default depth ignores COLORTERM (a11y F10, F9) | additive or default-changing: OW-24, OW-25 |
| Decoding a hostile picture cannot be cancelled mid-frame (InfoSec F4) | bounded at about 0.21 s a frame; the contract's limits say so |
| A large user style walked rule by rule (InfoSec) | the host's own style; the contract's limits |
| A tile job's key hashed at each call, and the queue scanned for it (perf F1) | working the key out when the job is made costs more allocations on every plan than it saves (NFR-4's marker-turn count went from 63 to 72), and a memo shared by goroutines needs care: v0.3.0 |
| The timeline rebuilt several times a render; `Pow` a cell; frames hashed twice; quadratic eviction; the near-duplicate check; single-flight (perf F2, F4, F5, F10, F11, F17) | same output, no defect: v0.3.0 quality pass |
| `Report` recomputes alert projection on each pan (perf F14) | memoized by key; motion's frames are kept across pans; the alert half is v0.3.0 |
| `Frame.steps` grows with places asked about (perf F8) | bounded by the heavier pixels of one loop version |
| The tile state machine bypassed; code only the asset tool uses (`archive`, `fetch.Describe`, `RootCAs`, the schema seam); the provider registry; test-only state; dead branches; names (`view4`, `mustFloat`, `takeLocked`); repeated loops and formulas; parallel option structs (code 1-5) | maintainability, no defect: the v0.3.0 quality pass (F-1, F-2) |
| Function length: `compose` 78 lines, `HandIn` 68 (code 8) | the P10 check passes them; the quality pass splits them |
| M1b compares key to library only (code 7) | v0.3.0 |
| History comments in the ratified ledger's text (code 6) | the ledger is the HUM LEAD's to reword; the others REVIEW named are rewritten as current behaviour |
| A public checkpoint tag; rc tags; dead branches (hygiene 1) | at SHIP, with the work branch (D-33); deleting a public tag is the HUM LEAD's to approve |
| `.gitignore` names the harness (hygiene 1) | it keeps a clone from committing a local harness; `.git/info/exclude` would not travel |
