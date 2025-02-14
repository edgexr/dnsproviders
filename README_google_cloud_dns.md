# Credentials for Google Cloud DNS

The Google Cloud DNS provider uses the REST API to
communicate with Google Cloud, so any application credentials
you have will suffice. However, for security, it is
recommended to create a service account with limited
permissions.

### 1. Create a Service Account

Follow the instructions to [create a service account](https://cloud.google.com/iam/docs/service-accounts-create).
When creating the service account, give the service account
the "DNS Administrator" role.

### 2. Generate a Key for the Service Account

Follow the instructions to [generate a service account key](https://cloud.google.com/iam/docs/keys-create-delete).

You will get a JSON file that looks like:
```json
{
  "type": "service_account",
  "project_id": "PROJECT_ID",
  "private_key_id": "KEY_ID",
  "private_key": "-----BEGIN PRIVATE KEY-----\nPRIVATE_KEY\n-----END PRIVATE KEY-----\n",
  "client_email": "SERVICE_ACCOUNT_EMAIL",
  "client_id": "CLIENT_ID",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/SERVICE_ACCOUNT_EMAIL"
}
```

These are your application default credentials. They can be
parsed into a `map[string]string` and passed in as the
`credentialsData` parameter to `NewGoogleCloudDNSProvider()`.
