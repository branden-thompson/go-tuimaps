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
| 2026-09-25T05:37:14Z | 8be0aac | fcee12227eab | full | FAILED | 1319 | - |
| 2026-09-25T06:00:32Z | 8be0aac | 3ad4ccd1e8ed | full | green | 1335 | - |
| 2026-09-25T06:32:22Z | 29bd4dc | 518651e01dbf | full | green | 1404 | - |
| 2026-09-25T07:13:05Z | 5838354 | 716cd8716409 | full | green | 1399 | - |
| 2026-09-25T10:38:51Z | 91fa894 | 19d113bb1fd4 | full | green | 1322 | - |
| 2026-09-25T16:53:14Z | 48d6873 | 805a2d65a711 | docs | green | 88 | - |
| 2026-09-25T20:41:05Z | 2c6f9ce | a54b8261ae3c | full | green | 1616 | - |
| 2026-09-25T23:59:24Z | 31b69b7 | fc7aafc08839 | full | FAILED | 1558 | - |
| 2026-09-26T00:25:04Z | 31b69b7 | d35c193f3514 | full | green | 1446 | - |
| 2026-09-26T02:42:11Z | acf94b4 | 97f11a420c37 | full | green | 1463 | - |
| 2026-09-26T04:54:39Z | ff62509 | 460f95e877ef | full | FAILED | 1550 | - |
| 2026-09-26T05:20:29Z | ff62509 | 4b97330f7ecb | full | green | 1383 | - |
| 2026-09-26T22:49:38Z | 90bb558 | a7794c8559c0 | full | green | 1569 | - |
| 2026-09-27T00:04:53Z | 7edb8d1 | c77e5d1485f5 | full | green | 1675 | - |
| 2026-09-27T01:50:11Z | 33e2016 | 2bf1450b1339 | full | FAILED | 1572 | - |
| 2026-09-27T02:18:37Z | 33e2016 | 13b640f70b76 | full | green | 1508 | - |
| 2026-09-27T04:33:49Z | 9e333bf | 142adb7dbe4e | full | FAILED | 1581 | - |
| 2026-09-27T04:34:35Z | 9e333bf | 142adb7dbe4e | full | green | 25 | - |
| 2026-09-27T05:00:35Z | 9e333bf | f602ea341e49 | full | FAILED | 1551 | - |
| 2026-09-27T05:29:22Z | 9e333bf | e2fe0923e0bb | full | green | 1620 | - |
| 2026-09-27T18:04:44Z | 28a8f1f | aeba5354a887 | full | green | 1534 | - |
| 2026-09-27T20:16:40Z | 9323eb7 | 8f3a7f7f44db | full | green | 1654 | - |
| 2026-09-27T21:59:53Z | 5c43aa4 | f809aaae6163 | full | FAILED | 1510 | - |
| 2026-09-27T22:25:18Z | 5c43aa4 | 1be38e75f419 | full | green | 1501 | - |
| 2026-09-28T01:27:10Z | 9280caf | 9d4298d3a16b | full | green | 1633 | - |
| 2026-09-28T03:47:34Z | 0e2add3 | 4f6754688c04 | full | green | 1624 | - |
| 2026-09-28T06:02:10Z | e252abc | a6802fef0cb6 | full | green | 1623 | - |
| 2026-09-28T14:50:47Z | 11372c8 | 2fa1aa21387d | full | green | 1646 | - |
| 2026-09-28T16:12:42Z | cfaa650 | 4ede675f5422 | full | green | 1684 | - |
| 2026-09-28T22:26:56Z | c52d49b | d90b4f6649a7 | full | green | 1782 | - |
| 2026-09-28T23:36:28Z | 7511c6e | 123cbea90915 | full | green | 1723 | - |
| 2026-09-29T01:17:18Z | 3b0f5a8 | b14db3f6ae44 | full | green | 1712 | - |
| 2026-09-30T08:08:54Z | 1c0ca83 | 9ab5bacb878b | full | FAILED | 1752 | - |
| 2026-09-30T08:37:08Z | 1c0ca83 | 13eb2476466f | full | green | 1648 | - |
| 2026-09-30T11:56:28Z | fc40cbc | 281b07c806d2 | full | green | 1653 | - |
| 2026-09-30T22:06:37Z | 0f0c829 | 0cbcc409f370 | full | FAILED | 1770 | - |
| 2026-09-30T22:35:24Z | 0f0c829 | 65e67af94749 | full | green | 1706 | - |
| 2026-10-01T00:28:07Z | 47f1b31 | 2540d6970bb5 | full | green | 1803 | - |
| 2026-10-01T16:21:43Z | 4935935 | bf245978c344 | full | FAILED | 1718 | - |
| 2026-10-01T16:50:27Z | 4935935 | a2ad6a4caddd | full | green | 1661 | - |
| 2026-10-01T18:12:22Z | 6a59519 | ab85b99b2b87 | full | green | 1769 | - |
| 2026-10-01T23:53:11Z | e9d2675 | 16d988838214 | full | FAILED | 1941 | - |
| 2026-10-02T00:24:59Z | e9d2675 | f625d474806c | full | green | 1815 | - |
| 2026-10-02T03:52:18Z | ecfbca7 | cd51ee8c29bd | full | green | 1820 | - |
| 2026-10-02T16:01:01Z | c7fc104 | c4eaa8d784cd | full | FAILED | 1730 | - |
| 2026-10-02T16:32:23Z | c7fc104 | 2178804d86c4 | full | green | 1599 | - |
| 2026-10-03T06:32:14Z | d84d68b | 52856175f7bd | full | FAILED | 1998 | - |
| 2026-10-03T07:05:11Z | d84d68b | 71d7e33862fe | full | green | 1878 | - |
| 2026-10-04T01:16:37Z | eeac4da | c87553b374d4 | full | green | 1859 | - |
| 2026-10-04T02:07:51Z | b494a6b | 605bc1818b51 | full | green | 1771 | - |
| 2026-10-04T02:59:42Z | 039e786 | 69d1dae50175 | release | green | 5 | - |
| 2026-10-04T02:59:45Z | 039e786 | 69d1dae50175 | release | FAILED | 3 | - |
| 2026-10-04T03:35:52Z | 039e786 | 17e363f55f31 | full | FAILED | 1989 | - |
| 2026-10-04T04:09:05Z | 039e786 | 2b0ed316fb26 | full | green | 1844 | - |
| 2026-10-04T05:23:40Z | 4916295 | 315a27d6934a | full | FAILED | 1824 | - |
| 2026-10-04T05:58:28Z | 4916295 | d4ae193b6f33 | full | green | 1797 | - |
| 2026-10-04T06:20:02Z | 80e3bd3 | 5d54244749aa | release | FAILED | 4 | - |
| 2026-10-04T07:04:21Z | 80e3bd3 | bb0c03f5dc91 | soak | green | 3603 | - |
| 2026-10-04T07:41:12Z | 80e3bd3 | fae101c3a682 | full | green | 2188 | - |
| 2026-10-04T07:42:56Z | 83ed367 | fae101c3a682 | release | green | 3 | - |
| 2026-10-04T08:38:27Z | 83ed367 | 72d214adbe15 | docs | green | 171 | - |
| 2026-10-04T13:08:39Z | 39e2a8a | 26a8c7131345 | full | green | 2320 | - |
| 2026-10-04T13:59:03Z | 39e2a8a | 7c9ddb708266 | full | green | 2154 | - |
| 2026-10-04T14:12:37Z | 228acd9 | 946588497590 | docs | green | 168 | - |
| 2026-10-04T16:59:29Z | 72cf0e5 | b19065491705 | full | green | 2796 | - |
| 2026-10-04T17:52:49Z | a7228b3 | 81ecb8e9127e | full | green | 2601 | - |
| 2026-10-04T18:13:41Z | 33dc8d7 | 95e08bd3d2a8 | docs | green | 302 | - |
| 2026-10-04T19:14:08Z | 609fec1 | 8cabb434206e | soak | green | 3604 | - |
| 2026-10-04T20:12:45Z | 609fec1 | c98ce8dc8bb8 | full | FAILED: govulncheck is not the pinned v1.8.0: go install golang.org/x/vuln/cmd/govulncheck@v1.8.0; cmd/tuimaps: govulncheck is not the pinned v1.8.0: go install golang.org/x/vuln/cmd/govulncheck@v1.8.0; examples: govulncheck is not the pinned v1.8.0: go install golang.org/x/vuln/cmd/govulncheck@v1.8.0; tools/answer-key: govulncheck is not the pinned v1.8.0: go install golang.org/x/vuln/cmd/govulncheck@v1.8.0; tools/atlas: govulncheck is not the pinned v1.8.0: go install golang.org/x/vuln/cmd/govulncheck@v1.8.0; tools/gen-assets: govulncheck is not the pinned v1.8.0: go install golang.org/x/vuln/cmd/govulncheck@v1.8.0; tools/oracle: govulncheck is not the pinned v1.8.0: go install golang.org/x/vuln/cmd/govulncheck@v1.8.0 | 2731 | - |
| 2026-10-04T21:00:01Z | 609fec1 | 0f1c30912b90 | full | FAILED: tools/oracle: fuzz FuzzAgree, 5700000x | 2699 | - |
| 2026-10-04T21:57:56Z | 609fec1 | 3926b15ddaf4 | full | green | 2836 | - |
| 2026-10-04T22:46:30Z | 8899289 | 3926b15ddaf4 | full | green | 2847 | - |
| 2026-10-04T23:33:02Z | 8899289 | 3926b15ddaf4 | full | green | 2792 | - |
| 2026-10-05T00:19:53Z | 8899289 | 3926b15ddaf4 | full | green | 2811 | - |
| 2026-10-05T01:08:48Z | 8899289 | 3926b15ddaf4 | full | green | 2934 | - |
| 2026-10-05T01:14:20Z | 8899289 | 18d13e38a8d3 | docs | green | 306 | - |
| 2026-10-05T01:21:17Z | 6b8994a | 844e3a686346 | docs | green | 310 | - |
| 2026-10-05T02:17:00Z | 32a3fc9 | 9a77347bf860 | docs | green | 315 | - |
| 2026-10-05T03:16:17Z | cc7eb2e | af6f063ab0d9 | full | green | 2884 | - |
| 2026-10-05T03:17:55Z | 2b6f480 | af6f063ab0d9 | release v0.2.0 | green | 24 | - |
| 2026-10-05T04:09:06Z | 3cacdfc | af6f063ab0d9 | release v0.2.0 | green | 24 | - |
| 2026-10-05T04:58:27Z | 3cacdfc | 6ea6f2eabe88 | full | green | 2782 | - |
| 2026-10-06T00:30:53Z | ed78e54 | 9bd37072f860 | docs | green | 343 | - |
| 2026-10-06T00:50:44Z | 5fc822a | 188b5eff8d62 | docs | green | 328 | - |
| 2026-10-06T03:28:06Z | d1c4d5e | f5f39c8eddb0 | docs | green | 337 | - |
| 2026-10-06T03:54:57Z | 0862f09 | b994acc0ebcc | docs | green | 320 | - |
| 2026-10-06T04:03:27Z | 0862f09 | 421aaeb41e38 | docs | green | 327 | - |
| 2026-10-06T04:14:07Z | a52c8f1 | c40f48c3c88c | docs | green | 324 | - |
| 2026-10-06T18:30:56Z | e5cd730 | a16d4b823e74 | docs | green | 359 | - |
| 2026-10-06T18:48:11Z | badfb0f | 18459db444af | docs | green | 360 | - |
| 2026-10-06T19:23:03Z | 5687075 | 60a037d93fa4 | docs | green | 347 | - |
| 2026-10-06T19:54:12Z | 9d02fb4 | 85a68b77fea4 | docs | green | 349 | - |
| 2026-10-06T20:13:37Z | 0ac931e | d1cc1950d417 | docs | green | 349 | - |
| 2026-10-06T21:54:05Z | 03b3c42 | e48259a7d817 | docs | green | 349 | - |
