package packagecloud

import "testing"

func TestDistro(t *testing.T) {
	id, ok := distributions.DebianDistroVersionID("debian", "stretch")
	if !ok {
		t.Fatal("not found")
	}
	if id != "149" {
		t.Fatalf("want 149, got %s", id)
	}
}
