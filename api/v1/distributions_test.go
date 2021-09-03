package packagecloud_test

import (
	"context"
	"os"
	"testing"

	packagecloud "github.com/atotto/packagecloud/api/v1"
	"github.com/atotto/packagecloud/api/v1/packagecloudtest"
)

func TestDistro(t *testing.T) {
	ctx := context.Background()
	ctx = packagecloudtest.SetupToken(t, ctx, os.Getenv("PACKAGECLOUD_TOKEN"))

	distributions, err := packagecloud.GetDistributions(ctx)
	if err != nil {
		t.Fatal(err)
	}

	id, ok := distributions.DebianDistroVersionID("debian", "stretch")
	if !ok {
		t.Fatal("not found")
	}
	if id != "149" {
		t.Fatalf("want 149, got %s", id)
	}
}
