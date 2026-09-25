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
| 2026-09-23T22:46:08Z | ec23e21 | 8222c07688f0 | docs | green | 75 | - |
| 2026-09-23T23:09:22Z | 0e2519c | 44d0b9709a01 | full | green | 1275 | - |
| 2026-09-23T23:11:04Z | b332ce8 | 029b186b984a | docs | green | 75 | - |
| 2026-09-23T23:28:47Z | ed576fe | e70ce0cc39b6 | docs | green | 74 | - |
| 2026-09-23T23:32:42Z | f90a79b | d4750d1ef26e | docs | green | 73 | - |
| 2026-09-23T23:45:38Z | 657b549 | 97dde8ff2b82 | docs | green | 76 | - |
| 2026-09-23T23:50:49Z | 6b5e989 | 39cd8abb2595 | docs | green | 75 | - |
| 2026-09-23T23:57:32Z | c3438c6 | 4864cc984fbc | docs | green | 73 | - |
| 2026-09-24T00:22:44Z | dfb641c | 44f62ce6fd07 | full | FAILED | 1143 | - |
| 2026-09-24T00:43:50Z | dfb641c | faeaa9e9c9c5 | full | green | 1190 | - |
| 2026-09-24T00:45:26Z | b613d80 | 0d59b47becd4 | docs | green | 73 | - |
| 2026-09-24T00:47:25Z | cb2facc | 1b789a2928c2 | docs | green | 72 | - |
| 2026-09-24T01:02:13Z | 12ef75b | dc0f939b509a | docs | green | 75 | - |
| 2026-09-24T01:06:40Z | f0a1af3 | 16bfd517b323 | docs | green | 74 | - |
| 2026-09-24T01:42:44Z | 6ff6d98 | 531130429575 | docs | green | 81 | - |
| 2026-09-24T01:45:49Z | 3141ff2 | e2643a23e424 | docs | green | 73 | - |
| 2026-09-24T02:48:38Z | bdc4737 | 5d291806307d | full | green | 1332 | - |
| 2026-09-24T03:40:40Z | 8b0ace7 | 33176a95b461 | full | green | 1286 | - |
| 2026-09-24T04:03:09Z | 2989387 | 9096e9e923a3 | full | green | 1242 | - |
| 2026-09-24T04:25:12Z | b513225 | ad2350913861 | full | green | 1132 | - |
| 2026-09-24T13:52:36Z | 63bc37b | 39714a60c711 | full | green | 1287 | - |
| 2026-09-24T14:13:34Z | a481cd9 | d03f0f9bb7e6 | full | green | 1197 | - |
| 2026-09-24T14:37:24Z | e2189f0 | 300d4cb2daad | full | FAILED | 1345 | - |
| 2026-09-24T14:57:46Z | e2189f0 | 300d4cb2daad | full | green | 1199 | - |
| 2026-09-24T15:35:26Z | 1863421 | 1b08824190e9 | full | green | 1312 | - |
| 2026-09-24T15:59:28Z | 75bd4fa | 6d70e764665d | full | green | 1194 | - |
| 2026-09-24T16:08:23Z | ee301c5 | 2354638a8460 | docs | green | 76 | - |
| 2026-09-24T16:39:51Z | d14c8d4 | db1ba2fdf4b8 | full | green | 1260 | - |
| 2026-09-24T17:07:12Z | d4db280 | a36a635a5ea4 | full | green | 1251 | - |
| 2026-09-24T17:35:45Z | 8b0d805 | 0bbe9c3d5361 | full | green | 1355 | - |
| 2026-09-24T17:56:31Z | 3702d68 | 27bc2f2d9ac2 | full | FAILED | 1121 | - |
| 2026-09-24T18:18:00Z | 3702d68 | 27bc2f2d9ac2 | full | green | 1273 | - |
| 2026-09-24T19:07:31Z | 39cf026 | 3ff947accbd8 | full | green | 1330 | - |
| 2026-09-24T19:56:50Z | af9c2ac | 61a5d493dc9f | full | FAILED | 1933 | - |
| 2026-09-24T20:23:44Z | af9c2ac | b88367c92ce2 | full | green | 1400 | - |
| 2026-09-24T20:53:32Z | 9a59a2d | 9afaf92b45ee | full | FAILED | 1370 | - |
| 2026-09-24T21:15:34Z | 9a59a2d | e77374b03779 | full | green | 1305 | - |
| 2026-09-24T21:44:23Z | be4b77a | f388a753be76 | full | green | 1395 | - |
| 2026-09-24T23:52:25Z | 02fcc06 | 0ab87124d81d | full | FAILED | 1393 | - |
| 2026-09-25T00:15:07Z | 02fcc06 | b04883967a81 | full | green | 1319 | - |
| 2026-09-25T00:37:54Z | deb2370 | bf834da504f7 | full | green | 1290 | - |
| 2026-09-25T02:17:21Z | f591eed | b1184a680dd1 | full | green | 1406 | - |
| 2026-09-25T02:43:19Z | 740a3d1 | a464e7c60573 | full | green | 1241 | - |
| 2026-09-25T03:15:28Z | 53ea6d8 | dd596242c5ec | full | green | 1273 | - |
| 2026-09-25T03:44:55Z | f37a78e | 62e54c6ca5bd | full | FAILED | 1333 | - |
| 2026-09-25T04:07:46Z | f37a78e | a4a4262d31d4 | full | green | 1334 | - |
| 2026-09-25T04:37:16Z | 2e7aa65 | 2ec89b1f20c2 | full | green | 1370 | - |
| 2026-09-25T04:37:54Z | c651486 | 6e6ba945bca4 | docs | FAILED | 0 | - |
| 2026-09-25T05:05:57Z | c651486 | 1401f9f14567 | full | green | 1401 | - |
