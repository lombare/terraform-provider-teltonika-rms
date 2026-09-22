resource "teltonika_rms_vpn_hub" "eu_central" {
  name        = "eu-central"
  description = "Primary EU hub, TAP topology"
  hub_zone    = "frankfurt-1"
  vpn_type    = "tap"
  enabled     = true
}

resource "teltonika_rms_vpn_hub_user" "operator" {
  name    = "alice"
  hub_id  = tonumber(teltonika_rms_vpn_hub.eu_central.id)
  enabled = true
}
