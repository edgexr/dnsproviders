// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dnsproviders

import (
	"context"
	"fmt"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/edgexr/dnsproviders/api"
)

const Cloudflare = "cloudflare"

type CloudflareAPI struct {
	api    *cloudflare.API
	logger api.Logger
}

// NewCloudflareProvider creates a new Cloudflare DNS provider.
func NewCloudflareProvider(ctx context.Context, zone string, credentialsData map[string]string, logger api.Logger, ops ...Option) (*CloudflareAPI, error) {
	token, ok := credentialsData["token"]
	if !ok {
		return nil, fmt.Errorf("missing token key from cloudflare dns provider credentials data")
	}
	opts := getOptions(ops)
	apiOptions := []cloudflare.Option{}
	if opts.client != nil {
		apiOptions = append(apiOptions, cloudflare.HTTPClient(opts.client))
	}
	api, err := cloudflare.NewWithAPIToken(token, apiOptions...)
	if err != nil {
		return nil, err
	}
	return &CloudflareAPI{
		api:    api,
		logger: logger,
	}, nil
}

// GetDNSRecords returns a list of DNS records for the given domain name. Error returned otherwise.
// if name is provided, that is used as a filter
func (s *CloudflareAPI) GetDNSRecords(ctx context.Context, zone, name string) ([]api.Record, error) {
	zoneID, err := s.api.ZoneIDByName(zone)
	if err != nil {
		return nil, err
	}

	queryRecord := cloudflare.DNSRecord{}
	if name != "" {
		queryRecord.Name = name
	}

	cfrecords, err := s.api.DNSRecords(zoneID, queryRecord)
	if err != nil {
		return nil, err
	}
	records := []api.Record{}
	for _, cfrec := range cfrecords {
		record := api.Record{
			Type:    cfrec.Type,
			Name:    cfrec.Name,
			Content: []string{cfrec.Content},
			TTL:     cfrec.TTL,
		}
		records = append(records, record)
	}
	return records, nil
}

// CreateOrUpdateDNSRecord changes the existing record if found, or adds a new one
func (s *CloudflareAPI) CreateOrUpdateDNSRecord(ctx context.Context, zone, name, rtype, content string, ttl int, proxy bool) error {
	zoneID, err := s.api.ZoneIDByName(zone)
	if err != nil {
		return err
	}

	queryRecord := cloudflare.DNSRecord{
		Name: strings.ToLower(name),
		Type: strings.ToUpper(rtype),
	}
	records, err := s.api.DNSRecords(zoneID, queryRecord)
	if err != nil {
		return err
	}
	found := false
	for _, r := range records {
		found = true
		if r.Content == content {
			s.logger.InfoContext(ctx, "CreateOrUpdateDNSRecord existing record matches", "name", name, "content", content)
		} else {
			s.logger.InfoContext(ctx, "CreateOrUpdateDNSRecord updating", "name", name, "content", content)

			updateRecord := cloudflare.DNSRecord{
				Name:    strings.ToLower(name),
				Type:    strings.ToUpper(rtype),
				Content: content,
				TTL:     ttl,
				Proxied: proxy,
			}
			err := s.api.UpdateDNSRecord(zoneID, r.ID, updateRecord)
			if err != nil {
				return fmt.Errorf("cannot update DNS record for zone %s name %s, %v", zone, name, err)
			}
		}
	}
	if !found {
		addRecord := cloudflare.DNSRecord{
			Name:    strings.ToLower(name),
			Type:    strings.ToUpper(rtype),
			Content: content,
			TTL:     ttl,
			Proxied: false,
		}
		_, err := s.api.CreateDNSRecord(zoneID, addRecord)
		if err != nil {
			s.logger.InfoContext(ctx, "CreateOrUpdateDNSRecord failed", "zone", zone, "name", name, "err", err)
			return fmt.Errorf("cannot create DNS record for zone %s, %v", zone, err)
		}
	}
	return nil
}

// DeleteDNSRecord deletes DNS record specified by recordID in zone.
func (s *CloudflareAPI) DeleteDNSRecord(ctx context.Context, zone, name string) error {
	zoneID, err := s.api.ZoneIDByName(zone)
	if err != nil {
		return err
	}

	queryRecord := cloudflare.DNSRecord{}
	if name != "" {
		queryRecord.Name = name
	}

	cfrecords, err := s.api.DNSRecords(zoneID, queryRecord)
	if err != nil {
		return err
	}
	for _, rec := range cfrecords {
		err := s.api.DeleteDNSRecord(zoneID, rec.ID)
		if err != nil {
			return fmt.Errorf("delete DNS record %v failed, %v", rec, err)
		}
	}
	return nil
}
