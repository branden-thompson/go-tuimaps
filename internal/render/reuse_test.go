package render

import (
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/style"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

// TestReuseKey is plan task 09.24: each thing a frame is a function of forces
// a redraw when it changes, and a call that changes none of them does not.
func TestReuseKey(t *testing.T) {
	v := fitted(60, 16)
	base := input(t, v)
	r, err := NewRenderer(v.Cols, v.Rows)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Draw(base); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Draw(base); err != nil || r.Redraws() != 1 {
		t.Fatalf("%d redraws for two calls with nothing changed, %v", r.Redraws(), err)
	}
	changes := map[string]func(in *Input){
		"the view":         func(in *Input) { in.View.Zoom += 0.25 },
		"the look":         func(in *Input) { in.Look++ },
		"the depth":        func(in *Input) { in.Depth = NoColour },
		"labels":           func(in *Input) { in.Labels = false },
		"the scale mark":   func(in *Input) { in.Scale = true },
		"the credit":       func(in *Input) { in.Credit = textsafe.Const("other") },
		"what is missing":  func(in *Input) { in.Missing = 2 },
		"the style":        func(in *Input) { in.Style = style.BuiltIn() },
		"the tiles":        func(in *Input) { in.Tiles = in.Tiles[:len(in.Tiles)-1] },
		"a tile sharpened": func(in *Input) { in.Tiles = append([]Drawn(nil), in.Tiles...); in.Tiles[0].Exact = false },
	}
	for name, change := range changes {
		before := r.Redraws()
		changed := base
		change(&changed)
		if _, err := r.Draw(changed); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if r.Redraws() != before+1 {
			t.Errorf("%s changed and the frame was not redrawn", name)
		}
		if _, err := r.Draw(base); err != nil || r.Redraws() != before+2 {
			t.Errorf("%s changed back and the frame was not redrawn", name)
		}
	}
}

// TestOnlyChangedRowsRebuilt and TestFrameValidUntilNextRender are plan task
// 09.25 (contract, section 5).
func TestOnlyChangedRowsRebuilt(t *testing.T) {
	v := fitted(120, 16) // wide enough for the scale mark beside the credit line, which comes first (FR-14)
	base := input(t, v)
	r, _ := NewRenderer(v.Cols, v.Rows)
	if _, err := r.Draw(base); err != nil {
		t.Fatal(err)
	}
	first := r.RowsBuilt()
	if first != 16 {
		t.Fatalf("%d rows built for the first frame of 16", first)
	}
	withScale := base
	withScale.Scale = true
	f, err := r.Draw(withScale)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.RowsBuilt() - first; got != 1 {
		t.Errorf("%d rows rebuilt when only the bottom row's scale mark changed", got)
	}
	if !strings.Contains(plain(f.Lines[15]), " km") {
		t.Error("the changed row does not show the change")
	}
}

func TestFrameValidUntilNextRender(t *testing.T) {
	v := fitted(60, 16)
	base := input(t, v)
	r, _ := NewRenderer(v.Cols, v.Rows)
	a, _ := r.Draw(base)
	kept := strings.Join(a.Lines, "\n")
	b, _ := r.Draw(base)
	if strings.Join(b.Lines, "\n") != kept {
		t.Error("the frame reused is not the frame drawn")
	}
	other := base
	other.Labels = false
	if _, err := r.Draw(other); err != nil {
		t.Fatal(err)
	}
	if strings.Join(a.Lines, "\n") == kept {
		t.Log("the first frame's rows survived the next render; the contract only promises them until then")
	}
}
