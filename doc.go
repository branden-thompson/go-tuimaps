// Package tuimaps draws a map in a terminal, as braille cells, and lets a
// host application lay its own data over it: alert areas, radar images,
// scalar fields such as temperature.
//
// The library starts no goroutine, opens no connection it was not told to,
// reads no keys and owns no clock. A host renders with Render and lets the
// slow work happen by calling Work from goroutines of its own.
package tuimaps
