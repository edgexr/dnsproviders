package otc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexr/dnsproviders"
)

const (
	credentialFile = "config.json"
)

var (
	provider          *dnsproviders.OTC
	testZone          = ""
	testRecordName    = fmt.Sprintf("test-%s", strings.Split(uuid.NewString(), "-")[0])
	testRecordContent = uuid.NewString()
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
		ExpectedError error
	}

	for name, tc := range map[string]testCase{
		"no error": {
			Zone:          testZone,
			ExpectedError: nil,
		},
		"invalid zone": {
			Zone:          "non-existing.edge.telekom.com.",
			ExpectedError: dnsproviders.ErrZoneNotFound,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := provider.CreateOrUpdateDNSRecord(context.Background(), tc.Zone, testRecordName, "TXT", testRecordContent, 300, false)

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

	require.Equal(t, 1, len(records))

	record := records[0]
	assert.Equal(t, testRecordName, record.Name)
	assert.Equal(t, "TXT", record.Type)
	assert.Equal(t, 300, record.TTL)
	assert.Equal(t, testRecordContent, record.Content[0])
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
