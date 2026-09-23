# round1 a11y — reviewer report, verbatim

Filed under v0.2.0 D-38/round 2 (H-9): the reviewer's final message exactly as returned, except that machine paths are replaced by `<workspace>` and `<scratch>`. The harness refused reviewers' own report files, so this is the only copy.

---

# go-tuiMaps v0.2.0 "Radar loops": accessibility red team at the DISCOVER exit

Reviewer: Staff Accessibility Advocate. Read-only: I only ran `cat` and `grep` over the tree and the specimens. `report.md` was not written: the harness refuses report files from a subagent, so this message is the whole report.

## 1. Is the output semantic?

**A1 (A11y) — Motion, the point of the loop, has no text alternative.**
- The locked problem is that the reader "cannot see a storm's motion" (`01-objectives/project-brief.md:42-44`).
- M1 is scored by sight alone (`project-brief.md:189`).
- Wave 1 says `Describe` should answer for the newest frame only (`02-analysis/wave1-findings.md:72-74`).
- The image answer gives a class and a bearing to heavier rain. It says nothing about change over time (`internal/describe/answer.go:145-153`).

A listener still learns where the rain is, never where it is going.
- Severity: **Critical**. Simplify/Delete? N.
- Action: add an L-1 requirement that the description reports motion from the frames' data: the heavier class's distance and bearing then and now, approaching or receding, and the time span covered. Give M1 a non-visual arm scored from the description. The wording is the HUM LEAD's.

**A2 (A11y) — Nothing requires a loop frame to show its valid time, or a gap, as text.** L-1.2 and L-1.3 say "stated" but not where.
- Severity: Important. Evidence: `project-brief.md:58-61`; the existing `stale` word shows the pattern at `internal/render/frame.go:622-624`. Simplify/Delete? N.
- Action: require the frame time, or "gap" in its place, as a word on the frame at every depth. Where it goes is the HUM LEAD's.

## 2. Can a host expose playback to a keyboard-only user?

**A3 (A11y) — There is no way to step through frames.**
- L-1.5 offers off / slow / normal only (`project-brief.md:64-70`). The proposed API is a setter plus a read-only `Shown` (`wave1-findings.md:50-53`).
- A keyboard user needs previous, next and newest frame.
- `Animate` could fake stepping (`clock.go:9-27`), but no one has said which clock drives the loop.
- Severity: Important. Simplify/Delete? N.
- Action: require step and seek-to-newest, and state which clock drives the loop.

**A4 (A11y) — `ReduceMotion` versus per-overlay playback is not ruled.** "ReduceMotion implies Off" exists only as a wave 1 proposal. Today `ReduceMotion` covers markers only.
- Severity: Important. Evidence: `wave1-findings.md:56`; `look.go:204-221`; `internal/render/motion.go:26-36`. Simplify/Delete? N.
- Action: make it a requirement. Also state what happens to playback when reduce motion is switched off again.

Otherwise the library does not get in the way: it owns no keys, and its setters do not block.

## 3. Can a host tell a listener the name, role and state?

**A5 (A11y) — Reading the loop's state is not a requirement.**
- `Shown` is only a proposal.
- It also lacks the frame count ("7 of 12") and the playback state actually in effect, including "off because reduce motion is on".
- Severity: Important. Evidence: `wave1-findings.md:50-56`; the brief's L-1 has no such item (`project-brief.md:54-70`). Simplify/Delete? N.
- Action: require a state read giving the playback setting in effect, the frame index, the count, the valid time and any gap.

**A6 (A11y) — A frame advance will probably look like a data change.**
- `Changed()` is the host's only redraw signal (`look.go:284-293`).
- The description cache is keyed on the overlays version (`describe.go:115-117`). The loop has to beat frame reuse (`internal/render/frame.go:285-302`), and bumping that version is the obvious way to do it.
- A screen-reader host then either announces every tick or misses real changes (WCAG 4.1.3).
- Severity: Important. Simplify/Delete? N.
- Action: require a frame advance to be distinguishable from a data change, and a tick not to change the description's cache key.

## 4. Contrast, and meaning beyond colour

**A7 (A11y) — The specimen 29 defect is wider than recorded, and no brief requirement carries it.**
- The record scopes it to "no colour" (`specimens/README.md:317`). Here is the mechanism.
- At `rampless` depths an image cell becomes taken text (`internal/render/field.go:200-204`). `rampless` is NoColour **and Colours16** (`field.go:103-105`).
- Everything written afterwards goes through `write`, which refuses taken cells (`frame.go:176-179`, order at `frame.go:493-509`).
- So at 16 colours as well, radar erases the outline, the warning label and the "you are here" marker.
- It also erases the **`stale` word** (`frame.go:599-602`) and the no-tiles notice (`frame.go:585-598`), which neither S29-2 nor D-14 lists. A stale loop can therefore look current, against L-1.3.
- D-14 and D-17 say the defect "becomes a v0.2.0 requirement" (`rulings.md:31,34`). Yet the brief has no D-14, D-17 or specimen 29 requirement (`project-brief.md:49-143`).
- Severity: **Critical**. Simplify/Delete? N.
- Action: add a brief requirement that outline, label, marker, stale word, frame time, notice, scale and credit survive an image or field at every depth. Test it at NoColour and at Colours16.

**A8 (A11y) — D-17 can fail silently.**
- (a) A label that does not fit is dropped whole, not cut short (`frame.go:171-180,199-203`). At 69×12 the label carrying the severity word can vanish, and nobody is told.
- (b) The five outline dashes must also differ from the existing line-overlay dash, 5 dots on and 4 off (`internal/render/hatch.go:12-19`).
- (c) The hatch runs only at NoColour (`frame.go:513,555`), so there is no second cue at 16 colours.
- Severity: Important. Simplify/Delete? N.
- Action: require a fallback for a dropped label (a short severity word, or a report to the host), require dashes distinct from the D-77 line dash, and put all three in the owed specimen.

**A9 (A11y) — D-14's checker measures only how far apart the radar classes are, and only on truecolor.**
- The ruled measure is class-to-class distance of at least 10 (`rulings.md:31`). It does not require the outline at 3:1 (WCAG 1.4.11) or the label at 4.5:1 against the blended inside and the plain radar outside.
- `Foreground` meets contrast by falling back to black or white (`internal/colour/contrast.go:40-50`). That is safe for AA, but over rain it strips the outline's severity colour.
- Nothing checks the blend after it is rounded to 256 colours (`frame.go:675-682`). `Warnings` checks nothing at 16 colours (`internal/colour/palette.go:122-125`).
- Severity: Important. Simplify/Delete? N.
- Action: extend the D-14 check to outline and label contrast over blended and plain radar, on both grounds, at Truecolor and 256.

**A10 (A11y) — The light ground has no fallback.** Every blend strength failed on it (`specimens/README.md:326`). D-14 rules the problem "solvable" without naming a fallback (`rulings.md:31`).
- Severity: Important. Simplify/Delete? N.
- Action: this is a fork for the HUM LEAD, not a ruling from me. Either a per-class tint found by search, or a named fallback such as 29b's order (radar over tint, outline and label kept). The requirement should say what ships if the search finds nothing.

## 5. Text alternatives

**A11 (A11y) — An area answer does not say which alert it is, or its severity.**
- `OfArea` sets no label and no severity (`internal/describe/answer.go:98-111`; `describe.go:219-221`). Severity is known from `Role` (`internal/overlay/store.go:55`).
- D-17 puts the severity word in the picture only.
- Severity: Important. Simplify/Delete? N.
- Action: carry the area's label and severity word in the answer.

**A12 (A11y) — An image answer is a bare class number.** `Value` and `ValueUnit` are left unset (`answer.go:67-74,145-153`), though the colour table carries values (`internal/overlay/image.go:42-46`). Nothing marks the D-19 MRMS table as approximate.
- Severity: Minor. Simplify/Delete? N.
- Action: return the value, its unit, and an "approximate" flag.

**A13 (A11y) — With no places registered, `Describe` returns nothing, and there is no summary of what is in view.** Watchpost deferred this as its F-8 (`watchpost/.../08-reports/red-team-discover.md:163`).
- Severity: Important. Evidence: `describe.go:77-79`. Simplify/Delete? N.
- Action: rule on it in this release rather than inherit the deferral.

## 6. Motion, auto-play, timing

**A14 (A11y) — The default playback state is not stated.** WCAG 2.2.2 requires that auto-playing motion longer than five seconds can be paused. L-1.5 supplies Off but no default.
- Severity: Important. Evidence: `project-brief.md:64-66`. Simplify/Delete? N.
- Action: a fork for the HUM LEAD. Defaulting to normal means every host must wire Off; defaulting to Off means motion is opted into.

**A15 (A11y) — "Off" has no defined still picture.** Watchpost ruled its still form keeps the newest frame and its age (`red-team-discover.md:53`). The library's requirement does not carry that, so Off could freeze on an old frame that looks current.
- Severity: Important. Evidence: `project-brief.md:64-70`. Simplify/Delete? N.
- Action: carry watchpost's still form into L-1.5.

**A16 (A11y) — The rate ceiling has no number and nothing to count against.**
- The only flash rule applies to marker blink (`internal/render/motion.go:5-9`, D-56 at 2.5 a second).
- A loop adds large-area brightness changes. Gap frames drawn empty would flash the whole radar area off and on, and the wrap from last frame to first is another big change. None of this is covered.
- Severity: Important. Evidence: `project-brief.md:68-70`. Simplify/Delete? N.
- Action: require a numeric ceiling that counts every frame change toward WCAG 2.3.1, and require that gaps are never flashed as empty frames. Whether to pause on the last frame before looping is the HUM LEAD's.

**A17 (A11y) — `ReduceMotion`'s comment overstates what it does** ("nothing is ever due"). It is about to become contract text that hosts rely on.
- Severity: Minor. Evidence: `look.go:204-205`; already on wave 1's fix list (`wave1-findings.md:197`). Simplify/Delete? N.
- Action: correct it when the loop's behaviour is written.

Clean: the library has no hard timeouts.

## Opinion

**Do not proceed to PLAN yet.** What is needed is changes to the brief, not rework:
- make D-14, D-17 and the specimen 29 defect, at its full scope (A7), into brief requirements;
- add a description of motion (A1) and a readable loop state (A5) to L-1;
- get rulings on the A10 and A14 forks.

Then proceed.

**Fix first: A1.** A loop whose motion exists only as a picture gives non-visual users the very failure the locked problem statement names.

All paths above are relative to `<workspace>/go-tuimaps/` (record files under `06_docs/02_features/radar-loops/`; specimens under `06_docs/02_features/go-tuimaps/02-analysis/specimens/`), except the watchpost record at `<workspace>/watchpost/06_docs/02_features/observer-maps/08-reports/red-team-discover.md`.
