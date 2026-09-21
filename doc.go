// Package tuimaps draws a map in a terminal, as braille cells, and lets a
// host application lay its own data over it: alert areas, radar images,
// scalar fields such as temperature.
//
// The library starts no goroutine, opens no connection it was not told to,
// reads no keys and owns no clock. A host renders with Render and lets the
// slow work happen by calling Work from goroutines of its own.
//
// # What changes the map, and what only reads it
//
// **Every call that moves the map changes it**: Recentre, Zoom, FitTo, Set,
// Remove, AddPlace, the look settings - and Render itself, which takes the
// size it is given and keeps it. Only Centre, Credits, Legend, Scale,
// Overlays, Places, Warnings, Pending, Changed and DeepestZoom just read.
//
// This matters to a host built the way terminal applications usually are,
// around a model that is updated in one place and drawn in another. **The
// drawing half must not move the map**: a view built from a value receiver
// cannot call Recentre, and a key press has to travel back to the updating
// half and change the map there. In a Bubble Tea host that means a command,
// not a call from View.
//
// Render is the exception worth naming twice, because it does not look like
// one: it is how a frame is drawn, and it is also what tells the map how big
// it is. A host that draws the same map at two sizes has resized it twice.
package tuimaps
