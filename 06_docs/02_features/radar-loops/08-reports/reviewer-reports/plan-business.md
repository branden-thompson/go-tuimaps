# Reviewer report — PLAN red team, business

Filed verbatim; machine paths redacted.

**PLAN-exit red team (business and architecture): go-tuiMaps v0.2.0 and watchpost 0.18.0**

Abbreviations used in citations:
- **LP** = `<projects>/go-tuimaps/06_docs/02_features/radar-loops/04-development/implementation-plan.md`
- **WP** = `<projects>/watchpost/06_docs/02_features/observer-maps/04-development/implementation-plan.md`
- **LR / WR** = each feature's `01-objectives/requirements.md`
- **LD / WD** = each feature's `02-analysis/rulings.md`
- **PS** = watchpost `01-objectives/problem-statement.md`
- **W1F** = watchpost `02-analysis/wave1-findings.md`
- **W2M** = go-tuimaps `02-analysis/wave2-measurements.md`
- **IM** = `integration-map.md`

I worked read-only and ran no gate. My scratch directory is empty.

## 1. Written requirements or remembered ones?

Mostly the written ones. I traced these end to end and they hold: D-59→L3.2 (LP:155), D-60→L3.13 (LP:166), D-55→L8.1/L8.3 (LP:207-209), D-56→L9.1/L9.2, D-44(4)→L6.7 (LP:191), D-26→L4.1/L4.11. On watchpost: D-41's four guards→W2.1-W2.5/W9.6, D-42→W5.4 (Fort Davis/TXZ277, WP:162), FR-5.7→W8.5 (including the decode cap), FR-3.2→W3.2, and D-43's numbers→W1.14/W5.6/W8.9/W2.9.

Where the plans depart from the record:

- **W3.9 makes zone fetches one at a time, which no requirement asks for.** NFR-4 says "sequential **or bounded**" (WR:142). The code is already bounded at `fetchAtOnce = 6`, with a politeness rationale (`watchpost/domains/weather/nws/zones/zones.go:37`). W3.9 tests "at most one zone request in flight" (WP:144). This is a remembered rule, and it breaks M5 (see question 10). Severity **Critical**. Evidence: WP:144, WR:142, W1F:107,112. Simplify/Delete? Y. Action: delete W3.9's "one in flight" and keep the existing bound of 6, with a test that it holds.
- **L-5.4 has no task.** L-5.4 says v0.2.0's REVIEW covers everything v0.1.0 shipped (LR:95). It is missing from the trace (LP:250-285). This is the largest single piece of work between here and the tag, and watchpost's P1-b waits on that tag. Severity Important. Evidence: LR:95, LP:250-285. Simplify/Delete? N. Action: add a task in L10, with its own scope statement.
- **Watchpost M6 is narrowed without a ruling.** PS:104 defines M6 as frame-interval jitter **and key-to-response latency**. W8.13 measures only "drift within one frame a minute" (WP:198), and that number is not in D-43. Details under question 9.

## 2. What does the listener lose?

**If both releases ship as planned:**
- A quiet season means MRMS ships with its heavy end unverified (accepted in D-44).
- The basemap still has one source.
- Motion is described only on P1-b.

The loss nobody recorded:

- **Watchpost never shows the MRMS "unverified" or `table-fallback` signal to the listener.** The library carries it in a legend entry and a warning (LP:188-191). Watchpost has no legend UI in phase 1, and no WP task surfaces library warnings beyond `Sharpening`/`tile-failed` (WP:138). MRMS can be picked in Settings (WP:188). So on a severe day the listener sees fallback-valued reds and is not told. Severity **Important**. Evidence: LP:191, WP:138, WP:188. Simplify/Delete? N. Action: add a W8 task that puts the fallback share or the "unverified" state on the notes line (W5.4 already has one), or keep IEM as the default source while `Unverified` is set.

**If watchpost ships P1-a alone (the fallback):**
RK-4 says this costs radar, HR-3 and HR-7 (WR:153). It also costs four things RK-4 does not list:
- The named place's label is lost at 69×12 and in partial areas (S29-3, S31-2). That is exactly M1's core case, and it is only fixed in W9.8.
- The description comes from v0.1.0's `Describe`, which has no "nearby" and no severity or alert name per area.
- Tiles go out under the library's `0.1.0-dev` user-agent (L-13.1).
- Retention is a byte cap, not 7 days (WP:256 records this last one).

Severity Important. Evidence: WR:153, WP:213, WLD D-42. Simplify/Delete? N. Action: expand RK-4's cost line to the full list, so the SHIP report cannot understate it.

## 3. Cuts and deferrals: recorded where the next reader will look?

Mostly yes. The "Deviations" sections are good (LP:287-291, WP:250-257). Four are unrecorded:
- **RK-5's basemap fallback** is labelled "PLAN's first network question" (WR:154), but no WP task or deviation mentions it. Severity Important. Action: one ruling, even if it is "accept single source; the FR-3.4 notice suffices".
- **What happens if OW-2's specimen fails.** If five dashes can't be told apart at braille resolution, L3.9-L3.11 have no fallback (LP:162-164, IM:76). HR-7 and W9.9 depend on them. Severity Important. Action: state now whether v0.2.0 would ship with the word only, or wait for a new ruling.
- The narrowing of M6 (question 9).
- W3.9's change to concurrency (question 1).

## 4. Is there a cheaper way to the same listener outcome?

Only small savings. The locked problem is alerts; radar is "context, not subject" (PS:83). P1-a is therefore the real deliverable, and RK-4 already ships it without waiting. The throwaway work (the host clamp W4.4, `Describe`-based W1.4, the directory delete W3.8, guard-4 partial W2.5) is the price of that hedge. With no date on v0.2.0 (D-37), I think it is worth paying. Cheap trims:

- **`PurgeWithReport` beside `Purge`:** fold them into one call (LP:222). Severity Minor. Simplify/Delete? Y.
- **M4's five-minute leg in every full gate** (LP:237): the gate already takes 18-21 minutes (`go-tuimaps/06_docs/gate-runs.md:16-23`) and it runs on every task. Keep the soak before SHIP plus a short gated form, sized in seconds rather than minutes. Severity Minor. Simplify/Delete? Y.

## 5. Optimisation target

- **Library: quality, and deliberately so.** D-36 says "it needs to work and be correct", and D-37 sets no date.
- **Watchpost: schedule flexibility**, through the P1-a/P1-b hedge (D-11, RK-4). Also a decision.
- **Speed was a default casualty in both.** About 75 library tasks each run the full gate (~20 minutes) plus a mutation check (LP:30-35), and nobody sized that. I think it is acceptable given the owner's quality rulings, but it should be said out loud once in the PLAN report. Severity Minor.

## 6. Incentive alignment

- **The two orders don't conflict inside the library.** Watchpost is gated on the tag (FR-5.6), so it gains nothing from the L2→L4→L3→L5 order.
- **The cost falls one way.** The library's no-split rule (D-36) makes every watchpost P1-b item wait for the slowest library item: the v0.1.0 REVIEW (L-5.4, unplanned), the soak, M1 grading and hosted CI. It was ruled and accepted, but the biggest term in it (L-5.4) is invisible.
- **A number that crosses the repos is owned by nobody.** Neither plan nor the IM states how many frames, and at what size, watchpost hands in. See MaxFrames under question 10. Severity Important. Evidence: IM:64-82 (no row for it), WP:191, LP:46. Simplify/Delete? N. Action: add an IM row for frames per source, frame size (the dot grid) and budget, owned by the host.

## 7. Failure propagation if v0.2.0 slips

This is stated: P1-a ships, and the HUM LEAD decides at REVIEW (WP:85-86, IM:60-62, WR:153). Two things are not stated:
- What ships if v0.2.0 slips *partially*, for example with OW-2 failing (question 3).
- That RK-4's cost list is incomplete (question 2).

## 8. Scope calibration

**Built for a problem these releases do not have:** nothing significant. The provider seam (L6.1) and the antimeridian box (L7.1) were ruled and are needed.

**Specified too thinly to build without guessing:**

- **L5.5 motion.** Nothing says how "the heavier rain" is located (nearest pixel above the threshold? a blob? which cell when there are several?), how `Threshold` is chosen "from the table", or what tolerance counts as "held" (LP:177; approach-4 lines 43-56). This is RK-8's whole mitigation and the substance of M1's non-visual arm, and D-42 itself warned that "two frames can compare different storms". The only test is one cell in two frames. Severity **Important**. Simplify/Delete? N. Action: specify the sighting rule and add a two-cell fixture plus a fixture with a gap at the oldest frame.
- **L6.9's fallback-share test has no number.** "Exceeds the threshold" (LP:193) is not among the numbers this plan proposes. Severity Important. Action: add it to the numbers table.
- **W1.4 builds "the alerts shown" from `Describe(places)`** (WP:106). v0.1.0's `Describe` needs a place, merges every feature of an overlay (LD D-43), and has no list of alerts shown (`go-tuimaps/internal/describe/answer.go:53-81`). The plan also never says whether each alert is its own overlay. Severity Important. Action: state one overlay per alert, or that the list comes from watchpost's own alert data.
- **L3.2's `Frame.Dropped []Drop`, and L3.13's `DropPlaceName`, depend on the Drop type in the pending L3.11** (LP:155,164,166). The dependency list omits this (LP:243-246). Severity Minor. Action: move `Drop` into L3.2.

## 9. Metrics: measurable, by whom, when, gameable?

**Library:**
- M2-M5 are automated and measurable.
- **M1:**
  - The visual arm has no pass mark. "Scored as M1's visual arm is" (LP:48) refers to a scoring rule that does not exist (LP:238, brief:203).
  - The non-visual arm is scored by the HUM LEAD, who assembled the five loops (L10.9). The plan does not carry watchpost's own lesson to score the non-visual arm FIRST (PS:138).
  - The ground truth for "direction" is undefined.
  - The arm reads `From`/`To` data, so it tests arithmetic rather than whether the right cell was picked.
  - Severity **Important**. Evidence: LP:48, LP:238, PS:138. Simplify/Delete? N. Action: define ground truth (for example, the visual-arm answer, or a track worked out by hand), score the non-visual arm first, and include a scene with several cells.
- **M6:** "unattributable" is attributed by the builder. Severity Minor.

**Watchpost:**
- **M1b:** W7.2 human-scores the P1-a description built on `Describe`. W9.2 then swaps the source to `Report` and calls M1b "green on the new source" (WP:207). That quietly turns a human-graded metric into an automated one, so the description that actually ships is never scored. Severity **Important**. Action: re-score M1b by hand on P1-b (the non-visual arm first), or score only the description that ships.
- **M5:**
  - It is measured over recorded fixtures through a "latency-shaped transport" whose shape is unspecified (WP:130). PS:102 defines cold as "network reachable".
  - The test author's latency profile decides pass or fail, so it is gameable.
  - Severity Important. Action: set the shape from W1F's measured latencies (0.16-0.36 s per zone, 260 ms for TileJSON, 50-100 ms per tile), and add one live-network sample to the SHIP report.
- **M6 (W8.13):** frame choice is a pure function of animation time (LP:139), so "schedule drift" cannot fail under load. Starvation shows up as skipped frames and lagging keys, which it does not measure. "One frame a minute" was never ruled. Severity **Important**. Evidence: WP:198, PS:104,115, LP:139. Simplify/Delete? N. Action: measure the intervals between delivered `mapTickMsg`s and key-to-response latency with the Observer live, and take the thresholds to the HUM LEAD.

## 10. Are the numbers justified by evidence?

- **MaxFrames 36** is justified as "three hours at a five-minute cadence" (LP:46), which is IEM's cadence. MRMS runs about every 120 s, irregularly (W1F:22; its time list holds 60 entries over 2 h, W2M:20). FR-3.9 keeps 2 hours of radar (WR:70; W8.14 at WP:199). So a full MRMS loop is 60 frames and the library would refuse it. The budget arithmetic (36 × 0.59/12 ≈ 1.8 MiB) is right, but the count is fitted to one source. Severity **Important**. Evidence: LP:46, W1F:22, WR:70. Simplify/Delete? N. Action: choose frames per loop per source (see the IM row in question 6), then set MaxFrames to cover it, or have watchpost subsample MRMS and say so.
- **File cap of 4 B/px + 64 KiB.** It is sound as an anti-padding bound. Real frames are 5.6-21 KB (W2M:64-65), so it sits 10-50× above them. Polish.
- **M1 non-visual pass mark "4 of 5 within one compass point".** There is no evidence behind it, and "compass point" doesn't say 8 or 16 points. Severity Minor.
- **M6 at five runs** (LP:49). A flake that fires one run in three still passes 13% of the time; a one-in-ten flake passes 59% of the time. The OW-4 rate is unknown. It is cheap, so keep it, but state the detection power it actually gives. Severity Minor.
- **The reference machine** is "one 18-core Apple M-series Mac" (W2M:93). Nobody else can reproduce it, and the 15 ms time is recorded but not gated (LP:167). The 15 ms figure rests on no per-advance measurement: the 27 ms was twelve frames composited together (W2M:100-101). With 500 ms per frame there is ample headroom. Polish.
- **Watchpost M5 at 2.0 s is not supported.** The measured cold 41-zone resolve is "~2 s" at 6 in parallel (W1F:112), and that is before TileJSON (260 ms) and tiles. W3.2 moves zone seeding onto the first `g`, and W3.9 serialises it: 41 × 0.16-0.36 s ≈ 7-15 s. The target cannot be met as planned, and only a generous fixture latency would hide that. Severity **Critical**. Evidence: WP:130,137,144, W1F:107,112, WD D-43. Simplify/Delete? Y. Action: restore the bound of 6, and either re-derive M5 or exclude zones outside the view from the first complete frame. Rule on it.
- **The 2 MB / 40-request warning and `NewShared(4 MiB)`** trace to W1F:27,84. Justified.

## Verdict

- **go-tuiMaps v0.2.0: approve for BUILD with conditions.** Before L5.5, specify the motion sighting rule. Before L6.9, set the fallback-share threshold. Before L2.2, reconcile MaxFrames with MRMS. Add L-5.4 as a task. The rest is sound and traced.
- **watchpost 0.18.0: do not approve yet.** Fix the M5/W3.9 conflict and the M6 instrument, carry M1b onto the description that ships, surface MRMS's unverified state to the listener, and complete RK-4's cost list. Each is small, and none needs a redesign.

**Fix first:** W3.9's single-flight zone fetch. It contradicts the existing bound (`zones.go:37`), is not required by NFR-4, and makes watchpost's primary M5 target impossible to meet honestly on the cold path W3.2 creates.
