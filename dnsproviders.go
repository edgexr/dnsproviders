// Copyright 2024 EdgeXR, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	case api.OpenTelekomCloudProvider:
		return NewOtcProvider(ctx, zone, credentialsData, logger)
	}
	return nil, errors.New("unknown dns provider " + string(typ))
}
