# How to run the acceptance sitting

Up: [readiness](.) · Plan tasks 14.15 to 14.18 · Carries: M1 (D-43, D-67), NFR-15, FR-5, D-52, D-57

| Field | Value |
|---|---|
| Who runs it | HUM LEAD. Nothing here is the coordinator's to judge |
| What it decides | **M1a** — whether the picture alone answers each scenario's question (14.15); whether the reference frames become goldens (14.16); whether the description is speakable (14.17); and NFR-15's no-colour reading (14.18) |
| Where the verdicts go | [`uat-record.md`](uat-record.md), which is blank and is filled in during the sitting |
| Before anything else | **The map is drawn with braille.** A terminal whose font has no braille shows boxes where the map should be, and the terminal cannot be asked what its font holds (D-57). Check yours with the first command below before judging anything |

## 0 · Build it, once

```
scripts/uat
```

That builds `./dist/tuimaps`. No remote exists until SHIP, so the script writes a
throw-away workspace file **outside** the repository and builds through it; the tree is
left as it was found.

Check the font before anything else:

```
./dist/tuimaps --offline --scenario 1 --size 69x12 --headless --no-colour
```

If that is a picture, the font is fine. If it is a grid of empty boxes, the font has no
braille and **nothing after this point can be judged** — a different terminal or font is
needed first.

## 1 · M1a — the picture alone (plan task 14.15)

**The rule that makes this worth doing: the answer is written down before the key is
shown.** An answer given after seeing the key is not evidence of anything.

For each scenario, at each size, at each colour depth, in this order:

1. Draw it, and look at it:

   ```
   ./dist/tuimaps --offline --scenario N --size 149x38 --headless
   ./dist/tuimaps --offline --scenario N --size 149x38 --headless --no-colour
   ./dist/tuimaps --offline --scenario N --size 69x12  --headless
   ./dist/tuimaps --offline --scenario N --size 69x12  --headless --no-colour
   ```

   N is 1, 2, 3, 4, 6 or 7. There is no 5: it is wind, which arrives with the wind
   overlay (D-44).

2. **Write the answer in [`uat-record.md`](uat-record.md) now**, before step 3. The
   question each scenario asks is in the record, and in
   [`m1-scenarios.md`](../01-objectives/m1-scenarios.md).

   "Cannot tell" is a real answer and is recorded as such. **It is a finding, not a
   pass** — except where the key puts the place less than one cell from the edge, where
   "on the edge" is the correct reading and the words are what settle it (D-67).

3. Then, and only then, show the key:

   ```
   ./dist/tuimaps --offline --scenario N --describe
   ```

   and the independent answer key, which shares no code with the library:

   ```
   cat 06_docs/02_features/go-tuimaps/02-analysis/scenarios/scenario-N-key.json
   ```

4. A scenario **fails** if the frame contradicts the key beyond the frame's own
   resolution. It **passes** if the frame agrees, or if it says "on the edge" where the
   key puts the place under one cell from it.

## 2 · The reference frames (plan task 14.16)

The twenty-four frames in [`reference-frames/`](reference-frames) are **candidates, not
goldens.** They become goldens only when approved here.

```
cat 06_docs/02_features/go-tuimaps/07-readiness/reference-frames/scenario-1-69x12-plain.txt
```

They are the same frames the commands in section 1 draw, so approving them is the same
act as judging M1a — the record has one column for it. If the map is later meant to
change, they are written again with `scripts/uat --frames`, and are candidates again
until approved again.

## 3 · The description, said aloud (plan task 14.17)

```
./dist/tuimaps --offline --scenario N --describe
```

Play each one through a speech engine and a screen reader. What is misread is written in
the record: the library hands out data and never a sentence of its own (D-52), so a
misreading is either the app's wording — which is the app's to change — or a word the
library hands out, which is the library's.

Nothing in this output is braille, a box-drawing character or a marker glyph; if any
appears, that is a defect and not a matter of taste.

## 4 · The no-colour reading (plan task 14.18, NFR-15)

The M1 questions, answered from the no-colour frames **alone** — no colour, no
description, no key:

```
./dist/tuimaps --offline --scenario N --size 69x12 --headless --no-colour
```

This may be the same sitting or a different reviewer; the record says which.

## 5 · The keys, if you want to move about (not a judged task)

```
./dist/tuimaps --offline --scenario 1
```

`?` lists the keys. `q` or `Esc` leaves. `d` shows the description in place of the map.
`Tab` focuses a place, and `a`/`z` then zoom about that place. `C` takes the colour off,
`s` keeps the library's own colours in a scale, `r` stops the blinking.

**The network is on by default in this app** and off in the library (D-65). Every command
above says `--offline`, which reaches nothing at all and draws from the tiles built into
the program — the world down to zoom 3. Leave `--offline` off to fetch a real basemap
from OpenFreeMap; the scenarios then sit on a drawn map rather than on an empty one. The
judged frames are the offline ones, so that they can be reproduced exactly.
