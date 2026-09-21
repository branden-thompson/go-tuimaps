package jsonsafe

import (
	"os"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/testkit"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

func TestWithin(t *testing.T) {
	cases := []struct {
		name  string
		body  string
		bytes int
		depth int
		want  bool
	}{
		{"flat", `{"a":1}`, 100, 1, true},
		{"one too deep", `{"a":[1]}`, 100, 1, false},
		{"exactly deep enough", `{"a":[1]}`, 100, 2, true},
		{"brackets in a string", `{"a":"[[[[[["}`, 100, 1, true},
		{"an escaped quote in a string", `{"a":"\"[[[["}`, 100, 1, true},
		{"an escaped backslash ends the string", `{"a":"\\","b":[[1]]}`, 100, 2, false},
		{"too long", `{"a":1}`, 3, 9, false},
		{"empty", ``, 100, 9, false},
		{"no limit given", `{}`, 0, 9, false},
		{"no depth given", `{}`, 100, 0, false},
		{"66 deep against 64", strings.Repeat("[", 66) + strings.Repeat("]", 66), 1000, 64, false},
	}
	for _, c := range cases {
		if got := Within([]byte(c.body), c.bytes, c.depth); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}
