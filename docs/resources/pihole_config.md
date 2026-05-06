---
page_title: "pihole_config Resource - terraform-provider-pihole"
subcategory: ""
description: |-
  Manages Pi-hole DNS, webserver, and system configuration settings.
---

# pihole_config

Manages Pi-hole DNS, webserver, and system configuration settings. Applying this resource causes Pi-hole's FTL process to restart when settings change. Resources that depend on `pihole_config` should declare that dependency so Terraform applies them after the restart settles.

~> **One resource per Pi-hole.** Only one `pihole_config` resource should exist per provider instance.

~> **`app_sudo` required.** The Pi-hole API rejects configuration changes from app-password sessions unless `webserver.api.app_sudo` is enabled. This setting cannot be managed by the provider itself because of a bootstrapping problem: the provider authenticates using an app password, so if `app_sudo` is false, the API will reject all config-change requests. Enable it outside of Terraform before first use (see the [provider docs](../index.md) for details).

## Example Usage

```hcl
resource "pihole_config" "main" {
  dns = {
    upstreams     = ["8.8.8.8", "8.8.4.4"]
    dnssec        = true
    query_logging = true

    blocking = {
      active = true
      mode   = "NULL"
    }

    rate_limit = {
      count    = 1000
      interval = 60
    }
  }

  webserver = {
    domain = "pihole.example"
  }

  misc = {
    privacy_level = 0
  }

  database = {
    max_db_days = 90
  }
}
```

## Schema

### Required

- `dns` (Attributes) DNS configuration. (see [below for nested schema](#nestedatt--dns))

### Optional

- `database` (Attributes) Database settings. (see [below for nested schema](#nestedatt--database))
- `misc` (Attributes) Miscellaneous settings. (see [below for nested schema](#nestedatt--misc))
- `webserver` (Attributes) Web interface settings. (see [below for nested schema](#nestedatt--webserver))

### Read-Only

- `id` (String) Resource identifier.

---

<a id="nestedatt--dns"></a>
### Nested Schema for `dns`

**Required:**

- `upstreams` (List of String) Upstream DNS server addresses.

**Optional:**

- `block_esni` (Boolean) Block ESNI (Encrypted Server Name Indication). Default: `true`.
- `block_ttl` (Number) TTL in seconds for blocked DNS responses. Default: `2`.
- `blocking` (Attributes) DNS blocking settings. (see [below for nested schema](#nestedatt--dns--blocking))
- `bogus_priv` (Boolean) Do not forward reverse lookups for private IP ranges to upstream. Default: `true`.
- `cache` (Attributes) DNS cache settings. (see [below for nested schema](#nestedatt--dns--cache))
- `cname_deep_inspect` (Boolean) Follow CNAME chains when checking blocklists. Default: `true`.
- `dnssec` (Boolean) Enable DNSSEC validation. Default: `false`.
- `domain_needed` (Boolean) Do not forward incomplete domain names to upstream DNS. Default: `true`.
- `expand_hosts` (Boolean) Add the local domain to simple hostnames in `/etc/hosts`. Default: `true`.
- `host_record` (String) Custom host record for this Pi-hole instance, e.g. `"pihole.example,192.0.2.1"`. Default: `""`.
- `interface` (String) Network interface to listen on. Used with `listening_mode` `"SINGLE"` or `"BIND"`. Default: `""`.
- `listening_mode` (String) Interface listening mode. Valid values: `"LOCAL"`, `"SINGLE"`, `"BIND"`, `"ALL"`, `"NONE"`. Default: `"LOCAL"`.
- `query_logging` (Boolean) Log DNS queries. Default: `true`.
- `rate_limit` (Attributes) Query rate-limit settings. (see [below for nested schema](#nestedatt--dns--rate_limit))
- `reply_when_busy` (String) How to handle queries when the gravity database is busy. Valid values: `"ALLOW"`, `"BLOCK"`, `"REFUSE"`, `"DROP"`. Default: `"ALLOW"`.
- `special_domains` (Attributes) Special domain blocking settings. (see [below for nested schema](#nestedatt--dns--special_domains))

<a id="nestedatt--dns--blocking"></a>
### Nested Schema for `dns.blocking`

- `active` (Boolean) Whether DNS blocking is active. Default: `true`.
- `edns` (String) EDNS info in blocked replies. Valid values: `"NONE"`, `"CODE"`, `"TEXT"`. Default: `"TEXT"`.
- `mode` (String) Blocking mode. Valid values: `"NULL"`, `"NXDOMAIN"`, `"NODATA"`, `"IP"`, `"IP-NODATA-AAAA"`. Default: `"NULL"`.

<a id="nestedatt--dns--cache"></a>
### Nested Schema for `dns.cache`

- `optimizer` (Number) Cache optimizer TTL in seconds. Default: `3600`.
- `size` (Number) DNS cache size (number of entries). Default: `10000`.

<a id="nestedatt--dns--rate_limit"></a>
### Nested Schema for `dns.rate_limit`

- `count` (Number) Number of queries allowed per interval before rate-limiting a client. Default: `1000`.
- `interval` (Number) Rate-limit interval in seconds. Default: `60`.

<a id="nestedatt--dns--special_domains"></a>
### Nested Schema for `dns.special_domains`

- `designated_resolver` (Boolean) Block DNS Designated Resolver records. Default: `true`.
- `icloud_private_relay` (Boolean) Block Apple iCloud Private Relay. Default: `true`.
- `mozilla_canary` (Boolean) Block Mozilla's canary domain to disable DNS-over-HTTPS in Firefox. Default: `true`.

---

<a id="nestedatt--webserver"></a>
### Nested Schema for `webserver`

**Optional:**

- `domain` (String) Hostname the Pi-hole web server redirects to. Pi-hole uses this to determine where to redirect API and UI requests — set it to the DNS name you use to reach your Pi-hole. Default: `"pi.hole"`.
- `interface` (Attributes) Web interface appearance settings. (see [below for nested schema](#nestedatt--webserver--interface))

<a id="nestedatt--webserver--interface"></a>
### Nested Schema for `webserver.interface`

- `boxed` (Boolean) Use boxed layout for the web interface. Default: `true`.
- `theme` (String) Web interface theme. Valid values: `"default-auto"`, `"default-light"`, `"default-dark"`, `"default-darker"`, `"high-contrast"`, `"high-contrast-dark"`, `"lcars"`. Default: `"default-auto"`.

---

<a id="nestedatt--misc"></a>
### Nested Schema for `misc`

**Optional:**

- `privacy_level` (Number) Privacy level for statistics. `0` = full detail, `1` = hide domains, `2` = hide domains and clients, `3` = anonymous. Default: `0`.

---

<a id="nestedatt--database"></a>
### Nested Schema for `database`

**Optional:**

- `max_db_days` (Number) How many days to retain queries in the database. Set to `0` to disable the database. Default: `91`.

## Import

Import using the literal string `config`:

```shell
terraform import pihole_config.main config
```
