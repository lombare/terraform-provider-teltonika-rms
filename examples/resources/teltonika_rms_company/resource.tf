resource "teltonika_rms_company" "eu_subsidiary" {
  name      = "EU Field Ops"
  parent_id = 125
  email     = "eu-ops@example.com"
}
