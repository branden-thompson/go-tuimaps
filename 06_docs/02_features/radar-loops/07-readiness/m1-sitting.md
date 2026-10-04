# M1 — the motion sitting (plan task L10.9)

Up: [release checklist](release-checklist.md) row 0.9 · [requirements](../01-objectives/requirements.md) M1, L-1.12 ·
Rulings D-24, D-71, D-121, D-122, D-123

**Human-graded.** M1 is done when the HUM LEAD's scores for both arms are recorded below. Nothing in the
scores is an agent's: they are transcribed from the HUM LEAD's answers.

## The loops

Five loops recorded from real radar (IEM's archived N0Q composite, twelve frames five minutes apart), in
`testdata/loops/`; `programs/m1-capture.py.txt` re-captures them. Loop 1, the Dayton outbreak of 28 May
2019, has several cells (D-71). (OW-11's Alabama outbreak is also in `testdata/loops`, for the demo app's
tests; it is not one of the sitting's five, because its words were seen before the sitting.) A snow loop,
the mid-Atlantic blizzard of 23 January 2016, was captured for the five and dropped before the sitting:
its snow never reached the heavier class, so it had no motion to score; the Mayfield tornado of 11
December 2021 took its place.

| Loop | Directory |
|---|---|
| 1 | `testdata/loops/outbreak-ohio-2019-05-28` |
| 2 | `testdata/loops/moore-oklahoma-2013-05-20` |
| 3 | `testdata/loops/derecho-indiana-2012-06-29` |
| 4 | `testdata/loops/mayfield-kentucky-2021-12-11` |
| 5 | `testdata/loops/harvey-houston-2017-08-26` |

## Before you start

In a terminal, in the go-tuimaps folder, build the app once:

```
scripts/uat
```

It writes `dist/tuimaps`. Nothing below reaches the network: every command carries `--offline`, and the
loops are files in the repository.

## The order (D-71), and why it is the order

**Step 1 - the ground truth, from the picture alone.** Do this for all five loops before Step 2, so no
words about a loop reach you first. For each loop, run its command, watch, and write in the Ground truth
column the direction the heavier rain (the reds and purples) moves, on an 8-point compass: N, NE, E, SE,
S, SW, W or NW.

```
dist/tuimaps --offline --loop testdata/loops/outbreak-ohio-2019-05-28
dist/tuimaps --offline --loop testdata/loops/moore-oklahoma-2013-05-20
dist/tuimaps --offline --loop testdata/loops/derecho-indiana-2012-06-29
dist/tuimaps --offline --loop testdata/loops/mayfield-kentucky-2021-12-11
dist/tuimaps --offline --loop testdata/loops/harvey-houston-2017-08-26
```

The loop starts playing. Keys: `p` plays and stops, `[` steps one frame back, `]` one frame on, `a` and
`z` zoom, the arrow keys pan, `q` quits to run the next loop. **Do not press `d`** in this step: it shows
the words.

**Step 2 - the non-visual arm, from the words alone.** For each loop, run its command, read only the line
it prints, and write in the Non-visual arm column the direction those words give. These are the library's
own words (D-121; watchpost's phrasing is scored at its M1b).

```
dist/tuimaps --offline --describe --loop testdata/loops/outbreak-ohio-2019-05-28
dist/tuimaps --offline --describe --loop testdata/loops/moore-oklahoma-2013-05-20
dist/tuimaps --offline --describe --loop testdata/loops/derecho-indiana-2012-06-29
dist/tuimaps --offline --describe --loop testdata/loops/mayfield-kentucky-2021-12-11
dist/tuimaps --offline --describe --loop testdata/loops/harvey-houston-2017-08-26
```

The words say which way the heavier rain near the centre of the view moves, and how fast, over the
whole loop, then where the nearest of it is (D-122). An invented example, not one of the five: "heavier
rain near the view's centre is moving east at about 50 kilometres an hour, over the 55 minutes to 18:25;
the nearest is 15 kilometres west, coming closer" - it moves E. Write the direction the words give.

**Step 3 - the visual arm.** Play each loop again with the Step 1 commands, and write in the Visual arm
column the direction you see now.

**Pass:** for each arm, four of five loops within one compass point of the ground truth (NE is within one
point of N and of E).

## Scores

### Sitting 1 - 2026-10-04

The HUM LEAD's answers, transcribed as given (their message of 2026-10-04, answers 1-15 in the order of
Steps 1, 2 and 3).

| Loop | Ground truth | Non-visual arm | Within one point? | Visual arm | Within one point? |
|---|---|---|---|---|---|
| 1 | E | E | yes | E | yes |
| 2 | NE | NW | no | NE | yes |
| 3 | SE | NW | no | SE | yes |
| 4 | NE | E | yes | NE | yes |
| 5 | N | NE | yes | N | yes |

| | |
|---|---|
| Scored by | HUM LEAD |
| Date | 2026-10-04 |
| Non-visual arm | 3/5 within one point - **fail** |
| Visual arm | 5/5 within one point - pass |

**Why the non-visual arm failed.** Every sentence spanned five minutes, two frames of twelve. All twelve
frames reached the tracker; the tracker followed one connected area of heavier rain back a frame at a time
and stopped at the first frame where that area had split, merged or shifted its centroid by more than
the step allowed. Over five minutes a centroid's jitter is larger than the rain's movement, so the two
sightings gave the wrong direction for loops 2 and 3. The words also left the listener to work the
direction out from two positions.

**What followed (D-122, D-123, D-127).** The motion is now measured over every frame of the loop, from
the heavier rain near the place, and the words say its heading and speed first. Over these five loops,
every pair of frames measured:

| Loop | Ground truth | Heading | Off by | Nearest point |
|---|---|---|---|---|
| 1 | E (90) | 99 | 9 | E |
| 2 | NE (45) | 63 | 18 | NE |
| 3 | SE (135) | 90 | 45 | E |
| 4 | NE (45) | 48 | 3 | NE |
| 5 | N (0) | 330 | 30 | NW |

Each nearest point is within one point of the ground truth; three match it, and loop 3 is at the edge.
**These are the loops the method was built on, with the ground truth in view**, so
`TestTheMotionOfSittingOneIsWithinAPoint` is a regression pin, not evidence for M1 (D-127). By D-123 there
is no second sitting: M1 is recorded as this sitting's fail, remedied by D-122. No host voices the words
yet (OW-21); a blind sitting on fresh loops is the evidence when one does. The visual arm's 5/5 is the
same reader's reading of the same pictures as the ground truth, minutes apart (D-71's design): it
confirms the reading, and is not quoted as evidence of its own.

**`--offline` hid the radar** in the interactive app, not in `--headless` or `--describe`. The tiles the
view wanted below the built-in zoom failed, and the app's pump stopped at the first failed job, so the
loop's preparation, queued behind them, never ran. Fixed: a failed job is that job's, and the pump goes on
(`TestAnOfflineLoopIsDrawnThroughThePump`; the rule is in contract section 2). The commands above are
correct as written.
