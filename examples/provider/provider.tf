terraform {
  required_providers {
    teltonika_rms = {
      source  = "registry.terraform.io/lombare/teltonika-rms"
      version = "~> 0.1"
    }
  }
}

provider "teltonika_rms" {
  # Alternatively set TELTONIKA_RMS_TOKEN in the environment.
  token = var.teltonika_rms_token
}

variable "teltonika_rms_token" {
  type      = string
  sensitive = true
}
