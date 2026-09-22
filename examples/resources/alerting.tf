resource "teltonika_rms_email_configuration" "ops_smtp" {
  name     = "ops-smtp"
  host     = "smtp.example.com"
  port     = 587
  email    = "alerts@example.com"
  username = "alerts@example.com"
  password = var.smtp_password
}

resource "teltonika_rms_alert_configuration" "device_offline" {
  # Payload mirrors the OpenAPI `alert_config_add` shape. See:
  # https://developers.rms.teltonika-networks.com/pages/api.html
  payload = jsonencode({
    name       = "device-offline-15m"
    type       = "device_offline"
    company_id = 125
    enabled    = true
    conditions = {
      period_seconds = 900
    }
    actions = [
      { email_configuration_id = 1 },
    ]
  })
}

variable "smtp_password" {
  type      = string
  sensitive = true
}
