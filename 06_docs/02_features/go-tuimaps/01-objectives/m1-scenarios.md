# M1 — Hazard Placement: the scenario set

| Field | Value |
|---|---|
| Ratified | HUM LEAD, 2026-09-18 (D-43): "Recommendation approved" |
| Why this exists | The project brief promised M1 a "fixed scenario set" and never defined one; two red-team reviewers found the primary metric unmeasurable as written. |
| Change control | This set is M1's denominator. It changes only by a recorded HUM LEAD ruling — the same guard M3 has. |

## The metric

M1 is the share of the scenarios below for which a viewer can answer the stated questions **from the frame alone**, without leaving the terminal. Target: all of them. *In v0.1.0, all of those whose overlay shape is in the first slice (D-44) — the wind scenario waits for the release that builds wind.*

## The scenarios

Each is rendered at both of the first host's map sizes — 69×12 and 149×38 cells — in the braille renderer, at truecolor **and** with no colour.

| # | Scenario | The viewer must say |
|---|---|---|
| 1 | An alert with its own polygon (5–21 vertices — about one alert in ten), the place inside it | Inside or outside |
| 2 | An alert defined by a **zone shape** (900–14,000 vertices — about nine alerts in ten), the place close to the edge | Inside or outside; which way the nearest edge lies |
| 3 | **Radar**, with rain near but not over the place — the most common view (D-35) | Direction and rough distance to the nearest rain; how heavy |
| 4 | A temperature field | The band at the place; which way it gets warmer |
| 5 | Wind | Direction and rough speed at the place |
| 6 | Point hazards: an earthquake and a fire detection | Direction and rough distance from the place |
| 7 | A storm track passing the place | Which side it passes; roughly how close |

"Rough distance" means within a factor of two, read from the map's distance reference (FR-33).

## How it is judged

**HUM LEAD judges, against a computed answer key.** For every scenario the right answers — inside or outside, bearing, distance, class — are calculated from the geometry and the data, not read off the picture. A scenario passes only if HUM LEAD's reading of the frame matches the key. The key is also an automated test: no frame may contradict it.

Scenario 2 is the one to watch. It is the common real-world case, it had not been rendered at the time of this ruling, and it is where simplifying a large outline could give a wrong "inside" answer when one cell spans kilometres (RS-19).

## Where the live session sits

**M1 gates go-tuiMaps' own SHIP on the library**, using fixtures shaped like real host data, shown live in the standalone app. **The session inside Watchpost is a Watchpost metric**, measured after Watchpost integrates. This replaces the brief's wording, "a live HUM LEAD acceptance session inside Watchpost", which could not gate SHIP: Watchpost cannot import the library until a remote exists, and that is decided at SHIP (D-19).

The limit of this, recorded when it was ruled: passing on fixtures is a claim about the library, not about a person using Watchpost, where the rectangle, the theme and the data are all different. The Watchpost session remains the real test of the problem statement; it simply cannot be this project's gate.
