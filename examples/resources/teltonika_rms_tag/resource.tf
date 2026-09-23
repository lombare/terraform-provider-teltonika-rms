resource "teltonika_rms_tag" "field" {
  name        = "field-deployments"
  description = "Devices currently deployed in the field."
  color       = "#4287f5"
  company_id  = 125
}
