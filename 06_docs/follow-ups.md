# Carried follow-ups

Work that is deliberately not done yet, with enough context to pick it up cold. **This file is the
record** — a follow-up that lives only in a chat log, an agent's memory or a commit body is not
carried, it is forgotten. Each row names what, why it waits, and what would make it due.

Close a row by deleting it in the commit that does the work.

## Open

| # | Item | Why it waits | What makes it due | Raised by |
|---|---|---|---|---|
| F-2 | **The package's architecture is not pluggable the way watchpost's is.** Watchpost separates `domains/` (data sources), `platform/` (shared leaves), `modes/` (surfaces) and `app/` (the only composition root), and enforces the boundaries with `scripts/lint-imports.sh`. go-tuiMaps has `internal/` packages but no declared extension seam for sources (tiles, radar, overlays) or renderers, and radar sources, colour tables and a block renderer are all arriving. | HUM LEAD, 2026-09-22: *"No action at this time - but make a note of it - we'll handle it in a quality pass."* The v0.1.0 contract is public; any seam must be additive (contract §9). | The quality pass. Its "consider before v0.2.0 adds a second radar source" fired at v0.2.0 D-19; **the first seam, for provider colour tables, is taken in v0.2.0 (D-35, L-2.5)**, and the broad restructure stays here. | HUM LEAD, during watchpost 0.18.0 DISCOVER |
| F-1 | **The repository root is crowded.** 61 tracked files sit at the root: 16 library sources, 39 `_test.go` files (several named by milestone — `m1a_test.go`, `m1b_test.go` — rather than by subject), plus module and project files. A reader looking for the public API has to find 16 files among 61. | HUM LEAD, 2026-09-22 (as F-2). Go constrains the fix: tests of the public package must sit in its directory, so hygiene here means naming and grouping, and possibly moving internal-only tests into `internal/`, not moving the package. | The quality pass, alongside F-2 — a restructure for pluggability would move files anyway. | HUM LEAD, during watchpost 0.18.0 DISCOVER |
