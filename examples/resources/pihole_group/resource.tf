resource "pihole_group" "iot" {
  name    = "IoT"
  comment = "IoT device group"
  enabled = true
}
