# Gate runs

`scripts/gate` appends one line here per run. M6 counts consecutive clean full runs from the current
table. A line is written by the script, never by hand. The log starts at v0.2.0 D-31; commits before it
have no line.

## D-31 to D-40

**Commit** here is HEAD while the run happened, and **Uncommitted** counts changed files, so these lines
cannot say exactly which tree was tested.

| When (UTC) | Commit | Uncommitted | Mode | Result | Seconds |
|---|---|---|---|---|---|
| 2026-09-23T18:47:45Z | 7288375 | 6 | full | green | 1091 |
| 2026-09-23T18:53:50Z | a3dd98e | 6 | docs | FAILED | 1 |
| 2026-09-23T19:11:50Z | a3dd98e | 6 | full | green | 1072 |
| 2026-09-23T19:12:46Z | f163b1d | 2 | docs | green | 43 |
| 2026-09-23T19:14:32Z | fbc8e9e | 4 | docs | green | 42 |
| 2026-09-23T19:16:05Z | 1a75d10 | 4 | docs | green | 41 |
| 2026-09-23T19:22:09Z | 86e0b3f | 4 | docs | green | 43 |
| 2026-09-23T19:50:08Z | aabfcea | 18 | full | green | 1257 |
| 2026-09-23T20:05:07Z | 11d986a | 2 | docs | green | 48 |
| 2026-09-23T20:07:34Z | 1998e52 | 3 | docs | green | 46 |

## From D-40

**HEAD** is the commit the run sat on — the parent of any commit made from it. **Tree** is the git tree
that was tested, leaving this file out: the staged change for `docs`, the whole working tree otherwise.
A commit made from exactly that has the same tree less this file, so the two can be matched.
**Overrides** names any setting of the gate that was changed for the run.

| When (UTC) | HEAD | Tree | Mode | Result | Seconds | Overrides |
|---|---|---|---|---|---|---|
| 2026-09-23T20:41:41Z | acd5ca3 | 74f68071c141 | full | FAILED | 1231 | - |
| 2026-09-23T21:05:52Z | acd5ca3 | dbcd52b9d83f | full | green | 1234 | - |
| 2026-09-23T21:07:34Z | a315410 | b72b2d9bcb7d | docs | green | 71 | - |
| 2026-09-23T21:18:03Z | 2762930 | a266b5754594 | docs | green | 73 | - |
| 2026-09-23T21:20:52Z | db2cca2 | db4b764ad156 | docs | green | 70 | - |
| 2026-09-23T21:42:28Z | 41b17bb | f8ed703de654 | full | green | 1175 | - |
| 2026-09-23T21:46:21Z | 0206f85 | dc610260b279 | docs | green | 72 | - |
| 2026-09-23T21:49:30Z | bd062b4 | 1f7758804035 | docs | FAILED | 70 | - |
| 2026-09-23T22:09:13Z | 6291726 | fd4bee6200bc | full | green | 1161 | - |
| 2026-09-23T22:30:45Z | 758e23a | b72aee74d8ba | full | green | 1237 | - |
| 2026-09-23T22:32:28Z | 4065b06 | ce37a1a5c94a | docs | green | 74 | - |
