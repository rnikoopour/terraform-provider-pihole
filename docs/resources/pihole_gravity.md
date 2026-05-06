---
page_title: "pihole_gravity Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  Triggers a Pi-hole gravity update.
---

# pihole_gravity

Triggers a Pi-hole gravity update, which downloads all configured block and allow lists and rebuilds the gravity database. Use `replace_triggered_by` in the `lifecycle` block to re-run gravity automatically when lists or domains change.

Gravity is not "deleted" when this resource is destroyed — destruction only removes it from Terraform state.

## Example Usage

```hcl
resource "pihole_list" "blocklist" {
  address = "https://lists.example/blocklist.txt"
  type    = "block"
}

resource "pihole_gravity" "update" {
  lifecycle {
    replace_triggered_by = [
      pihole_list.blocklist,
    ]
  }
}
```

## Schema

### Read-Only

- `id` (String) Resource identifier (always `"gravity"`).

## Import

Import using the literal string `gravity`:

```shell
terraform import pihole_gravity.update gravity
```
