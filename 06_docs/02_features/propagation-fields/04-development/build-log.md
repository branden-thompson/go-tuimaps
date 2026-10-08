# BUILD log — go-tuiMaps v0.3.0, propagation fields

Up: [implementation plan](implementation-plan.md)

| Field | Value |
|---|---|
| Phase | BUILD, opened 2026-10-07 at the PLAN gate (watchpost D-139; D-19) |
| What this is | One row for each plan task as it is finished: the commit, the test first, and anything learned that the plan did not say |
| Gate for every row | `scripts/gate`, full; a row that changes the public surface ends in a release candidate tag (`v0.3.0-rc.N`) |

## P0 — Specimens first (C-5)

| Task | Commit | The test first | Learned |
|---|---|---|---|
| P0.1 The whole-globe and host-type specimens | this commit | `TestTheSpecimensAreWhatTheLibraryDraws` (`specimen_test.go`): three specimens, a whole-globe 2° grid with the temperature preset and the same grid in a host type with lines off and on, each at truecolor, 256 and 16 colours on a dark and a light ground, 18 golden frames in `testdata/specimens` (448 KB). The frames were written with `-update-specimens`; it passes twice in a row under the race detector, and fails with one byte added to a golden frame and with one frame removed | **L-2.4's blank depends on the depth.** At truecolor and 256 colours a host-type grid without lines draws nothing on either ground: its frame is the bare world's, byte for byte. At 16 colours it draws its labelled contours whether lines are on or off, and the two frames are identical. So P2.3's matrix must cover depth × ground × lines, as L-2.4 says, and the 16-colour case is a different finding from the blank: lines off is not honoured there. The whole-world view clips the grid at about 84°N and 56°S (`FitWorld`), so the polar rows are not in these frames. No public name changes; no release candidate |
