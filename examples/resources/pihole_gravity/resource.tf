resource "pihole_gravity" "update" {
  lifecycle {
    replace_triggered_by = [
      pihole_list.ads,
    ]
  }
}
