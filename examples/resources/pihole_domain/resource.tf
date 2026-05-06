resource "pihole_domain" "allow_exact" {
  domain = "example.com"
  type   = "allow"
  kind   = "exact"
}

resource "pihole_domain" "deny_regex" {
  domain = "^ads\\..*\\.example\\.com$"
  type   = "deny"
  kind   = "regex"
}
