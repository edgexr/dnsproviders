package dnsproviders

import (
	"context"
	"errors"
	"fmt"
	"strings"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/recordsets"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/zones"

	"github.com/edgexr/dnsproviders/api"
)

type OTC struct {
	client *golangsdk.ProviderClient
	dns    *golangsdk.ServiceClient
	logger api.Logger
	region string
}

var _ api.Provider = OTC{}

const (
	CredentialKeyRegion     = "region"
	CredentialKeyTenantName = "tenantName"
	CredentialKeyDomainName = "domainName"
	CredentialKeyUsername   = "username"
	CredentialKeyPassword   = "password"
	identityEndpointFormat  = "https://iam.%s.otc.t-systems.com/v3"
)

var (
	ErrZoneNotFound   = errors.New("could not find zone by the given name")
	ErrRecordNotFound = errors.New("could not find record by the given name")
)

func NewOtcProvider(_ context.Context, _ string, credentialsData map[string]string, logger api.Logger) (*OTC, error) {
	for _, key := range []string{CredentialKeyRegion, CredentialKeyDomainName, CredentialKeyTenantName, CredentialKeyUsername, CredentialKeyPassword} {
		if _, isSet := credentialsData[key]; !isSet {
			return nil, fmt.Errorf("missing key %s is credentialData", key)
		}
	}

	client, err := openstack.AuthenticatedClient(golangsdk.AuthOptions{
		IdentityEndpoint: fmt.Sprintf(identityEndpointFormat, credentialsData[CredentialKeyRegion]),
		DomainName:       credentialsData[CredentialKeyDomainName],
		TenantName:       credentialsData[CredentialKeyTenantName],
		Username:         credentialsData[CredentialKeyUsername],
		Password:         credentialsData[CredentialKeyPassword],
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize authenticated client: %v", err)
	}

	dns, err := openstack.NewDNSV2(client, golangsdk.EndpointOpts{
		Region: credentialsData[CredentialKeyRegion],
	})

	if err != nil {
		return nil, fmt.Errorf("failed to init dns client: %v", err)
	}

	return &OTC{
		client: client,
		dns:    dns,
		region: credentialsData[CredentialKeyRegion],
		logger: logger,
	}, nil
}

func (o OTC) GetDNSRecords(ctx context.Context, zone, name string) ([]api.Record, error) {
	z, err := o.findZoneByName(ctx, zone)
	if err != nil {
		return nil, err
	}

	zoneID := z.ID

	recordSets, err := o.listRecordSets(ctx, zoneID, name)
	if err != nil {
		return nil, err
	}

	var apiRecords []api.Record
	for _, rec := range recordSets {
		fqdn := fmt.Sprintf("%s.%s", name, zone)
		if name == "" || rec.Name == fqdn {
			apiRecords = append(apiRecords, api.Record{
				Type:    rec.Type,
				Name:    name,
				Content: rec.Records,
				TTL:     rec.TTL,
			})
		}
	}

	return apiRecords, nil
}

func (o OTC) CreateOrUpdateDNSRecord(ctx context.Context, zone, name, rtype, content string, ttl int, proxy bool) error {
	z, err := o.findZoneByName(ctx, zone)
	if err != nil {
		return err
	}

	zoneID := z.ID

	records, err := o.listRecordSets(ctx, zoneID, name)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		if err := o.createDNSRecord(ctx, zoneID, fmt.Sprintf("%s.%s", name, zone), rtype, content, ttl, proxy); err != nil {
			return fmt.Errorf("failed to create record in  zoneID '%s' (zone name '%s') with name %s: %v", zoneID, zone, name, err)
		}

		return nil
	}

	if len(records) > 1 {
		return fmt.Errorf("found more than one matching DNS record with name %s in zone %s", name, zone)
	}

	recordID := records[0].ID

	// wrap content in quotation marks, if not already quoted
	if !strings.HasPrefix(content, "\"") && !strings.HasSuffix(content, "\"") {
		content = fmt.Sprintf("\"%s\"", content)
	}

	result := recordsets.Update(o.dns, zoneID, recordID, recordsets.UpdateOpts{
		TTL:     ttl,
		Records: []string{content},
	})

	if result.Err != nil {
		return fmt.Errorf("failed to update record for zone %s (name='%s'): %v", zone, name, err)
	}

	return nil
}

func (o OTC) DeleteDNSRecord(ctx context.Context, zone, name string) error {
	z, err := o.findZoneByName(ctx, zone)
	if err != nil {
		return err
	}

	zoneID := z.ID

	records, err := o.listRecordSets(ctx, zoneID, name)
	if err != nil {
		return fmt.Errorf("failed to list record sets by zoneID '%s' (zone name '%s'): %v", zoneID, zone, err)
	}

	if len(records) == 0 {
		return ErrRecordNotFound
	}

	if len(records) > 1 {
		return fmt.Errorf("found more than one matching DNS record with name %s in zone %s", name, zone)
	}

	recordID := records[0].ID

	return recordsets.Delete(o.dns, zoneID, recordID).Err
}

func (o OTC) findZoneByName(_ context.Context, name string) (*zones.Zone, error) {
	pages := zones.List(o.dns, zones.ListOpts{Name: name})

	allPages, err := pages.AllPages()
	if err != nil {
		return nil, err
	}

	allZones, err := zones.ExtractZones(allPages)
	if err != nil {
		return nil, err
	}

	for _, zone := range allZones {
		if zone.Name == name {
			return &zone, nil
		}
	}

	return nil, ErrZoneNotFound
}

func (o OTC) listRecordSets(_ context.Context, zoneID, name string) ([]recordsets.RecordSet, error) {
	pages := recordsets.ListByZone(o.dns, zoneID, recordsets.ListOpts{
		Name: name,
	})

	page, err := pages.AllPages()
	if err != nil {
		return nil, err
	}

	recordSets, err := recordsets.ExtractRecordSets(page)
	if err != nil {
		return nil, err
	}

	output := make([]recordsets.RecordSet, 0)
	for _, recordSet := range recordSets {
		values := make([]string, len(recordSet.Records))
		for idx, value := range recordSet.Records {
			values[idx] = strings.Trim(value, "\"")
		}

		recordSet.Records = values
		output = append(output, recordSet)
	}

	return output, nil
}

func (o OTC) createDNSRecord(_ context.Context, zoneID, fqdn, rtype, content string, ttl int, _ bool) error {
	if !strings.HasPrefix(content, "\"") && !strings.HasSuffix(content, "\"") {
		content = fmt.Sprintf("\"%s\"", content)
	}

	result := recordsets.Create(o.dns, zoneID, recordsets.CreateOpts{
		Name:    fqdn,
		Records: []string{content},
		TTL:     ttl,
		Type:    rtype,
	})

	return result.Err
}
