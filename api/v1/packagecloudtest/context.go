package packagecloudtest

import (
	"context"
	"testing"

	packagecloud "github.com/tyklabs/packagecloud/api/v1"
)

func SetupToken(tb testing.TB, ctx context.Context, token string) context.Context {
	if token == "" {
		tb.Fatalf("empty token")
	}
	return packagecloud.WithPackagecloudToken(ctx, token)
}
