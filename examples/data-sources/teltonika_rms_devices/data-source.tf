data "teltonika_rms_devices" "field" {
  company_id = 125
  search     = "RUT"
}

output "device_count" {
  value = length(data.teltonika_rms_devices.field.devices)
}
