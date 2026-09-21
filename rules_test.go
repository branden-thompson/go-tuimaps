package tuimaps_test

import (
	"os"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/rules"
	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

// TestStaticRules is plan task 00.11 applied to the library itself: no
// goroutine, no clock, no output, no forbidden import, and every test
// package held to loopback.
func TestStaticRules(t *testing.T) {
	found, err := rules.Check(".", testkit.ModulePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range found {
		t.Error(f)
	}
}
