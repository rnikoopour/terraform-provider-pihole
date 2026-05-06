---
page_title: "pihole_cname_record Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  Manages a Pi-hole local CNAME record.
---

# pihole_cname_record

Manages a Pi-hole local CNAME record. CNAME records allow one hostname to be an alias for another, resolved by Pi-hole's built-in DNS resolver.

## Example Usage

```hcl
resource "pihole_cname_record" "alias" {
  domain = "alias.example"
  target = "myserver.example"
}
```

## Schema

### Required

- `domain` (String) The CNAME domain (alias). Changing this forces a new resource.
- `target` (String) The target domain this CNAME points to. Changing this forces a new resource.

### Read-Only

- `id` (String) Resource identifier (same as `domain`).

## Import

Import using the CNAME domain as the ID:

```shell
terraform import pihole_cname_record.alias alias.example
```
