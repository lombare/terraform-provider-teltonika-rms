data "teltonika_rms_devices_monitoring" "live" {}

output "online_devices" {
  value = [for d in data.teltonika_rms_devices_monitoring.live.devices : d.name if d.status == "online"]
}
