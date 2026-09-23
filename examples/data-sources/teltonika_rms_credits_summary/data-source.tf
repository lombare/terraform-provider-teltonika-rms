data "teltonika_rms_credits_summary" "now" {}

output "credits" {
  value = data.teltonika_rms_credits_summary.now.json
}
