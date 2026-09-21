package testkit

import (
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) { os.Exit(Main(m)) }

type fakeRun struct {
	guarded bool
	code    int
}

func (f *fakeRun) Run() int {
	tr := http.DefaultTransport.(*http.Transport)
	f.guarded = tr.DialContext != nil
	return f.code
}

func TestMainGuardsTheTransportWhileTestsRun(t *testing.T) {
	f := &fakeRun{code: 7}
	if got := Main(f); got != 7 {
		t.Errorf("Main returned %d; want the tests' own exit code, 7", got)
	}
	if !f.guarded {
		t.Error("the default transport was not guarded while the tests ran")
	}
	if got := Main(nil); got == 0 {
		t.Error("Main(nil) must not report success")
	}
}
