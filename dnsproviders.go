package dnsproviders

import (
	"context"
	"errors"
	"log/slog"

	"github.com/edgexr/dnsproviders/api"
)

func GetProvider(ctx context.Context, typ api.ProviderType, zone string, credentialsData map[string]string, logger api.Logger) (api.Provider, error) {
	if logger == nil {
		logger = slog.Default()
	}

	switch typ {
	case api.CloudflareProvider:
		return NewCloudflareProvider(ctx, zone, credentialsData, logger)
	case api.GoogleCloudDNSProvider:
		return NewGoogleCloudDNSProvider(ctx, zone, credentialsData, logger)
	}
	return nil, errors.New("unknown dns provider " + string(typ))
}
