resource "teltonika_rms_company" "eu_subsidiary" {
  name      = "EU Field Ops"
  parent_id = 125
  email     = "eu-ops@example.com"
}

resource "teltonika_rms_tag" "field" {
  name        = "field-deployments"
  description = "Devices currently deployed in the field."
  color       = "#4287f5"
  company_id  = teltonika_rms_company.eu_subsidiary.id
}

resource "teltonika_rms_role" "field_operator" {
  title          = "Field operator"
  description    = "Can view and reboot deployed devices."
  company_ids    = [tonumber(teltonika_rms_company.eu_subsidiary.id)]
  permission_ids = [1, 2, 3]
}

resource "teltonika_rms_user_invitation" "operator" {
  email      = "alice@example.com"
  role       = "end_user"
  company_id = tonumber(teltonika_rms_company.eu_subsidiary.id)
}
