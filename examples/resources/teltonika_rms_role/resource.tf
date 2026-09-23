resource "teltonika_rms_role" "field_operator" {
  title          = "Field operator"
  description    = "Can view and reboot deployed devices."
  company_ids    = [125]
  permission_ids = [1, 2, 3]
}
