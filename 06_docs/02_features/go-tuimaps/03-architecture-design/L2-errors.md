# Level 2 — Errors and warnings

Up: [architecture](architecture.md) · [contract, section 7](contract.md) · Carries: FR-22b, FR-34, NFR-22

One rule decides which a problem becomes: **refused → an error; accepted but off, or found later during `Work` → a warning.** Both pass through the one way out, so neither can carry an escape byte, a tile address or a foreign error.

```mermaid
flowchart TB
    subgraph IN["Where problems are found"]
      direction LR
      H["A hand-in or a call<br/>Set · Remove · settings · intents"]
      W["During Work<br/>fetch · gate · decode · prepare · describe"]
      R["During Render"]
    end
    H --> REF{"Refused?"}
    REF -- yes --> ERR["Error: a kind from the closed list<br/>what happened · why · what to do"]
    REF -- "yes, and it was a Set" --> ALSO["…and a 'set-refused' warning,<br/>so a discarded error is still visible"]
    REF -- "no, but something is off" --> WARN
    W -- "a job failed or met something odd" --> WARN["Warning: kind · the overlay or tile it concerns · a count<br/>at most 64 kept · duplicates counted, not repeated"]
    W -- "Work itself cannot proceed" --> ERR
    R -- "a panic, recovered at the call's edge" --> FAILF["Frame with status 'failed', last good rows kept<br/>and a 'render-failed' warning"]
    ERR --> SAFE
    WARN --> SAFE
    ALSO --> WARN
    FAILF --> SAFE
    SAFE{{"The way out: text-safety's own type<br/>cleaned · outside text cut to 64 clusters · scheme and host only, never an address<br/>no foreign error is ever wrapped (FR-34, FR-22b)"}}
    SAFE --> HOST["Host: the returned error · Warnings() · the frame's status"]
```

| Error kinds (closed) | Warning kinds (closed) |
|---|---|
| invalid-coordinates · size-mismatch · unsorted-breaks · malformed-ramp · missing-table · malformed-table · unknown-preset · invalid-id · over-vertex-cap · over-image-cap · image-refused · ring-too-short · bad-currency · unsupported-schema · unsupported-tile · over-limit · fetch-refused · fetch-failed · cache-refused · no-size · reentrant-call · cancelled · closed · internal | ramp-rule-broken · unmatched-image-colours · stale-overlay · future-valid-time · implausible-unit · near-duplicate-id · set-refused · borrow-changed · no-work-called · tile-failed · cache-write-failed · render-failed · cache-under-need |

## What can change this diagram

| If this changes… | …this moves |
|---|---|
| A kind is added — a minor version (NFR-22) | The two lists, here and in the contract |
| The warning cap | The WARN box and the constants table |
