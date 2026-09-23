resource "teltonika_rms_alert_configuration" "device_offline" {
  # Payload mirrors the OpenAPI `alert_config_add` shape.
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
