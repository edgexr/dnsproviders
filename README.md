# DNS Providers for Edge Cloud

DNS Providers for Edge Cloud provide a generic interface for a DNS
provider for the [Edge Cloud Platform](https://github.com/edgexr/edge-cloud-platform).
A DNS provider allows the platform to list, get, create, and delete DNS
records in a given DNS provider.

The list of currently available DNS Providers:
- Cloudflare
- Google Cloud DNS
- Open Telekom Cloud

## Adding a New Provider

New providers must implement the DNSProvider interface in (api/api.go).

New providers should have a unit test that uses the ProviderTest in (dnsapi_test.go).

## Testing

Each provider requires specific credentials, so in general there are no general
tests to run. To test a specific provider, prepare and set the credentials for
the provider and then run the specific unit test(s) for that provider.
