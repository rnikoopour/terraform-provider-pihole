---
page_title: "pihole_dns_record Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  Manages a Pi-hole custom local DNS A/AAAA record.
---

# pihole_dns_record

Manages a Pi-hole custom local DNS A or AAAA record. These records are served by Pi-hole's built-in DNS resolver for local name resolution.

## Example Usage

```hcl
resource "pihole_dns_record" "server" {
  hostname = "myserver.example"
  ip       = "192.0.2.10"
}
```

## Schema

### Required

- `hostname` (String) Hostname to resolve to the IP address. Changing this forces a new resource.
- `ip` (String) IP address (IPv4 or IPv6). Changing this forces a new resource.

### Read-Only

- `id` (String) Resource identifier in the format `ip hostname`.

## Import

Import using the format `ip hostname` (space-separated):

```shell
terraform import pihole_dns_record.server "192.0.2.10 myserver.example"
```
