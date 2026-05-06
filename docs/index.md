---
page_title: "Provider: Pi-hole"
description: |-
  Use the Pi-hole provider to manage Pi-hole v6 resources via its REST API.
---

# Pi-hole Provider

Use the Pi-hole provider to manage [Pi-hole v6](https://pi-hole.net/) resources via its REST API. The provider supports managing DNS records, block/allow lists, domains, groups, and Pi-hole server configuration.

~> **Pi-hole v6 only.** This provider targets the Pi-hole v6 REST API and is not compatible with Pi-hole v5.

## Authentication

The provider authenticates using a Pi-hole web password or app password.

If 2FA (TOTP) is enabled on your Pi-hole, you must use an **app password** instead of the main web password. Generate one in the Pi-hole web UI under **Settings → API**.

### app_sudo

Pi-hole's `webserver.api.app_sudo` setting controls whether app-password sessions can modify configuration. This setting cannot be managed by the provider itself (see [`pihole_config`](resources/pihole_config) for details). Enable it outside of Terraform — via a Docker environment variable or the Pi-hole web UI — before first use:

```yaml
environment:
  FTLCONF_webserver_api_app_sudo: "true"
```

## Example Usage

```hcl
terraform {
  required_providers {
    pihole = {
      source  = "rnikoopour/pihole"
      version = "~> 0.1"
    }
  }
}

provider "pihole" {
  url      = "https://192.0.2.1"
  password = var.pihole_password
}
```

## Schema

### Required

- `url` (String) Base URL of the Pi-hole server (e.g. `https://192.0.2.1`).
- `password` (String, Sensitive) Pi-hole web password or app password (if 2FA is enabled).

### Optional

- `insecure` (Boolean) Skip TLS certificate verification. Useful for self-signed certificates.
