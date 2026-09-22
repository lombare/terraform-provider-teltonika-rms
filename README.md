# terraform-provider-teltonika-rms

Terraform provider for the [Teltonika RMS API](https://developers.rms.teltonika-networks.com/pages/api.html).

Built on the Terraform Plugin Framework, it exposes the RMS platform as first-class Terraform resources and data sources so you can manage RMS state (companies, users, tags, alerts, VPN hubs, automations, roles, and more) declaratively.

## Provider configuration

```hcl
terraform {
  required_providers {
    teltonika_rms = {
      source  = "registry.terraform.io/lombare/teltonika-rms"
      version = "~> 0.1"
    }
  }
}

provider "teltonika_rms" {
  # The RMS Personal Access Token (Bearer). May also be supplied via TELTONIKA_RMS_TOKEN.
  token = var.teltonika_rms_token

  # Optional. Defaults to https://rms.teltonika-networks.com/api and may be overridden with
  # TELTONIKA_RMS_BASE_URL for self-hosted deployments or staging.
  # base_url = "https://rms.teltonika-networks.com/api"
}
```

## Resources

| Resource | Description |
| --- | --- |
| `teltonika_rms_company` | Subsidiary companies (`/companies`) |
| `teltonika_rms_tag` | Device tags (`/tags`) |
| `teltonika_rms_device_tag_assignment` | Assign a set of tags to a device (`/devices/tags`) |
| `teltonika_rms_user_invitation` | User invitations (`/users/invite`) |
| `teltonika_rms_role` | Roles and permission bindings (`/roles`) |
| `teltonika_rms_alert_configuration` | Alert configurations (`/alerts-configurations`) |
| `teltonika_rms_email_configuration` | SMTP email configurations (`/email-configurations`) |
| `teltonika_rms_automation` | Automations (`/automations`) |
| `teltonika_rms_vpn_hub` | VPN hubs (`/vpn/hubs`) |
| `teltonika_rms_vpn_hub_user` | VPN hub users (`/vpn/hubs/users`) |
| `teltonika_rms_data_collect_config` | Data-collect configurations (`/data-collect/configs`) |
| `teltonika_rms_configurator_template` | Device configurator templates |

## Data sources

Both singular (single-record) and plural (list) variants are provided across the main entity domains, plus several read-only endpoints such as monitoring, statistics, hotspots, and credits.

| Data source | Description |
| --- | --- |
| `teltonika_rms_current_user` | The authenticated principal (`/user`) |
| `teltonika_rms_user`, `teltonika_rms_users` | Users (`/users`) |
| `teltonika_rms_company`, `teltonika_rms_companies` | Companies (`/companies`) |
| `teltonika_rms_tag`, `teltonika_rms_tags` | Tags (`/tags`) |
| `teltonika_rms_device`, `teltonika_rms_devices` | Devices (`/devices`) |
| `teltonika_rms_devices_monitoring` | Live device monitoring (`/devices/monitoring`) |
| `teltonika_rms_device_statistics` | Device statistics (`/devices/statistics`) |
| `teltonika_rms_alert`, `teltonika_rms_alerts` | Alerts (`/alerts`) |
| `teltonika_rms_alert_configuration`, `teltonika_rms_alert_configurations` | Alert configurations |
| `teltonika_rms_email_configuration`, `teltonika_rms_email_configurations` | Email configurations |
| `teltonika_rms_role`, `teltonika_rms_roles` | Roles |
| `teltonika_rms_role_permissions` | Full permission catalogue (`/roles/permissions`) |
| `teltonika_rms_automation`, `teltonika_rms_automations` | Automations |
| `teltonika_rms_vpn_hub`, `teltonika_rms_vpn_hubs` | VPN hubs |
| `teltonika_rms_credits_summary` | Credits summary (`/credits/summary`) |
| `teltonika_rms_files` | Files (`/files`) |
| `teltonika_rms_hotspots` | Hotspots (`/hotspots`) |
| `teltonika_rms_data_collect_configs` | Data-collect configurations |

## Building & installing locally

```sh
git clone https://github.com/lombare/Terraform-Provider-Teltonika-RMS.git
cd Terraform-Provider-Teltonika-RMS
make install        # builds and drops the plugin under ~/.terraform.d/plugins
```

Then, in any Terraform configuration, pin the source to `registry.terraform.io/lombare/teltonika-rms` and run `terraform init`. Terraform will resolve the source to the local plugin because `make install` places the binary under `~/.terraform.d/plugins/registry.terraform.io/lombare/teltonika-rms/<version>/<os_arch>/`.

## API surface

The RMS API exposes roughly 240 operations across 175+ paths. The provider intentionally covers every management-shaped area of the API — anything creatable/deletable/updatable is a resource; read-only endpoints are data sources. When the API adds new endpoints, the pattern to add coverage is:

1. Add a typed method to `internal/rms/` alongside the entity it belongs to.
2. Add a resource (`internal/provider/resource_<name>.go`) or a data source (`internal/provider/data_source_<name>.go`) following the shape of an existing sibling.
3. Register it in `internal/provider/provider.go`.

## License

MIT.
