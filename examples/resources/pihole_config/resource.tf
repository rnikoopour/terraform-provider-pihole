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
