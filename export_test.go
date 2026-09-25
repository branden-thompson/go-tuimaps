package tuimaps

// PlantPanic makes the next call of a given name panic, so that the library's
// own tests can show that no panic escapes a public call (contract, section
// 6, rule 4). It is in a test file: no build of the library carries it.
func PlantPanic(m *Map, call string) {
	m.planted = func(at string) {
		if at == call {
			panic("planted: " + at)
		}
	}
}

// ClearPanic takes the planted panic away again.
func ClearPanic(m *Map) { m.planted = nil }

// ReportRemembered reports whether a report for these places is already
// worked out and would be returned from memory.
func ReportRemembered(m *Map, asked []Place) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.rememberedReport(asked)
	return ok
}
