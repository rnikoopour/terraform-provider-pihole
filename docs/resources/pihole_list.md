---
page_title: "pihole_list Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  Manages a Pi-hole block or allow list.
---

# pihole_list

Manages a Pi-hole block or allow list. After adding or removing lists, run a gravity update via [`pihole_gravity`](pihole_gravity.md) to apply the changes.

## Example Usage

```hcl
resource "pihole_list" "blocklist" {
  address = "https://lists.example/blocklist.txt"
  type    = "block"
  comment = "Primary block list"
}

resource "pihole_list" "allowlist" {
  address = "https://lists.example/allowlist.txt"
  type    = "allow"
  groups  = ["Default"]
}
```

## Schema

### Required

- `address` (String) URL of the list. Changing this forces a new resource.
- `type` (String) List type: `"block"` or `"allow"`. Changing this forces a new resource.

### Optional

- `comment` (String) Optional comment.
- `enabled` (Boolean) Whether the list is enabled. Default: `true`.
- `groups` (Set of String) Group names to assign this list to. Default: `["Default"]`.

### Read-Only

- `id` (String) Resource identifier (same as `address`).

## Import

Import using the list URL as the ID:

```shell
terraform import pihole_list.blocklist "https://lists.example/blocklist.txt"
```
