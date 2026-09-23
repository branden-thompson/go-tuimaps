# round2 business discover — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

**Verdict: proceed to PLAN, with conditions.** Round 1's business and DISCOVER findings are mostly answered in substance, not just marked answered. What's still wrong is a handful of requirement rows that contradict each other, and a risk register that gives ratings without the evidence behind them. Both can be fixed in the docs lane before PLAN opens. **Fix first:** the severe-weather-day blind spot (N-1).

## Job 1 — Do the round-1 answers hold?

**R1-1 (no requirement set, no register, no owed list): mostly holds.** `requirements.md` exists, is declared normative, and traces 78 L-rows to their sources. Residuals:
- **R1-1a** `requirements.md:186`. The register says its ratings are "judged from the evidence cited", but no row cites any: there is no evidence column. It also has no status, owner or trigger column; watchpost's register has a status column. Severity Important. Simplify N. **Action:** add an evidence column (the file:line or measurement behind each L and S rating) and a status column.
- **R1-1b** `requirements.md:15-16,211`. The "OPEN — R1-n" marker is defined and OW-7 tracks it, but no row carries it, so OW-7 tracks nothing. Minor. Simplify Y. **Action:** close OW-7 and delete the marker text.
- **R1-1c** The brief is still stale in places.
  - `project-brief.md:57` says "D-14 to D-22"; it should say D-14 to D-38.
  - `project-brief.md:88` still says "Which to ship is a HUM LEAD ruling"; D-19 has decided it.
  - `project-brief.md:203`: M1 is still scored by sight only. D-24's non-visual arm is missing, and the metrics exist only in the brief, not in the normative file.
  Important, because M1 is the release's primary metric. Simplify N. **Action:** move M1–M6 into `requirements.md` with M1's non-visual arm, and correct lines 57 and 88.
- **R1-1d** OW-2 has two due dates. `requirements.md:206` says it is due "before PLAN commits to L-8.1"; `00-REQUIRED-READING.md:31` says it gates DISCOVER exit. Minor. **Action:** pick one.
- **R1-1e** `requirements.md:120`: L-8.4's cause is "not yet found", but no owed item tracks the diagnosis (OW-3 does track S29-3's). Minor. **Action:** add an OW row.
- **R1-1f** `wave2-measurements.md:11-12` still says the programs were "kept outside the repository", which contradicts D-32 (`rulings.md:49`). Minor. **Action:** correct the line.

**R1-2 (storm motion has no non-visual form): partly holds.** L-1.12 (`requirements.md:56`) exists, but it conflicts with L-13.7 (see N-2). "Heavier rain" is undefined, and M1's non-visual grader is left to PLAN with no bar set.

**R1-15 (scope and value questions): answered by ratified rulings.** D-36 and D-37 carry the human lead's words verbatim, and I do not reopen the no-split decision. Two answers are weaker than their rows say:
- **B-5 (the MRMS heavy end):** `requirements.md:64` makes L-2.3 "required and tested before SHIP", but no test oracle exists for colours nobody has observed. See N-1.
- **P-4 (timeline):** `requirements.md:193` (RK-4) accepts the slip but does not carry what watchpost's own RK-4 says a slip costs (N-5).

## Job 2 — What is new

- **N-1 Severe-weather days, the product's core moment, were never examined** (`requirements.md:64,191`; `wave2-measurements.md:46,50-51`; specimen `README.md:306`).
  - All the MRMS evidence comes from one afternoon at 48 dBZ or less.
  - Specimen 29 has one warning over moderate rain.
  - The blend checker measures one tint at a time. Overlapping warnings over heavy rain, the typical outbreak picture, have no specimen and no requirement.
  - L-2.3's test cannot be written without heavy-end colours, and the server keeps only about two hours of history (`wave2-measurements.md:19`). The evidence therefore waits on live severe weather, a timing dependency the register does not carry.
  - Watchpost's own record calls an outbreak "this product's core moment" (watchpost `requirements.md:154`).

  Severity **Critical**. Simplify N. **Action:** add a register row for the heavy-end evidence depending on the weather, and an OW item for a severe-day capture with a named trigger. Have the human lead rule what L-2.3's oracle is if no severe day arrives, rather than leaving an escape hatch in the notes. Add an OW item for a specimen with overlapping warnings over heavy radar.
- **N-2 L-1.12 contradicts L-13.7** (`requirements.md:56` vs `:173`). L-13.7 defers "intensity in words" and "new place-to-weather statements beyond inside / outside / nearby". L-1.12 requires exactly that: "heavier rain 30 km west … came closer". D-29 (`rulings.md:46`) says D-24 stands, but L-13.7's text does not exempt it, so PLAN will read the two rows as conflicting. L-1.12's "heavier" also rests on the valuation of the heavy end that RK-2 calls untrustworthy. Important. Simplify N. **Action:** add "except L-1.12" to L-13.7, and define "heavier" as a class threshold owned by the table.
- **N-3 The L-11 rows contradict each other.**
  - L-11.2 (`requirements.md:149`) requires the blend to pass "on both grounds".
  - L-11.3 (`:150`) says the light ground "is solved".
  - The specimen's ruled text (`specimens/README.md:329`) says it "must pass on both grounds".
  - L-11.4 (`:151`) names a fallback for when it does not.
  - L-8.8 (`:124`) measures contrast against a "blended inside" on both grounds, which does not exist under the fallback.

  Important. Simplify Y: L-11.3 is a hope, not a requirement. **Action:** reword L-11.2 as "on each ground where the blend is used", fold L-11.3 into L-11.4, make L-8.8 conditional, and correct the specimen README line.
- **N-4 Nothing requires the warning to stay visible inside the blend.** The human lead chose the blend for "conveying the information" (`rulings.md:31`). L-11.2 guards only against washing out (the gap between classes). No row requires a tinted cell to be distinguishable from the same class untinted outside the area, and at the only passing strength (20 %) the blend is "faint" (`specimens/README.md:325`). A blend that passes the checker can still make the tint invisible. Important. Simplify N. **Action:** a ruling-backed requirement for a minimum distance between each class inside and outside the area, or an explicit statement that the outline and label alone carry the warning.
- **N-5 RK-4 misstates the fallback** (`requirements.md:193`). Watchpost's fallback costs more than radar: watchpost's RK-4 says it also loses HR-3 and HR-7. That fallback ships watchpost on v0.1.0, which D-18 calls unreviewed and L-5.6 (`requirements.md:93`) retracts. **Defending this means saying we retracted the version our only host ships on.** Important. **Action:** state both costs in RK-4, and have the human lead confirm L-5.6 against the fallback path.
- **N-6 RK-3 likelihood is under-rated** (`requirements.md:192`). The rating is M, but 0 of 15 light-ground configurations clear the bar (`blend-measure.txt`; S29-7). The evidence says H. Minor. **Action:** re-rate it to H.
- **N-7 RK-10 likelihood is under-rated** (`requirements.md:199`). The rating is L, but D-32 accepts that the programs go stale "and nothing notices", and BUILD rewrites the loop API. Minor. **Action:** re-rate it to H.
- **N-8 RK-5's detector is blind to small shifts** (`requirements.md:194`). If a provider shifts a colour by less than 10 (ΔE), the tolerance fallback matches it to the wrong value silently, and nothing is counted. Important. **Action:** state this as a residual, or count fallback matches (not only unmatched pixels) as drift.
- **N-9 The per-image cap is left unresolved** (`specimens/README.md:319`, S29-4; `requirements.md:45,157`). The fixed cap refused the county picture. L-12.1 makes the total budget settable, but no row says whether a host can raise the per-image cap. Minor. **Action:** one sentence in L-12.
- **N-10 Several rows change existing behaviour under the "additive only" rule** (`requirements.md:179`):
  - L-9.3 widens what `Purge` empties.
  - L-7.3 wraps every host fetcher.
  - L-7.4 makes a fetcher take effect at once.
  - L-11.1 changes what frames look like.

  None was weighed against NFR-1. Minor, since watchpost is the only host. **Action:** one line per row in the record saying why it is not breaking.

## Business Quality

1. **Written or remembered?** Written. Watchpost's HR-1..HR-10 and FR-5.8/5.9 all have L-rows. The exception is M1, whose written definition is the stale one (R1-1c).
2. **What does the user lose if it ships as it stands?** On a severe day, possibly the heavy end of the radar (N-1), and possibly the visible warning tint inside the blend (N-4).
3. **What was cut, and is the cut findable?** L-13.7's deferrals and F-1/F-2 are recorded where a reader will find them. D-37's heavy-end escape clause is findable only in L-2.3's prose.
4. **Is there a cheaper way?** The split was offered and ruled out (D-36). Within that ruling: delete the context rows (L-6.2, L-8.2, L-9.2; L-2.4 duplicates L-2.1) and merge L-1.14 into L-12.4. There is no user-level shortcut.

## DISCOVER lens

1. **Problem framing:** the right problem is "is dangerous weather moving toward a place I care about?", and L-1.12 plus L-11 reach it. The failure is that everything was validated in fair weather (N-1).
2. **Stakeholder coverage:** no real listener or screen-reader user was heard, only personas. The one "embedding developer" is also the author, which the brief now admits (`project-brief.md:33-34`). Minor.
3. **Requirements depth:** "heavier rain" in L-1.12 is undefined (N-2). L-2.3's "a better way than nearest" has no acceptance bar. L-11.3 is an assertion, not a requirement.
4. **Inaction analysis:** stated only for loops (`project-brief.md:41-43`). The cost of doing nothing for the other two thirds of the scope (contract repair, first reviewed release) is not stated. Minor.
5. **Implicit requirements:**
   - The time zone of L-1.10a's frame time is unspecified.
   - Bytes written to the terminal per frame advance (watchpost may run over SSH) are unconsidered.
   - At 69×12, L-1.10a (frame time), the `stale` word, L-8.5 (label fallback), scale and credit compete for one row, and OW-2's specimen does not include frame time (`requirements.md:206`). Layout is the human lead's call; the gap is that the specimen does not test it.
6. **External context:**
   - The two-hour server history window and severe-weather seasonality are missing (N-1).
   - IEM steps every 5 minutes and MRMS every 2, so "12 frames" covers different spans; L-1.12's "span" handles this.
   - No provider terms apply to the library, since L-1.4 keeps fetching with the host.
7. **Risk assessment quality:** the ratings have no evidence column (R1-1a), and RK-3 and RK-10 are mis-rated (N-6, N-7). Missing categories: total scope and effort (58 rows marked "NO INSTRUMENT YET", no estimate, no date, and no trigger to re-check D-36 with numbers at PLAN exit); severe-day evidence depending on the weather (N-1); and fallback versus retraction (N-5).

**Standing rules:** no reason here was issued by the team itself. Everything I cite is a ratified ruling, and I treated D-14, D-36 and D-37 as final.
