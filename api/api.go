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

// Package api defines the common API interface for DNS managers
package api

import (
	"context"
)

const (
	RecordTypeA     = "A"
	RecordTypeAAAA  = "AAAA"
	RecordTypeCNAME = "CNAME"
)

// Provider common interface for managing DNS entries.
type Provider interface {
	// GetDNSRecords returns a list of DNS records. If name is
	// provided, that is used as a filter.
	GetDNSRecords(ctx context.Context, zone, name string) ([]Record, error)
	// CreateOrUpdateDNSRecord changes the existing record if found,
	// or adds a new one
	CreateOrUpdateDNSRecord(ctx context.Context, zone, name, rtype, content string, ttl int, proxy bool) error
	// DeleteDNSRecord deletes all DNS records for the name.
	DeleteDNSRecord(ctx context.Context, zone, name string) error
}

// ProviderType enumerates the types of providers supported
type ProviderType string

const (
	CloudflareProvider     ProviderType = "cloudflare"
	GoogleCloudDNSProvider ProviderType = "googleclouddns"
)

// Record represents a DNS record in a zone.
type Record struct {
	Type    string   `json:"type,omitempty"`
	Name    string   `json:"name,omitempty"`
	Content []string `json:"content,omitempty"`
	TTL     int      `json:"ttl,omitempty"`
}

// Logger interface allows a logger to be used by the providers.
// This uses a context to support opentracing span-based logging.
type Logger interface {
	InfoContext(ctx context.Context, msg string, keysAndValues ...interface{})
}
