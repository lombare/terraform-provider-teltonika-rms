resource "teltonika_rms_user_invitation" "operator" {
  email      = "alice@example.com"
  role       = "end_user"
  company_id = 125
}
