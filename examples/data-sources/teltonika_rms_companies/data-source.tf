data "teltonika_rms_companies" "all" {}

output "company_names" {
  value = [for c in data.teltonika_rms_companies.all.companies : c.name]
}
