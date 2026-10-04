---
title: "v0.2.0 — BUILD-exit red team"
date: 2026-10-04
phase: BUILD (exit)
sev: SEV-0
authority: HUM LEAD
status: "ROUND 2 REMEDIATED — round 1's decisions ruled (D-108..D-121) and built; M1 recorded (D-122, D-123); round 2 reviewed that change, its behaviour questions ruled (D-124..D-127), every other finding fixed or recorded with its reason."
---

# BUILD-exit red team

## Dispatch

Eight reviewers with no prior context, each briefed from watchpost's `06_docs/red-team-brief.md` template
(copied, the five fields filled), each on one lens: the four axes (code quality, project hygiene, docs
quality, business quality), li-A2DH's BUILD phase lens, and three personas - Staff Accessibility Advocate
and Principal InfoSec Engineer (confirmed at PLAN), and Principal Performance Engineer, **run under the HUM
LEAD's standing clearance of 2026-10-03 for the HUM LEAD to confirm or set aside**: BUILD added hot-path
render work (allocation budgets, the blink fix, M4) that matches its trigger. Scope: the change since
`v0.1.0` at the BUILD-exit commit (161 Go files, about 16,600 lines added; 217 documents), read from git,
no builds or tests.

| Reviewer | Verdict | Fix first |
|---|---|---|
| Code quality | do not ship yet | the release check failed open (any last cell but `[ ]` passed; an unlogged override pointed it anywhere) |
| Project hygiene | do not ship | the version: the plan said the release check held it, and nothing did |
| Docs quality | do not ship yet | the same, and the normative file contradicting rulings and code |
| Business quality | do not ship | the same; the record contradicts the release in five places |
| BUILD lens | do not ship yet | motion recomputed per place and per pan, under the map's lock |
| Accessibility | do not ship yet | a large alert loses every non-colour cue |
| InfoSec | ship with conditions | a hostile many-coloured picture could hold a worker for seconds a frame |
| Performance | do not ship | **a refreshed loop never decoded its new frames** - confirmed, Critical |

## The Critical finding

**A refreshed radar loop showed its old frames for good.** `overlayWork` asked `Raster` whether an image
still needed preparing, and `Raster` answers from a refreshed loop's stand-in (L-27.1) so that drawing never
blinks - so the new version read as prepared and was never decoded. Confirmed by a Map-level test
(`TestARefreshedLoopDrawsItsNewFrames`: heavy frames refreshed to light ones, settled, and the frame was
byte-identical), fixed by a store query that ignores stand-ins (`Store.OwnPicture`), held by the test and a
caught mutant. No earlier test went Set → Settle → Set → Settle at Map level.

## Every finding, and what was done

**Key:** fixed - changed and held by a test or a document; recorded - left as it is, with the reason;
HUM - changes behaviour, rendering or API, brought to the HUM LEAD.

| Finding | Disposition |
|---|---|
| Version `0.1.0-dev`; the claimed release check absent (business F1, hygiene F9, docs F10, BUILD 4, InfoSec 6) | fixed: `fetch.Version` "0.2.0"; `--release` refuses a tag that is not the library's version (`TestTheTagIsTheReleaseTheLibrarySays`); checklist row 2.7 |
| The release check failed open; `GATE_CHECKLIST` (code 9, 10) | fixed: every data row must end `[x]` (`TestARowIsDoneOnlyWhenItSaysSo`); the override deleted |
| Refreshed loop never decoded (perf F1) | fixed, above |
| Motion recomputed per place and pan (BUILD 1, perf F2) | fixed: each frame's cells kept on the frame and shared by every place (`TestTrackingKeepsEachFramesCells`); a 72-frame, 60-place report 2.16 ms / 3.0 MB → 1.10 ms / 141 KB (`BenchmarkAReportOverALongLoopWithManyPlaces`). The view in the memo key: recorded (the no-place answer depends on it) |
| Host bytes copied before any size check (BUILD 3) | fixed: a picture or frame over 8 MiB is refused before the copy (`TestAnOversizedFrameIsRefusedBeforeItIsCopied`) |
| Lab recomputed in the matcher's loop (InfoSec 1, perf F3) | fixed: worked out once a table and once a colour; the worst case (250,000 colours, 256 entries) measured at 0.21 s a frame (`BenchmarkAPictureOfEveryColour`); `TestANearColourTakesItsNearestEntry`. A cap on distinct colours: recorded (0.21 s is bounded) |
| Shared-set key without provider or lengths (InfoSec 2, 3) | fixed (`TestAPictureReadThroughItsProviderIsKeptApart`) |
| Reserved ranges incomplete (InfoSec 5) | fixed: NAT64 local use, site-local, Teredo, the documentation and protocol ranges (`TestTheLesserReservedRangesAreRefused`); L-10.3 updated |
| Nearby by the string "miles" (code 2) | fixed: answers carry km |
| An unmeasurable sighting read as 0 km (code 4) | fixed: skipped |
| `Step(MaxInt)` overflowed to the oldest frame (BUILD 8) | fixed (`TestTheLongestStepLandsAtTheEnd`) |
| A no-op Play counts as a change (BUILD 8) | recorded: `Changed` counts inputs (D-66); a pressed Play is one |
| A proof that skips reads as a pass (code 9) | fixed in `firsthost_test.go`; `purge_test.go`'s skip is a platform's (no `/dev/fd`), named NOT RUN |
| A 25-second "full green" run counted by M6 (hygiene F6) | fixed: M6 counts no full run under 900 s |
| The docs lane logs one tree and tests another (code 9) | fixed within D-41: the run's line says the tests saw unstaged files |
| Stale doc-comment names (code 5); history in comments (code 6, BUILD 10) | fixed: eleven names, and a sweep of twenty files |
| The normative file contradicts rulings and code (business F2-F5, docs F4-F6, a11y 4.3, 4.4) | fixed: L-12.2 and NFR-2 to D-68 with OW-14 for the revisit; L-1.10b, L-1.13, NFR-3 to D-67 and D-76; dash → digit; every row instrumented (no "NO INSTRUMENT YET" left; each test named exists); RK-8 open until M1 |
| README, contract, plan, checklist, release notes stale (business F6, F9-F11, docs F1-F3, F7-F9, F11-F14, hygiene F4, F5, F10, F11, F13) | fixed: the docs sweep; release notes list every break and deviation; checklist row 2.2 reads "additions plus §12's breaks"; work-branch hashes out of the checklist |
| Duplicate L-28 (docs F6) | fixed: the second is L-30 |
| Gate NOT RUN line, `--quick` header, the archive test (hygiene F7, F8) | fixed |
| FuzzAgree runs hand-transcribed (hygiene F14) | fixed: raw output filed in `programs/output/fuzzagree-L10.7/`; the count not recalibrated, with the reason |
| Two nested modules not tidy (hygiene F3) | fixed in the checklist: row 3.3, at the tag (the library's tag must exist first) |
| Demo app: no paging, Esc quits from a panel, no `d` in the key row, the non-visual path silent on alerts and motion (a11y 2.1-2.3, 3.2) | fixed: `fitted`, `keyEscape`, the key row, `inView` and `moved` (`TestTheMotionIsSaidWithNoPlace` over the recorded outbreak) |
| Host-facing limits unstated (BUILD 2, 7, 12; InfoSec 4; a11y 5.1) | fixed: the contract's "Limits a host should know" |
| 11 MB of analysis material in the module zip (hygiene F1) | recorded: the root package's own tests read `06_docs`; moving it needs a ruling |
| Dead branches (hygiene F2) | at SHIP, with the work branch (D-33) |
| Trace kept three times; timeline rebuilt per call; four parallel maps; test-only fields; `overlay.Label`; `OwnedBytes`; `dropPreparedLocked`'s name; frames hashed twice; near-duplicate check O(n²); returned slices shared (business F12, code 3, 4, 5, BUILD 9, 11, perf F4, F5, F10) | recorded: maintainability, no defect; v0.3.0 quality pass |
| Image sampling per cell, Mercator recomputed (perf F9) | recorded: same output, v0.3.0 |
| L-12.5 measured on a near-empty case (perf F12) | recorded: re-measured cost of the realistic report above; a dense-region frame-advance measure is owed to VALIDATE |
| Cache root owner unchecked (InfoSec 8) | recorded: shared-machine only, its impact map content |
| `govulncheck` on PATH trusted; `GATE_ARCH_LEG` honoured locally (code 10) | recorded: the oracle disclaims evasion; both are logged |
| The architecture leg's test runs `--quick` (code 7); `tables_test` drops parse errors (code 7) | recorded: the leg is exercised by the hosted run itself |
| A11y 3.1, 3.3, 3.4, 4.1, 4.2; business F7, F13; perf F6, F8; code 8; BUILD 5, 13; InfoSec 4's API | **HUM** - see the BUILD report's decisions list |
| M1 unscored (a11y 5.3) | **HUM** - the sitting is ready (`07-readiness/m1-sitting.md`) |

## Round 2 — the change since `v0.2.0-rc.35`

**Why a second round.** Round 1's fourteen decisions were ruled and built, and M1's first sitting failed
its non-visual arm, which brought a new way of measuring motion (D-122) and a fix to the demo app's pump
(it stopped at the first failed job, so `--offline` drew no radar). Three blind reviewers, briefed from the
same template, read the uncommitted change against `v0.2.0-rc.35` with no builds or tests (a gate was
running): code quality, business quality, and the Staff Accessibility Advocate. Each said "do not ship as
it stands"; none found a Critical defect in the code.

| Finding | Disposition |
|---|---|
| A pair of frames that meets nothing at any shift counted as rain that held (code 10) | fixed: not measured (`TestAPairWithNothingToMeetIsNotMeasured`) |
| The span claimed the whole loop when pairs were skipped (business F8) | fixed: `Span` and `From` are the time measured (`TestTheSpanIsTheTimeMeasured`); the M1 pin's 55 minutes now proves every pair measured |
| "Barely moved" followed by "coming closer" (a11y F2) | fixed: `Trend` is `Held` when not `Moving` (`TestRainThatBarelyMovedHeld`) |
| A place on an image wider than half the world measured off its west edge (code 10) | fixed (`TestAWideImageFindsTheNearestRain`) |
| A view inside a very large alert: no label, nothing in `Dropped`; the label at the middle of the visible edge (a11y F1) | fixed: the label at the middle of the alert's whole box, from its index (`TestABorrowedAlertAroundTheViewKeepsItsLabel`) |
| An outline walked dot by dot off the screen (code 8, P10-02) | fixed: clipped to the view and its pad (`TestAnOutlineIsWalkedOnlyWhereTheViewIs`); the worst-case alert 38 µs a frame against 26 µs as a line, reading no more of the host's memory (`TestTheWorstCaseAlertReadsNoMore`, `BenchmarkWorstCaseSyntheticAlert`) |
| The render budget claimed re-measured with no Go behind it (code 9) | fixed: the test and benchmark above |
| `Painter.keeping`, `max(5, …)`, `TrackedFrames` returning `any` (code 4) | fixed: deleted, deleted, typed |
| `From` reads as seen; `Heading` from or towards; `Heading` at zero; the 5 km/h number duplicated in a doc; the fit unnormalised; the window fixed (code 1, 5, 6; a11y F10) | fixed in the doc comments and the contract |
| AP-HIST in `blend_test.go` (code 6) | fixed |
| "1 kilometres"; times with no zone; forecast and gap left out of the status (a11y F6, F4, F8) | fixed in the demo and the example (`TestTheMotionWordsSayTheHeadingFirst`, `TestTheLoopIsSaidPlayingOnlyWhileItMoves`) |
| M1's row unamended; the test quoted as evidence; raw headings unpublished; the blizzard loop dropped unexplained; the visual arm quoted (business F1, F6, F7, F4, F11) | fixed: M1's row, the sheet's table of raw headings, the caveats, the blizzard's reason (D-127) |
| No host voices motion; the cut unrecorded (business F2, F3) | fixed: the release notes say so; OW-21 |
| "Held" right after Play (a11y F7) | **HUM, D-124: kept**; the contract says so |
| `TooFewFrames` over a full loop (code 5, a11y F5) | **HUM, D-125: kept**; the contract names every case |
| Speed always in km/h (a11y F3) | **HUM, D-126**: the contract tells hosts to convert |
| A resit on fresh loops (business F5) | **HUM, D-127: no resit**; raw headings published |
| No confidence on a heading; speed unchecked on real loops (business F9) | recorded: stated in the contract; speed is checked on synthetic loops only |
| The no-place case alone checked on real loops (business F10) | fixed: the M1 pin asks the view's centre by name too |
| Surviving mutants: the tie-break, `widestShift`, the pointer clause of `samePictures` (code 7) | recorded: the tie-break decides only between equal positive fits, which real frames do not give; `widestShift` is a bound; `Set` and `Remove` drop the kept frames, so the pointer clause is a defence no path reaches |
| `Frame.steps` grows while a loop is unchanged; a frame that never decodes reads "being read"; the block side by latitude (code 10) | recorded: bounded by the heavier pixels; a decode failure is warned; a recount per latitude, measured within the report's cost |
| `loopWords` copies the example's switch (code 2) | recorded: the demo is a host, and a host words it; each is tested |
| "Off until turned on" names no control (a11y F9) | recorded: playback is the host's to offer; the demo turns it on with `--loop` |
| The mirror test skips with no ledger (code 9) | recorded: it says NOT RUN; the gate runs where the ledger is |
