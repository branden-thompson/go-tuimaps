# Level 2 — Tests and gates

Up: [architecture](architecture.md) · Carries: NFR-22, D-19, D-81, D-86

**There is no remote until SHIP (D-19), so there is no hosted continuous integration until then.** Until SHIP the gate is a script run locally before every merge; at SHIP the same script becomes the hosted workflow. What cannot be run locally is said to be untested, not assumed.

```mermaid
flowchart TB
    subgraph MODS["Every module in the repository, one by one — not only the library"]
      direction LR
      M0["library"] --- M1["cmd/tuimaps"] --- M2["examples"] --- M3["tools/gen-assets"] --- M4["tools/answer-key"] --- M5["tools/oracle"] --- M6["tools/atlas"]
    end
    GATE["The gate script<br/>writes a throw-away workspace file so the nested modules resolve the library from this tree —<br/>with a replace for the library at exactly v0.0.0 inside that throw-away file —<br/>the tracked module files carry no replace line"] --> MODS
    MODS --> T0["A module with no packages yet is named and called EMPTY — not failed, not hidden;<br/>a module go cannot list FAILS (v0.2.0 D-31);<br/>tests or a module graph that cannot be listed FAIL (D-40)"]
    MODS --> T1["Tests under the race detector, on the floor toolchain"]
    MODS --> T2["A second leg WITHOUT the race detector: the zero-allocation and allocation-count tests live here"]
    MODS --> T3["Fuzz: every target, a COUNT of runs calibrated to ~60 s of its own work — not a clock<br/>(an hour-long run before a release is NOT BUILT)<br/>a time budget PROBABLY made the engine report its own deadline as a failure — not reproduced (L-6.4, v0.2.0 D-9, D-10)<br/>each leg also has a wall-clock limit, TERM then KILL, and fails as TIME LIMIT (v0.2.0 D-31, D-40) · every leg prints its duration"]
    MODS --> T4["Vulnerability scan: each module AND the standard library —<br/>on the machine's NEWEST toolchain, named in the report, not on the floor:<br/>the floor keeps a newer API out, and its standard library carries every vulnerability fixed since"]
    MODS --> T5["Licence file present for every module in each graph"]
    M0 --> A1["Allow-list: with the workspace file off, the library's graph is go-runewidth and uax29 only (D-81)"]
    M0 --> A2["Static checks on library packages: no 'go' statement · no timer or ticker ·<br/>no write to standard output or error · render imports neither tiles nor fetch"]
    M0 --> A3["Cross-compile, C toolchain off: macOS arm64 and amd64 · Linux amd64 and arm64 · Windows amd64"]
    M0 --> A4["Reference frames byte-identical on arm64 (native) and amd64 (emulated locally)<br/>FAILS — never skips — if the second architecture cannot be run"]
    M0 --> A5["Loopback only: the dial hook is installed for every test binary by default, not opted into<br/>a static check fails any test package that does not link it"]
    M0 --> A6["Sub-process test: a program that reuses borrowed memory too early must be caught by the race detector (D-86)"]
    M0 --> A7["From the first tag: the public contract compared with the last tag (NFR-22)"]
    T1 & T2 & T3 & T4 & T5 & A1 & A2 & A3 & A4 & A5 & A6 & A7 --> OK{"All green?"}
    OK -- yes --> MERGE["Merge allowed"]
    OK -- no --> STOP["Stop: what failed, why, what to do"]
```

| Untested until SHIP, and said so | Why |
|---|---|
| Running on Linux and Windows (only compiled) | No such machine before a remote exists |
| amd64 on real hardware | Emulated locally; the two-architecture claim is provisional until a hosted amd64 runner confirms it |
| The one-hour soak as a nightly job | Run by hand before each phase exit until then |

## What can change this diagram

| If this changes… | …this moves |
|---|---|
| SHIP creates the remote | The script becomes the hosted workflow; the "untested" table empties |
| A module is added | The MODS row |
