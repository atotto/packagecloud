package packagecloud

import "context"

type packagecloudContextKey string

const (
	packagecloudTokenKey packagecloudContextKey = "packagecloud token"
)

func WithPackagecloudToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, packagecloudTokenKey, token)
}

func packagecloudToken(ctx context.Context) string {
	return ctx.Value(packagecloudTokenKey).(string)
}
