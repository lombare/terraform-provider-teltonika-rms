terraform {
  required_providers {
    teltonika = {
      source  = "registry.terraform.io/lombare/teltonika-rms"
      version = "~> 1.0"
    }
  }
}

provider "teltonika" {
  # Alternatively set TELTONIKA_RMS_TOKEN in the environment.
  token = var.teltonika_rms_token
}

variable "teltonika_rms_token" {
  type      = string
  sensitive = true
}
