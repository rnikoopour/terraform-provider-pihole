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
