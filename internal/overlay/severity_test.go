package overlay

import (
	"testing"
	"time"

	"github.com/branden-thompson/go-tuimaps/internal/colour"
	"github.com/branden-thompson/go-tuimaps/internal/fault"
	"github.com/branden-thompson/go-tuimaps/internal/project"
)

func alertFeature(role colour.Token, s Severity) Feature {
	return Feature{Kind: Polygon, Role: role, Severity: s,
		Rings: [][]project.LonLat{{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}}}}
}

// TestAFeatureCarriesItsSeverity is L5.1 (L-13.9): a feature with no
// severity reports the one its role implies; one it states wins; a severity
// outside the list is refused at hand-in; a feature that is no alert has none.
func TestAFeatureCarriesItsSeverity(t *testing.T) {
	for role, want := range map[colour.Token]Severity{
		colour.AlertExtremeOutline: SeverityExtreme, colour.AlertSevereOutline: SeveritySevere,
		colour.AlertModerateOutline: SeverityModerate, colour.AlertMinorOutline: SeverityMinor,
		colour.AlertUnknownOutline: SeverityUnknown, colour.AlertSevereTint: SeveritySevere,
	} {
		if got := SeverityOf(alertFeature(role, 0)); got != want {
			t.Errorf("role %v with no severity: %v, want %v", role.Name(), got, want)
		}
	}
	if got := SeverityOf(alertFeature(colour.AlertSevereOutline, SeverityMinor)); got != SeverityMinor {
		t.Errorf("a stated severity: %v, want minor", got)
	}
	if got := SeverityOf(Feature{Kind: Line, Role: colour.LabelRegion}); got != 0 {
		t.Errorf("a feature that is no alert: %v, want none", got)
	}
	bad := Overlay{ID: "alerts", Valid: noon, Keeps: time.Hour, Features: []Feature{alertFeature(colour.AlertSevereOutline, SeverityExtreme+1)}}
	if _, err := store(t).HandIn(bad); !isKind(err, fault.UnknownPreset) {
		t.Errorf("a severity outside the list: %v", err)
	}
	good := bad
	good.Features = []Feature{alertFeature(colour.AlertSevereOutline, SeverityExtreme)}
	good.Features[0].Valid, good.Features[0].Expires = noon, noon.Add(time.Hour)
	if _, err := store(t).HandIn(good); err != nil {
		t.Errorf("a stated severity and times: %v", err)
	}
}
