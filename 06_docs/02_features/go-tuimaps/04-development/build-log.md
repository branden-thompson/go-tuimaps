# BUILD log — go-tuiMaps v0.1.0

Up: [implementation plan](implementation-plan.md)

| Field | Value |
|---|---|
| Phase | BUILD, opened 2026-09-19 by HUM LEAD's approval of the Plan of Record (D-95) |
| What this is | One row for each plan task as it is finished: the commit, what the failing test was, and anything learned that the plan did not say. Where a task changes the design, the diagram changes in the same commit (D-71) and the row says which |
| Gate for every row | Tests pass with the race detector and on the floor toolchain (Go 1.25.0); the code-quality check reports no finding in changed code; the structure check is 18 of 18 |

## WP-00 — Scaffold and gates

| Task | Commit | The failing test, first | Learned |
|---|---|---|---|
| 00.1 BUILD-entry checklist | `828d712`, `5ccec3d` | — | Go restored to the language declaration (D-20). First code-quality check: every tool ran; the only findings were the invariant-density rule on three placeholder commands. **Decision, the coordinator's:** the placeholders were removed and not padded with checks that assert nothing — each command arrives with its own work package; its module file stays. Floor toolchain 1.25.0 is on this machine and runs the tests with the race detector |
| 00.2 Module layout; no `replace` | `828d712` | `TestModuleHasNoReplace`, `TestReplaceLinesFindsEveryForm`, `TestWorkspaceFileIsIgnored` — failed to compile, then failed on the missing module files | Six module files. The workspace file is ignored before it exists |
| 00.9 The fixture is pinned | `8524ee8` | `TestFixturePinned` and five more — failed to compile | The code-quality check's bounded-loop rule refuses a scanner loop; ranging over the file's lines is bounded by the file. `internal/testkit` holds the loaders |

**Next:** 00.10 (the scene types), 00.3 (the allow-list), 00.4 to 00.6 (the rest of the test kit), 00.7 and 00.12 (the gate script), 00.8 and 00.11 (lint and the static checks). 00.13, the parity mapping, was written in PLAN.
