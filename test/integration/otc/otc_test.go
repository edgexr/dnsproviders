package otc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexr/dnsproviders"
	"github.com/edgexr/dnsproviders/api"
)

const (
	credentialFile = "config.json"
)

var (
	provider       *dnsproviders.OTC
	testZone       = "filled-with-data-from-config.json"
	testRecordName = fmt.Sprintf("test-%s", strings.Split(uuid.NewString(), "-")[0])
	ipv4           = "80.0.0.0"
	ipv4Alt        = "80.0.0.1"
	ipv6           = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
)

func TestMain(m *testing.M) {
	if _, err := os.Stat(credentialFile); errors.Is(err, os.ErrNotExist) {
		log.Println("no credential file found, skipping tests for otc")
		os.Exit(0)
	}

	credentialData, err := readCredentials(credentialFile)
	if err != nil {
		panic(fmt.Sprintf("failed to read credential file: %v", err))
	}

	testZone = credentialData.TestZone

	otc, err := dnsproviders.NewOtcProvider(context.Background(), testZone, credentialData.ToMap(), nil)
	if err != nil {
		panic(err)
	}

	provider = otc

	os.Exit(m.Run())
}

func TestCreateRecord(t *testing.T) {
	type testCase struct {
		Zone          string
		Type          string
		Content       string
		ExpectedError error
	}

	for name, tc := range map[string]testCase{
		"no error create a-record": {
			Zone:          testZone,
			Type:          api.RecordTypeA,
			Content:       ipv4Alt,
			ExpectedError: nil,
		},
		"no error create aaaa-record": {
			Zone:          testZone,
			Type:          api.RecordTypeAAAA,
			Content:       ipv6,
			ExpectedError: nil,
		},
		"no error update a-record": {
			Zone:          testZone,
			Type:          api.RecordTypeA,
			Content:       ipv4,
			ExpectedError: nil,
		},
		"invalid zone": {
			Zone:          "non-existing.edge.telekom.com.",
			ExpectedError: dnsproviders.ErrZoneNotFound,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := provider.CreateOrUpdateDNSRecord(context.Background(), tc.Zone, testRecordName, tc.Type, tc.Content, 300, false)

			if tc.ExpectedError == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Equal(t, tc.ExpectedError, err)
		})
	}

}

func TestGetRecord(t *testing.T) {
	records, err := provider.GetDNSRecords(context.Background(), testZone, testRecordName)
	require.NoError(t, err)

	require.Equal(t, 2, len(records))

	// make sure A-Record comes before AAAA-Record
	sort.Slice(records, func(i, j int) bool {
		return len(records[i].Type) < len(records[j].Type)
	})

	assert.Equal(t, testRecordName, records[0].Name)
	assert.Equal(t, api.RecordTypeA, records[0].Type)
	assert.Equal(t, 300, records[0].TTL)
	assert.Equal(t, ipv4, records[0].Content[0])

	assert.Equal(t, testRecordName, records[1].Name)
	assert.Equal(t, api.RecordTypeAAAA, records[1].Type)
	assert.Equal(t, 300, records[1].TTL)
	assert.Equal(t, ipv6, records[1].Content[0])
}

func TestDeleteRecord(t *testing.T) {
	err := provider.DeleteDNSRecord(context.Background(), testZone, testRecordName)
	require.NoError(t, err)
}

type credentialsJson struct {
	TestZone   string `json:"testZone"`
	Region     string `json:"region"`
	DomainName string `json:"domainName"`
	TenantName string `json:"tenantName"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func (c credentialsJson) ToMap() map[string]string {
	return map[string]string{
		dnsproviders.CredentialKeyRegion:     c.Region,
		dnsproviders.CredentialKeyTenantName: c.TenantName,
		dnsproviders.CredentialKeyDomainName: c.DomainName,
		dnsproviders.CredentialKeyUsername:   c.Username,
		dnsproviders.CredentialKeyPassword:   c.Password,
	}
}

func readCredentials(file string) (*credentialsJson, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var js credentialsJson
	if err := json.Unmarshal(bytes, &js); err != nil {
		return nil, err
	}

	return &js, nil
}
