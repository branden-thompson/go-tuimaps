package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunSaysNotBuiltYet(t *testing.T) {
	var out bytes.Buffer
	if code := run(&out); code != 2 {
		t.Errorf("exit code %d, want 2", code)
	}
	if !strings.Contains(out.String(), "not built yet") {
		t.Errorf("message %q does not say the command is not built yet", out.String())
	}
}
