resource "teltonika_rms_vpn_hub_user" "operator" {
  name    = "alice"
  hub_id  = 7
  enabled = true
}
