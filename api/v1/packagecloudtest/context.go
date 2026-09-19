package packagecloudtest

import (
	"context"
	"testing"

	packagecloud "github.com/atotto/packagecloud/api/v1"
)

func SetupToken(tb testing.TB, ctx context.Context, token string) context.Context {
	if token == "" {
		tb.Skip("PACKAGECLOUD_TOKEN is not set")
	}
	return packagecloud.WithPackagecloudToken(ctx, token)
}
