data "teltonika_rms_role_permissions" "all" {}

output "permission_titles" {
  value = [for p in data.teltonika_rms_role_permissions.all.permissions : p.title]
}
