# Credentials for Cloudflare

The Cloudflare DNS provider requires an API Key.
Follow the [cloudflare instructions](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/)
to create an API Key.

It is recommended to use the `Edit Zone` template.
Choose permissions as `Zone - DNS - Edit`, and
Zone Resources as `Include - Specific Zone - <your zone>`.

You will get a single string API Key. Set this as the token
value in the credentials map passed to `NewCloudflareProvider()`.

```
credentialsData := map[string]string{
    "token": "<api key>"
}
```
