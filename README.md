# terraform-provider-pihole

A Terraform provider for managing Pi-hole v6 configuration.

## Resources

- `pihole_config` — DNS and webserver settings
- `pihole_list` — Block/allow lists
- `pihole_domain` — Block/allow domains (exact or regex)
- `pihole_dns_record` — Custom local DNS A/AAAA records
- `pihole_cname_record` — Custom local CNAME records
- `pihole_gravity` — Triggers a gravity update; use `replace_triggered_by` to run it after list/domain changes
- `pihole_group` — Groups for organizing lists and domains

## Configuration notes

### `webserver.api.app_sudo`

The `app_sudo` setting controls whether app-password API sessions are allowed to modify Pi-hole configuration. **This setting cannot be managed by this provider** because of a bootstrapping problem: the provider authenticates using an app password, so if `app_sudo` is `false`, the API rejects all config-change requests — including the one that would set `app_sudo = true`.

**Recommended approach:** Set `app_sudo` outside of Terraform, either via an environment variable in your container configuration or by setting it once manually through the Pi-hole web UI.

For Docker deployments:

```yaml
environment:
  FTLCONF_webserver_api_password: "your-password"
  FTLCONF_webserver_api_app_sudo: "true"
```

Once `app_sudo` is enabled, all other `pihole_config` settings can be managed by this provider.

### App passwords and 2FA

If 2FA (TOTP) is enabled on your Pi-hole, the provider must authenticate using an app password rather than the main web password. Generate one in the Pi-hole web UI under Settings → API, then use it as the `password` provider argument.

### `pihole_config` and FTL restarts

Applying `pihole_config` causes Pi-hole's FTL process to restart when settings change. The provider retries authentication automatically during the restart window, but resources that depend on `pihole_config` should declare that dependency so Terraform applies them after the config settles.

### `webserver.domain`

Pi-hole uses `webserver.domain` to determine the hostname it redirects to. If this is not set correctly, the web UI and API will redirect to the container's IP address instead of the DNS name, causing API calls to fail.

**Set this before first use** via the Docker environment variable:

```yaml
environment:
  FTLCONF_webserver_domain: "pihole.example"
```
