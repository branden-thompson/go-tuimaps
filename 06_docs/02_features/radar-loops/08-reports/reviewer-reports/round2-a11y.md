# round2 a11y — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

## Round 2 accessibility review: go-tuiMaps v0.2.0 "Radar loops", DISCOVER exit

**Verdict: do not proceed to PLAN yet.** Three contradictions in the normative text (F1, F2, F3) each need a short ruling. If they go to PLAN unruled, PLAN will settle them silently, and each one decides what a screen-reader user hears. The other findings can be carried into PLAN as owed items.

**Fix first: F1**, which frame the description and the `stale` word follow.

### Do the rulings close A1–A17?
- **Closed as written:** A2, A4, A5, A6, A10, A14, A15, A16 (A16 has a scope gap, F6).
- **Partly closed:**
  - A1 (F1, F2, F3)
  - A3 (stepping gives a listener nothing, F1)
  - A7 (the hatch never draws in rain, F7)
  - A8 (F7)
  - A9 (F8)
  - A11 (F4, F5)
- **Deferred:** A12 and A13 by ratified D-29. See F12.
- **A17 is missing.** It is in no disposition (F11).

---

### 1. Is the output semantic?
Only through `Describe`. The picture is not semantic: blank cells are U+2800 (for example `29b-…-69x12.txt:2`), so a terminal screen reader reads "braille pattern blank" over and over.

**F10: A11y. Minor.** Two wording rules bind nobody who produces words.
- Evidence: an `Answer` is "data … never a sentence of the library's own" (`internal/describe/answer.go:52-53`, D-52). Yet L-1.12 requires wording "as observation, never forecast", and L-13.6 requires "never says 'you'" (`requirements.md:56,172`). The library says nothing, so both rules pass trivially inside it. The words are made in watchpost, and no requirement follows them there.
- Simplify/Delete? N
- Action: restate both rules as data shape. For example: motion returns two timed positions and a relation, and has no field that can express a forecast. Test the phrasing where it is actually made (the example host, or watchpost M1b).

### 2. Can a host give loop playback to a keyboard-only user?
Yes on paper (L-1.10b, L-1.13), with one exception.

**F9: A11y. Important.** The state read can report "playing" while the loop is frozen.
- Evidence: L-1.10b says "the library's animation clock drives the loop" (`requirements.md:48`). But `Animate(at)` freezes that clock (`clock.go:13-27`), and `NextCall` drops animation wake-ups while the host drives it (`clock.go:94`). A host that freezes animation therefore stops the loop, while L-1.10d's state read (`requirements.md:50`) can still say "normal".
- Simplify/Delete? N
- Action: L-1.10d reports "playing" only while the shown frame is actually advancing. Separately, state what `Animate` does to a loop.

### 3. Can a host tell a listener the loop's name, role and state?
State: mostly. Name: no.

**F13: A11y. Minor.** No name, no span, and no rule for which reason for "off" wins.
- Evidence:
  - An overlay has an `ID`, but no human name (`internal/overlay/store.go:61-69`).
  - The state read (`requirements.md:50`) has no span of time (oldest to newest).
  - Three reasons for "off" can apply at once: "off (default)" (L-1.11), "off because reduce motion is on" (L-1.10d) and the host's own choice. Nothing says which one the state read reports.
- Simplify/Delete? N
- Action: add the span and a precedence order to L-1.10d. The name can stay the host's job, but say so in the contract.

### 4. Contrast, and meaning beyond colour
**F7: A11y. Important.** The D-17 specimen is not required to be drawn over radar, which is the one place the dashes are likely to fail.
- Evidence:
  - OW-2 (`requirements.md:206`) asks for 69×12 and 149×38 at NoColour and Colours16, but not over rain.
  - Without colour, a rain cell becomes one shade glyph that takes the whole cell (`internal/render/field.go:196-201`, `taken: true`). A cell holds one character, so the gaps in a dash fill with `░▒▓`, and five dash patterns read as texture.
  - The hatch draws only on blank cells (`internal/render/hatch.go:51`). L-8.7's hatch at 16 colours therefore adds nothing inside rain, which is where the warning matters most.
- Simplify/Delete? N
- Action: OW-2 is drawn over specimen 29's radar at both depths. L-8.7 states what carries severity inside rain cells.

**F8: A11y. Minor.** L-8.8's 16-colour arm can only pass trivially.
- Evidence:
  - Alert outline tokens have no 16-colour entry (`internal/colour/sixteen.go:34-36`).
  - Rain is drawn without a ramp at 16 colours (`field.go:103-105`), so there is no blend to check.
  - The reference table is "indicative only" (`sixteen.go:3-6`).
- Simplify/Delete? **Y**
- Action: delete the 16-colour arm. State plainly that severity at 16 colours rests on the word and the dash.

The blend itself (L-11.2, measured in S29-7) and the light-ground fallback (L-11.4) are sound.

### 5. Text alternatives: does the description say what the picture shows?
**F1: A11y. Critical.** The requirements do not say which frame drives staleness and the description. The three rows involved conflict.
- Evidence:
  - L-1.3 says each frame's own valid time drives staleness and the description (`requirements.md:38`).
  - The description's cache key includes staleness (`describe.go:110,116`), and the cache is dropped whenever staleness changes (`describe.go:282-286`).
  - If the *shown* frame drives it, the `stale` word switches on and off every loop cycle. That is a false alarm, a blinking element, and a cache change on every cycle, which breaks L-1.10e (no re-announcing on each tick).
  - If the *newest* frame drives it, L-1.3 is wrong as worded.
  - Either way, a listener who steps to frame 3 (L-1.10b) learns only a time. `imageAnswer` reads the overlay's single raster (`describe.go:245-256`), and the state read carries no content.
- Simplify/Delete? N
- Action (a fork for the human lead):
  - (A) The description and the `stale` word follow the newest frame, and stepping is ratified as sight-only.
  - (B) As A, plus a per-frame answer for the frame the user has stepped to.

**F2: A11y. Important.** Motion is only described relative to a place. With no place, the listener gets nothing, and the core term is undefined.
- Evidence:
  - With no places, `Describe` returns `nil, nil` (`describe.go:73-79`). A host that registers no place gives a screen-reader user no loop information at all, and L-1.12 is still met.
  - "Heavier rain" is not defined. Today "heavier" means heavier *than the class at the place* (`describe.go:255`, `answer.go:70-71`). That reference point moves as the place's own rain changes between frames, so "30 km west, then 18 km west" can compare two different things.
- Simplify/Delete? N
- Action: fix one intensity threshold for L-1.12 across all frames. Then either require motion without a place, or have the human lead ratify the exclusion of place-less hosts by name.

**F3: A11y. Important.** L-13.7 defers exactly the kind of statement L-1.12 requires.
- Evidence: `requirements.md:173` defers "new place-to-weather statements beyond inside / outside / nearby". L-1.12's motion statement is one of those. The ruling that keeps it, "D-24 stands as ruled" (`rulings.md:46`), is missing from the normative file, and the normative file wins. v0.1.0's `Heavier`/`HeavierAt` fields (`answer.go:70-71`) also already go beyond inside/outside/nearby.
- Simplify/Delete? **Y**
- Action: L-13.7 adds "except L-1.12 and v0.1.0's existing answers".

**F4: A11y. Important.** L-13.6's premise is false, and A11 stays open.
- Evidence:
  - v0.1.0 has only `Outside` and `Inside` (`internal/describe/area.go:21-22`). There is no "nearby", so it is not something "v0.1.0 does", and it has no threshold.
  - Area answers carry no `Label`: `OfArea` never sets one (`answer.go:95-110`).
  - There is one answer per overlay, merged across every feature in it (`describe.go:199-220`).
  - So D-29's own example, "Hamlin: inside the Flash Flood Warning", cannot be produced. With two warnings in one overlay, a listener cannot tell which one the place is inside.
- Simplify/Delete? N
- Action: the area answer names the containing alert or alerts, with severity words, and "nearby" gets a threshold or is deleted.

**F5: A11y. Important.** L-13.5 cannot be built on today's data or API shape.
- Evidence:
  - Valid time is kept per overlay, not per alert (`store.go:62-64`).
  - Severity exists only as a colour token in `Role` (`store.go:55`). An alert drawn in a host-chosen role gets no severity word.
  - `Description` is per place (`describe.go:19-28`), so "independent of any place" has nowhere to go.
- Simplify/Delete? N
- Action: L-13.5 names a place-independent surface and a per-alert valid time. Severity comes from data, not from an ink.

**F12: A11y. Important (a fork, not a finding against the ruling).** The deferral of intensity words leaves the text alternative short of what the picture shows (WCAG 1.1.1).
- Evidence: an image answer gives `Class int` (`answer.go:69`) with no word or unit. A sighted user reads intensity from colour; a listener hears "class 4". D-29 deferred this as necessity unproven, and that deferral is ratified.
- Simplify/Delete? N
- Action: let M1's non-visual arm decide. If its grader asks "how heavy?", necessity is proven. The human lead decides the grader.

### 6. Motion, auto-play and timing
Default-off (L-1.11), `ReduceMotion` (L-1.10c) and no flashed gaps (L-1.10g) are sound. **L-1.2 against L-1.10g** is reconciled by D-25(7). Residual: a low-vision user who cannot read the corner text is looking at a neighbour frame's picture.

**F6: A11y. Important.** The flash ceiling has no stated scope.
- Evidence:
  - L-1.10g (`requirements.md:53`) does not say whether the ceiling is per map. L-1.13 gives each overlay its own controls, so two loops out of step could each run at 2.5 changes a second and together double it.
  - It does not say whether blinking markers count toward it. They make 2 changes a second (`internal/render/motion.go:5-9`).
  - It does not cover F1's toggling `stale` word.
  - L-1.10a does not say which loop's time is shown when two loops play.
- Simplify/Delete? N
- Action: one ceiling per map, counting every moment anything changes (frames, blink, furniture), and a rule for which loop's time is shown.

**L-1.11 against L-1.13: no conflict in the library, but a gap in the measure (Minor).**
- Evidence: default-off protects only hosts that never call the API. A Settings row that defaults to "normal" brings auto-play back (L-1.13 measure, `requirements.md:55`).
- Simplify/Delete? N
- Action: L-1.13's watchpost measure asserts that watchpost's Settings default is off.

### New
**F11: A11y. Minor.** A17 is missing from the record.
- Evidence: the consolidated round 1 table (`08-reports/red-team-discover.md:49-66`) cites A1–A16 only. A17 appears nowhere under `06_docs`.
- Simplify/Delete? N
- Action: find A17 and give it a disposition, or record it as lost.

I worked read-only: I edited nothing in the tree, and ran no gate, fuzzing or tests. The scratch directory `<scratch>/rt2-a11y` was created and is empty.
