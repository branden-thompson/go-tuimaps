# M1 — the motion sitting (plan task L10.9)

Up: [release checklist](release-checklist.md) row 0.9 · [requirements](../01-objectives/requirements.md) M1, L-1.12 ·
Rulings D-24, D-71

**Human-graded.** M1 is done when the HUM LEAD's scores for both arms are recorded below. Nothing here is
filled in by an agent.

## The loops

Five loops recorded from real radar (IEM's archived N0Q composite, twelve frames five minutes apart), in
`testdata/loops/`; `programs/m1-capture.py.txt` re-captures them. One has several cells (D-71); the first is
OW-11's outbreak.

| Loop | Directory |
|---|---|
| 1 | `testdata/loops/outbreak-alabama-2011-04-27` |
| 2 | `testdata/loops/moore-oklahoma-2013-05-20` |
| 3 | `testdata/loops/derecho-indiana-2012-06-29` |
| 4 | `testdata/loops/blizzard-mid-atlantic-2016-01-23` |
| 5 | `testdata/loops/harvey-houston-2017-08-26` |

## The order (D-71), and why it is the order

1. **Ground truth first, from the picture alone.** For each loop, before reading any words about it: play
   it and write the direction the heavier rain moves, on an 8-point compass (N, NE, E, SE, S, SW, W, NW).
   Build the app with `scripts/uat`, then `dist/tuimaps --offline --loop <directory>`; `p` plays and
   stops, `[` and `]` step a frame. Do not press `d` yet.
2. **The non-visual arm, from the words alone.** For each loop, read only the words and write the direction
   they give: `dist/tuimaps --describe --offline --loop <directory>` prints them. (D-71 names watchpost's
   worded description; until watchpost pins a v0.2.0 candidate it cannot voice these recorded loops, and the
   library's own demo words stand in - a ruling for the HUM LEAD.)
3. **The visual arm.** For each loop, play it again at the map's ordinary size and write the direction.

**Pass:** four of five loops within one point of the ground truth, for each arm.

## Scores

| Loop | Ground truth | Non-visual arm | Within one point? | Visual arm | Within one point? |
|---|---|---|---|---|---|
| 1 | | | | | |
| 2 | | | | | |
| 3 | | | | | |
| 4 | | | | | |
| 5 | | | | | |

| | |
|---|---|
| Scored by | |
| Date | |
| Non-visual arm | /5 within one point - pass / fail |
| Visual arm | /5 within one point - pass / fail |
