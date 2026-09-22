data "teltonika_rms_current_user" "me" {}

data "teltonika_rms_companies" "all" {}

data "teltonika_rms_devices" "field" {
  company_id = tonumber(data.teltonika_rms_current_user.me.company_id)
  search     = "RUT"
}

data "teltonika_rms_credits_summary" "now" {}

output "device_count" {
  value = length(data.teltonika_rms_devices.field.devices)
}

output "credits" {
  value = data.teltonika_rms_credits_summary.now.json
}
