package tuimaps

import (
	"github.com/branden-thompson/go-tuimaps/internal/style"
)

// SetStyle draws the basemap by a style of the host's own: the same JSON
// format as the library's, with the legacy filter form (FR-20). It takes
// effect at the next Render and reads no tile again - a style says what is
// drawn and in what colour, and the tiles on hand already carry every layer
// the library reads.
//
// **The library is handed the bytes and never reads a file.** A host that
// keeps its style on disk reads it itself, as the app's --style does (D-94).
// Passing no bytes at all puts the library's own style back, as passing no
// names does to the palette.
//
// A style that cannot be read is refused with the malformed-style kind and
// the map goes on drawing with the style it had.
func (m *Map) SetStyle(body []byte) (err error) {
	defer guard("SetStyle", &err)
	m.plant("SetStyle")

	if m == nil {
		return closed()
	}
	if len(body) == 0 {
		return m.useStyle(style.BuiltIn())
	}
	// The style is read outside the lock: it is a host's own bytes, of any
	// size up to the limit, and nothing else waits on it.
	theirs, err := style.Parse(body)
	if err != nil {
		return err
	}
	return m.useStyle(theirs)
}

// useStyle puts a style in place for the next frame.
func (m *Map) useStyle(s *style.Style) error {
	if s == nil {
		return closed()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shut {
		return closed()
	}
	m.style = s
	m.changed++
	return nil
}
