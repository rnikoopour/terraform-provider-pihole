---
page_title: "pihole_client Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  Manages a Pi-hole client record for group-scoped filtering.
---

# pihole_client

Manages a Pi-hole client record. Assigning a client to a non-default group is what makes group-scoped allow/deny domains apply to just that client instead of the whole instance.

Clients with no explicit record belong to the `Default` group only, so they receive instance-wide blocking.

## Example Usage

```hcl
resource "pihole_group" "server" {
  name    = "server"
  comment = "Allowlist scope for one host"
}

resource "pihole_client" "server" {
  client  = "192.0.2.10"
  comment = "Example host"
  # Keeping "Default" means normal ad blocking still applies to this client;
  # the extra group only widens what its scoped allow entries can resolve.
  groups = ["Default", pihole_group.server.name]
}
```

## Schema

### Required

- `client` (String) Client identifier: IP address, CIDR, MAC address, hostname, or interface. Changing this forces a new resource.

### Optional

- `comment` (String) Optional comment.
- `groups` (Set of String) Group names this client belongs to. Include `Default` to keep instance-wide ad blocking in effect. Default: `["Default"]`.

### Read-Only

- `id` (String) Resource identifier (same as `client`).

## Import

Import using the client identifier as the ID:

```shell
terraform import pihole_client.server 192.0.2.10
```
