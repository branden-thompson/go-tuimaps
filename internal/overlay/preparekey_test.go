package overlay

import "testing"

// TestAPreparationsKeyReadsBack is the key's one owner both ways: what Key
// writes PreparesOf reads, an id with slashes in it whole; any other work's
// key is not a preparation's; and the reading allocates nothing, since a
// render pays it for every job waiting or running.
func TestAPreparationsKeyReadsBack(t *testing.T) {
	s := &Store{}
	for _, c := range []struct {
		id     string
		bucket int
	}{{"radar/us", 2}, {"radar/fc-us-c", 11}, {"temperature/us/2026-10-02T18", 0}} {
		key := s.PrepareJob(c.id, c.bucket).Key()
		id, bucket, ok := PreparesOf(key)
		if !ok || id != c.id || bucket != c.bucket {
			t.Errorf("%q reads back as %q at %d (%v); want %q at %d", key, id, bucket, ok, c.id, c.bucket)
		}
	}
	for _, key := range []string{"tile/3/2/1", "overlay/radar", "overlay/radar/x", ""} {
		if _, _, ok := PreparesOf(key); ok {
			t.Errorf("%q reads as a preparation", key)
		}
	}
	if n := testing.AllocsPerRun(100, func() { _, _, _ = PreparesOf("overlay/radar/us/2") }); n != 0 {
		t.Errorf("reading a key allocates %v times", n)
	}
}
