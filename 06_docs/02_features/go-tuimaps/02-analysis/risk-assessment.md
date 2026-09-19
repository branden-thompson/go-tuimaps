# go-tuiMaps — Risk Assessment at DISCOVER exit

| Field | Value |
|---|---|
| Phase | DISCOVER (FULL RCC) |
| Date | 2026-09-18 |
| Builds on | Risk signals RS-1..RS-9 in the [project brief](../08-reports/project-brief.md); statuses in [`tier1-synthesis.md`](tier1-synthesis.md) §4; Tier 2 evidence in [`specimens/README.md`](specimens/README.md) and [`research/AI-9`](research/AI-9-pmtiles-and-embedded-tiles.md) |
| Status | For HUM LEAD approval in the Discovery Report |

Severity is the exposure that remains **after** what DISCOVER settled. Every risk has an owner-phase: where its mitigation is due.

| ID | Risk | At intake | Now | What changed | Mitigation | Due |
|---|---|---|---|---|---|---|
| RS-1 | "Parity" is unbounded | High | **Low** | Parity is behavioural against the code (D-11); 72 rows frozen with a 68-row denominator; 17 defects in a closed ledger (D-37). | Change control on the matrix: a row moves only by recorded ruling. | Held |
| RS-2 | The overlay contract is a new design seam | High | **High** | All five shapes are in v1 (D-14); one rendering core carried every specimen; a compositing order is known. The contract is still undesigned, and expensive to change once the first host imports it (D-13). | PLAN proposes more than one contract shape; the contract is exercised by the app and a host spike before it is published; versioned pre-1.0 so it may still move. | PLAN |
| RS-3 | Field legibility at braille resolution | High | **Low** | Shown legible in HUM LEAD's terminal (D-32): fields, radar, polygons and wind, at both host sizes, in both renderers, at four colour depths, with the rules in the specimen findings. | Remaining specimens in PLAN: light-background terminal (A-3), smooth contours (A-5), light-rain shade. | PLAN |
| RS-4 | Embed-first against app-first | Medium | **Low** | The model is settled: synchronous render from cache, separate cancellable fetch, host-supplied time, control as intents (FR-4, FR-23..FR-25). | The app is built on the library as an outside host would use it (D-13). | BUILD |
| RS-5 | Generality against building only what is needed | Medium | **High** | Scope was widened deliberately: five shapes (D-14), PMTiles (D-18), four colour depths and two renderers (D-23), themeable ramps and re-coloured images (D-36). Each is a ruled decision; together they are a large v1. | PLAN sets a build order with an early usable milestone — basemap, features, one field, the image shape — before the rest; no optional setting enters the contract without a consumer; bypasses and extras stay out (D-16). | PLAN |
| RS-6 | Dependence on a free third-party tile server | Medium | **Low** | The map must keep working without it (D-18): embedded zoom 0–3 (D-33), a long-lived cache, a PMTiles source, a replaceable URL. | Track the provider's schema migration (D-21). | PLAN |
| RS-7 | Embed cost inside a host with published budgets | Medium | **High** | The ceiling is set at 8 MB with structural rules (D-29), but nothing is measured; radar animation, image classes and simplified-shape caches all draw on it. Storing images as one class byte per pixel should help (D-36). | A measured prototype at PLAN exit validates or revises the figure; byte-capped caches from the first line of design; allocation tests as gates from BUILD entry. | PLAN exit |
| RS-8 | Port fidelity | Medium | **Low** | Every parity row cites file and line; nine ledger entries and the headline claims of every research report were re-checked against primary sources. | Differential tests against an independent decoder in a test-only module (AI-3 §9); reference frames. | BUILD |
| RS-9 | Watermark enforcement | Low | **Low** | Calibration in force; every commit message and file scanned; one trailer caught and amended before any push. | Red-team hygiene lens at every phase exit. | Every exit |
| RS-10 | Licence and attribution compliance | — | **Low** | Styles are fresh, so nothing is inherited (D-24); credits are reported and shown by default (D-25); the embedded data carries its own notice and its generating script is published (D-33). | A notices file crediting both upstreams, OpenMapTiles and OpenStreetMap. Reported as documented by the licensors; not legal advice. | PLAN / BUILD |
| RS-11 | Metric M1 depends on a change in the first host | — | **Medium** | 90% of live alerts carry no polygon; the host keeps none (CD-1). Specimen 4 proves the library's side. | Record in the host's backlog; M1 is measured on the library with host-shaped test data until the host catches up. | Host |
| RS-12 | Hand-written parsers of untrusted input | — | **Medium** | The PMTiles reader is ruled dependency-free (D-18); the tile and style decoders are a PLAN choice (X-5). | Fuzzing with real tiles as seeds, bounded allocation, differential tests, a published supported subset with load-time warnings (NFR-10). | BUILD |
| RS-13 | Overlay data-source terms | — | **Closed** | The host fetches; terms bind the host, not the library (D-15). | — | — |
| RS-14 | *new* **Themeable ramps that must still make sense** | — | **Medium** | D-36 lets a host theme the data ranges and requires them to stay ordered, distinct and readable, across four colour depths. How a poor ramp is detected or repaired is undesigned. | PLAN designs the ramp rules and a checker a host can run in its own tests, as the first host already does for theme contrast. | PLAN |
| RS-15 | *new* **Two renderers × four depths × style profiles is a large approval surface** | — | **Medium** | Every visual change must be re-approved across the grid (D-23), and the basemap is now a family of profiles (FR-19). | The headless render makes each combination a one-line reference test; specimens are generated, not hand-made; approvals are by contact sheet. | PLAN |
| RS-16 | *new* **The pinned planet file is undocumented by its provider** and could be withdrawn | — | **Low** | AI-9 §8. | Pin version and checksum; the generated asset is committed, so a withdrawal blocks only regeneration, not builds. | BUILD |
| RS-17 | *new* **Evidence limits of the specimens** | — | **Low** | Throwaway code; one region, one evening, a dark terminal; no timing; clipping and simplification not exercised; radar resampling rule unsettled (S6-3). | The limits are stated in the specimen record; each becomes a PLAN or BUILD check. | PLAN |

## Summary

**High: 3** — RS-2 (the contract), RS-5 (scope), RS-7 (memory). All three are about PLAN getting the design right inside a fixed budget; none is about whether the idea works. **Medium: 5.** **Low: 8.** **Closed: 1.**

The risk that could have ended the project — that a weather field cannot be read under a braille map — is retired by evidence HUM LEAD has seen in his own terminal.
