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
	"os"
	"testing"

	"github.com/edgexr/dnsproviders/api"
	"github.com/stretchr/testify/require"
)

func TestCloudflareDNS(t *testing.T) {
	// skip unless needed to debug
	t.Skip("skipping cloudflare DNS test")

	ctx := context.Background()

	key := os.Getenv("CFKEY")
	domain := os.Getenv("DOMAIN")

	if key == "" {
		t.Errorf("missing CFKEY environment variable")
	}
	if domain == "" {
		t.Errorf("missing DOMAIN environment variable")
	}

	prov, err := GetProvider(ctx, api.CloudflareProvider, "", map[string]string{"token": key}, nil)
	require.Nil(t, err)
	ProviderTest(t, ctx, prov, domain)
}
