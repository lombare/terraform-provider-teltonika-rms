resource "teltonika_rms_email_configuration" "ops_smtp" {
  name     = "ops-smtp"
  host     = "smtp.example.com"
  port     = 587
  email    = "alerts@example.com"
  username = "alerts@example.com"
  password = var.smtp_password
}

variable "smtp_password" {
  type      = string
  sensitive = true
}
