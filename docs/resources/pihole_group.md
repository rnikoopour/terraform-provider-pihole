---
page_title: "pihole_group Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  Manages a Pi-hole group.
---

# pihole_group

Manages a Pi-hole group. Groups are used to organize lists and domains and apply them selectively to clients.

## Example Usage

```hcl
resource "pihole_group" "iot" {
  name    = "IoT"
  comment = "IoT device group"
  enabled = true
}
```

## Schema

### Required

- `name` (String) Group name. Changing this forces a new resource.

### Optional

- `comment` (String) Optional comment.
- `enabled` (Boolean) Whether the group is enabled. Default: `true`.

### Read-Only

- `id` (String) Resource identifier (same as `name`).

## Import

Import using the group name as the ID:

```shell
terraform import pihole_group.iot IoT
```
