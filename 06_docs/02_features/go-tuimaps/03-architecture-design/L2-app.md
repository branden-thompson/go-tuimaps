# Level 2 — The standalone app

Up: [architecture](architecture.md) · Carries: FR-5, FR-34, NFR-15, D-17, D-52, D-65, D-73, D-75, D-94, P-68a, P-68b, P-72a, P-72b, L-16

A separate module (D-75). It is the library's first host: it writes its own pump, as any host must (D-73), and it is what HUM LEAD judges M1a in.

```mermaid
flowchart TB
    ARGS["Flags: --headless · --describe --place · --offline · --no-cache · --purge · --verify ·<br/>--safe-ramps · --reduce-motion · --no-colour · --lang · --style PATH (D-94) · --size COLSxROWS · --scenario (loads an M1 scenario's places and overlays)"] --> MODE{"Mode"}
    MODE -- "--describe" --> D["New · SetPlaces · Set(overlays) · Settle · Describe → plain text, no map, no terminal control<br/>a screen-reader user's path (D-52)"]
    MODE -- "--headless" --> HL["New(WithSize from --size, else the terminal's) · Settle · Render → one complete frame to standard output"]
    MODE -- "interactive" --> TUI

    subgraph TUI["Interactive"]
      direction TB
      TERM["Terminal: raw mode, restored on EVERY exit path, including a panic (L-16)"]
      KEYS["Keys → intents: pan · zoom — about the focused place when one is focused, the keyboard equivalent of zoom-toward-the-pointer (FR-5) · fit world · labels · water ·<br/>safe ramps · reduce motion · no colour · layers (P-68a)<br/>every state reachable by keys alone (NFR-15)"]
      LOOP["Draw loop: Render on a tick and whenever the pump says 'changed';<br/>sleeps until NextCall when nothing is due"]
      PUMP["Its own pump: two goroutines calling Work, woken by OnPending —<br/>the same ten lines the examples show"]
      FOOT["Footer and help (P-72a): states that maps need a font with braille,<br/>that the network is ON by default here, and lists every accessibility switch"]
      KEYS --> LOOP
      PUMP --> LOOP
    end
    TUI --> NET{"--offline?"}
    NET -- no --> ON["Source(OpenFreeMap) — the app's default, said in help (D-65)"]
    NET -- yes --> OFF["Embedded tiles only"]
    OUTR["Everything printed passes through the library's cleaning (FR-34)"] -.-> D
    OUTR -.-> HL
```

## What can change this diagram

| If this changes… | …this moves |
|---|---|
| Pointer operations are built (after v0.1.0) | A mouse branch beside KEYS, each with its key equivalent (D-17) |
| Tours are built | Keys `g` and `t` (P-68b), the footer's tour segment (P-72b) |
