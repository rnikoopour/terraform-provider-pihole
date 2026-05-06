---
page_title: "pihole_domain Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  Manages a Pi-hole allow or deny domain entry.
---

# pihole_domain

Manages a Pi-hole allow or deny domain entry. Entries can be exact domain matches or regular expressions.

## Example Usage

```hcl
resource "pihole_domain" "allow_exact" {
  domain = "example.test"
  type   = "allow"
  kind   = "exact"
}

resource "pihole_domain" "deny_regex" {
  domain = "^ads\\..*\\.example\\.test$"
  type   = "deny"
  kind   = "regex"
}
```

## Schema

### Required

- `domain` (String) The domain or regex pattern. Changing this forces a new resource.
- `kind` (String) Match kind: `"exact"` or `"regex"`. Changing this forces a new resource.
- `type` (String) Entry type: `"allow"` or `"deny"`. Changing this forces a new resource.

### Optional

- `comment` (String) Optional comment.
- `enabled` (Boolean) Whether the entry is enabled. Default: `true`.
- `groups` (Set of String) Group names to assign this domain entry to. Default: `["Default"]`.

### Read-Only

- `id` (String) Resource identifier in the format `type/kind/domain`.

## Import

Import using the format `type/kind/domain`:

```shell
terraform import pihole_domain.allow_exact "allow/exact/example.test"
```
