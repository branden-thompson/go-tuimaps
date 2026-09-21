package textsafe

import "testing"

// TestEach: a label is laid out a cluster at a time, each with its width; a
// cluster is never split, and the widths add up to Width's answer.
func TestEach(t *testing.T) {
	text := Clean("Ze\xcc\x81 \xe6\x9d\xb1\xe4\xba\xac") // "Ze" with a combining acute, a space, two wide characters
	var clusters []string
	total := 0
	Each(text, func(cluster string, width int) bool {
		clusters = append(clusters, cluster)
		total += width
		return true
	})
	if len(clusters) != 5 || clusters[1] != "e\xcc\x81" || total != Width(text) || total != 7 {
		t.Errorf("clusters %q, %d cells; Width says %d", clusters, total, Width(text))
	}
	seen := 0
	Each(text, func(string, int) bool { seen++; return seen < 2 })
	if seen != 2 {
		t.Errorf("walked %d clusters after being told to stop at 2", seen)
	}
	Each(Text{}, func(string, int) bool { t.Error("an empty text has no cluster"); return true })
	Each(text, nil)
}
