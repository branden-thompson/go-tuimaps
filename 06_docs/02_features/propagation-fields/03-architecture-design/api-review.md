# go-tuiMaps v0.3.0's public API, reviewed for a host developer (watchpost D-216)

Up: [rulings](../02-analysis/rulings.md) · [contract](../../go-tuimaps/03-architecture-design/contract.md) · [public surface](../../go-tuimaps/07-readiness/public-surface.txt)

**What this is.** Before v0.3.0-rc.2 is tagged (watchpost D-206), the public API was read as a host
developer meets it: the names, the shape of each call, and how easily a host invokes it. Each change
below is made before the tag and is in the contract's section 12 and the release notes; what was weighed
and left is said with why.

## Made

| Before | After | Why |
|---|---|---|
| any host type of two classes drawn as its edge alone | `Type.Edge`, asked for on a host type of one break; a preset, or another number of classes, refused | a two-class field meant to be filled was drawn as a line nobody asked for (BUILD red team C-18; D-216's "explicit edge mark") |
| a wrong number of colours refused as `UnsortedBreaks` | refused as `MalformedRamp`, the kind that already names it | the breaks were fine; a host switching on the kind was told the wrong fault (C-17, D-216's "truthful error kind") |
| one `RampRuleBroken` warning a `Set` of a host type with bad colours | one an overlay, replaced on each `Set`, cleared once a `Set` passes | a host setting its overlay each refresh saw the same warning pile up (C-26) |
| the night composed by hand: `NightSide`, `Terminator`, two roles and an `Overlay` | `NightOverlay(id, at)`, as `MUFGrid` is for a field | the one shape every host drawing the night needs, in the library's roles, in one call (watchpost's `mapNight` was that composition) |
| `socks5h` proxies with no port given none | port 1080, as `socks5` | the remote-resolving form a host names to keep DNS off its own resolver (I-7) |

`CheckedDialerFor(ProxyFunc)` stays as it is and is adopted by watchpost for its tiles (D-216's
"CheckedDialerFor in use"), retiring watchpost's own proxy exemption.

## Weighed and left

| Call | Seen | Left because |
|---|---|---|
| `Map.Units(miles, fahrenheit bool)` | two booleans read poorly at a call site (`m.Units(true, false)`) | since v0.1.0; an options type would break every host for a readability gain, short of the significant upside a revisited decision needs (watchpost D-58); its doc names both |
| `Map.Verify() (int, int, error)` | two counts | its results are named in the source, `checked` and `removed`, and its doc says which is which |
| `Map.Close() int` | a count from a close | it says how many of the host's calls are still inside the map, which a host shutting down reads |
| the preset constructors' argument order (`TemperatureGrid(id, grid, unit, valid)`, `MUFGrid(id, grid, valid)`) | the unit sits between grid and time only where a preset has units | consistent within itself: a preset with units takes them in that place |
