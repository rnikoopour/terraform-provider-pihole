resource "pihole_cname_record" "alias" {
  domain = "alias.example"
  target = "myserver.example"
}
