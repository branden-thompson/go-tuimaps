package fault

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/branden-thompson/go-tuimaps/internal/testkit"
	"github.com/branden-thompson/go-tuimaps/internal/textsafe"
)

func TestMain(m *testing.M) { os.Exit(testkit.Main(m)) }

func sample(k Kind) *Error {
	return Make(k, textsafe.Const("the overlay was refused"), textsafe.Const("its breaks are not in rising order"), textsafe.Const("sort the breaks and set it again"))
}

// contractKinds reads the two closed lists from the contract's own text, so
// the code and the document cannot drift apart unnoticed.
func contractKinds(t *testing.T) (errorKinds, warningKinds []string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "06_docs", "02_features", "go-tuimaps", "03-architecture-design", "contract.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "| Kinds ") {
			continue
		}
		cells := strings.Split(line, " | ")
		if len(cells) != 3 {
			t.Fatalf("the contract's kinds row has %d cells, want 3", len(cells))
		}
		split := func(cell string) []string {
			var names []string
			for _, n := range strings.Split(strings.TrimSuffix(strings.TrimSpace(cell), " |"), "\u00B7") {
				names = append(names, strings.TrimSpace(n))
			}
			return names
		}
		return split(cells[1]), split(cells[2])
	}
	t.Fatal("the contract has no kinds row")
	return nil, nil
}

func TestKindsEqualTheContract(t *testing.T) {
	wantErrors, wantWarnings := contractKinds(t)
	var gotErrors, gotWarnings []string
	for k := Kind(1); k <= lastKind; k++ {
		gotErrors = append(gotErrors, k.String())
	}
	for k := WarningKind(1); k <= lastWarningKind; k++ {
		gotWarnings = append(gotWarnings, k.String())
	}
	if !reflect.DeepEqual(gotErrors, wantErrors) {
		t.Errorf("error kinds\n got %v\nwant %v", gotErrors, wantErrors)
	}
	if !reflect.DeepEqual(gotWarnings, wantWarnings) {
		t.Errorf("warning kinds\n got %v\nwant %v", gotWarnings, wantWarnings)
	}
	if Kind(0).String() != "unknown" || (lastKind+1).String() != "unknown" || WarningKind(0).String() != "unknown" {
		t.Error("a value outside a list must name itself unknown, never borrow a real kind's name")
	}
}

func TestErrorSaysWhatWhyAndWhatToDo(t *testing.T) {
	err := sample(UnsortedBreaks)
	if err.Kind() != UnsortedBreaks {
		t.Errorf("kind %v", err.Kind())
	}
	for _, part := range []string{"the overlay was refused", "its breaks are not in rising order", "sort the breaks and set it again", "unsorted-breaks"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("%q does not contain %q", err.Error(), part)
		}
	}
	var target *Error
	if !errors.As(error(err), &target) || target.Kind() != UnsortedBreaks {
		t.Error("a host must be able to reach the kind with errors.As")
	}
}

func TestAnUnknownKindBecomesInternal(t *testing.T) {
	for _, k := range []Kind{0, lastKind + 1, 255} {
		if got := sample(k).Kind(); got != Internal {
			t.Errorf("New with kind %d gave %v; a kind that does not exist is the library's own fault", k, got)
		}
	}
}

// TestAnErrorWithASentenceMissingIsTheLibrarysFault: every error says what
// happened, why, and what to do next. One that cannot is a defect in the
// library, and says so, never an empty message.
func TestAnErrorWithASentenceMissingIsTheLibrarysFault(t *testing.T) {
	empty := textsafe.Text{}
	full := textsafe.Const("something")
	for _, err := range []*Error{Make(OverLimit, empty, full, full), Make(OverLimit, full, empty, full), Make(OverLimit, full, full, empty)} {
		if err.Kind() != Internal {
			t.Errorf("kind %v; an error missing a sentence must be internal", err.Kind())
		}
		if !strings.Contains(err.Error(), "over-limit") || strings.Contains(err.Error(), ": : ") || strings.HasSuffix(err.Error(), " .") {
			t.Errorf("%q must name the kind that was meant and leave no empty sentence", err.Error())
		}
	}
}

// TestNoForeignErrorWrapped is plan task 02.11, second half.
func TestNoForeignErrorWrapped(t *testing.T) {
	if _, has := reflect.TypeOf(&Error{}).MethodByName("Unwrap"); has {
		t.Fatal("Error has an Unwrap method; no foreign error is ever wrapped")
	}
	for k := Kind(1); k <= lastKind; k++ {
		err := error(sample(k))
		if errors.Unwrap(err) != nil {
			t.Errorf("%v unwraps to something", k)
		}
		cancelled := errors.Is(err, context.Canceled) && errors.Is(err, context.DeadlineExceeded)
		if k == Cancelled && !cancelled {
			t.Error("the cancelled kind must answer errors.Is for both context errors, so a host's usual check works")
		}
		if k != Cancelled && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
			t.Errorf("%v answers errors.Is for a context error", k)
		}
	}
	text := sample(Cancelled).Error()
	if strings.Contains(text, context.Canceled.Error()) || strings.Contains(text, context.DeadlineExceeded.Error()) {
		t.Errorf("%q carries a context error's own text", text)
	}
}

func TestOutsideTextCanOnlyArriveCleaned(t *testing.T) {
	err := Make(InvalidID, textsafe.Quote("the id \x1b[2Jbad\u202E"), textsafe.Const("it holds characters that cannot be shown"), textsafe.Const("use plain text"))
	if strings.ContainsAny(err.Error(), "\x1b\u202E") {
		t.Errorf("%q carries what cleaning should have removed", err.Error())
	}
}

func TestWarningCarriesKindSubjectAndCount(t *testing.T) {
	w := Warning{Kind: StaleOverlay, Subject: textsafe.Const("radar"), Count: 3}
	if w.Kind.String() != "stale-overlay" || w.Subject.String() != "radar" || w.Count != 3 {
		t.Errorf("%+v", w)
	}
}
