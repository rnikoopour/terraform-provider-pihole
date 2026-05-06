terraform {
  required_providers {
    pihole = {
      source  = "rnikoopour/pihole"
      version = "~> 0.1"
    }
  }
}

provider "pihole" {
  url      = "https://192.0.2.1"
  password = var.pihole_password
}
