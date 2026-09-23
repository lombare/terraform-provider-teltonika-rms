data "teltonika_rms_current_user" "me" {}

output "logged_in_as" {
  value = data.teltonika_rms_current_user.me.email
}
