resource "pihole_group" "server" {
  name    = "server"
  comment = "Allowlist scope for one host"
}

resource "pihole_client" "server" {
  client  = "192.0.2.10"
  comment = "Example host"
  groups  = ["Default", pihole_group.server.name]
}
