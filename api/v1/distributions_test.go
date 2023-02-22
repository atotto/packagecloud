package packagecloud

import (
	"context"
	"os"
	"testing"

	"github.com/atotto/packagecloud/api/v1/packagecloudtest"
)

func TestDistro(t *testing.T) {
	ctx := context.Background()
	ctx = packagecloudtest.SetupToken(t, ctx, os.Getenv("PACKAGECLOUD_TOKEN"))

	distributions, err := GetDistributions(ctx)
	if err != nil {
		t.Fatal(err)
	}

	id, ok := findDistroVersionID(distributions.Deb, "debian", "stretch")
	if !ok {
		t.Fatal("not found")
	}
	if id != "149" {
		t.Fatalf("want 149, got %s", id)
	}
}
