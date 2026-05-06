resource "pihole_dns_record" "server" {
  hostname = "myserver.example"
  ip       = "192.0.2.10"
}
