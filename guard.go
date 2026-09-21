package tuimaps

import (
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// plant is how the library's own tests plant a panic inside a public call,
// to show that none of them escapes (contract, section 6, rule 4). Nothing
// but those tests ever sets the hook, and the one comparison it costs is the
// price of proving the rule rather than asserting it.
func (m *Map) plant(call string) {
	if m == nil || m.planted == nil {
		return
	}
	m.planted(call)
}

// guard turns a panic inside a public call into an error of the internal
// kind. No panic escapes a public call, and the map stays usable after one
// (contract, section 6, rule 4).
func guard(call string, err *error) {
	r := recover()
	if r == nil {
		return
	}
	if err != nil {
		*err = panicked(call)
	}
}

// guardQuiet is guard for a call that returns no error: the panic is stopped
// and noted as a warning the host can ask for.
func (m *Map) guardQuiet(call string) {
	r := recover()
	if r == nil {
		return
	}
	m.keepWarning(fault.Warning{Kind: fault.RenderFailed, Subject: textsafe.Clean(call)})
}

// keepWarning keeps a warning the map raised itself, for the next Warnings
// call. The list is capped, as every warning list is (NFR-20).
func (m *Map) keepWarning(w fault.Warning) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.own) < 64 {
		m.own = append(m.own, w)
	}
}

// panicked is the error a recovered panic becomes. It says nothing of what
// the panic was: that is the library's own business, and a host can do
// nothing with it but report it.
func panicked(call string) error {
	return fault.Make(fault.Internal, textsafe.Const("the library recovered from a defect of its own"),
		textsafe.Clean("a call to "+call+" raised a panic, which was stopped at the edge of the library"),
		textsafe.Const("this is a defect in the library; report it, with what the map was showing"))
}
