# Changelog

All notable changes to this provider are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] — 2026-09-23

Initial release. Terraform provider for the [Teltonika RMS API](https://developers.rms.teltonika-networks.com/pages/api.html),
built on the Terraform Plugin Framework.

### Provider

- Address: `registry.terraform.io/lombare/teltonika-rms`.
- Local block name: `teltonika` (Terraform disallows underscores in provider local names); resource/data-source prefix `teltonika_rms_*`.
- Authenticates with a Personal Access Token (Bearer). Configurable via `token` in
  the provider block or the `TELTONIKA_RMS_TOKEN` environment variable.
- Optional `base_url` (or `TELTONIKA_RMS_BASE_URL`) for staging or on-prem RMS
  deployments; defaults to `https://rms.teltonika-networks.com/api`.

### Resources

| Name | RMS endpoint |
| --- | --- |
| `teltonika_rms_company` | `/companies` |
| `teltonika_rms_tag` | `/tags` |
| `teltonika_rms_device_tag_assignment` | `/devices/tags/overwrite` |
| `teltonika_rms_user_invitation` | `/users/invite` |
| `teltonika_rms_role` | `/roles` |
| `teltonika_rms_alert_configuration` | `/alerts-configurations` |
| `teltonika_rms_email_configuration` | `/email-configurations` |
| `teltonika_rms_automation` | `/automations` |
| `teltonika_rms_vpn_hub` | `/vpn/hubs` |
| `teltonika_rms_vpn_hub_user` | `/vpn/hubs/users` |
| `teltonika_rms_data_collect_config` | `/data-collect/configs` |
| `teltonika_rms_configurator_template` | `/devices/configurator/templates` |
| `teltonika_rms_task_group` | `/devices/tasks/groups` (+ `/devices/tasks?group_id=…` on read) |
| `teltonika_rms_task_group_task` | *(none — state-only helper; see below)* |

All resources support `terraform import` by RMS id and reconcile drift on read.
Resources whose RMS payloads vary wildly by entity type (alert configurations,
automations, data-collect configs, configurator templates) accept a raw JSON
`payload` attribute matching the OpenAPI request body — this keeps the provider
compatible with the full RMS grammar without pinning us to one shape.

### Data sources

Singular and plural variants across the main entity domains, plus read-only
endpoints such as monitoring, statistics, hotspots, credits, and the permission
catalogue.

| Name | RMS endpoint |
| --- | --- |
| `teltonika_rms_current_user` | `/user` |
| `teltonika_rms_user`, `teltonika_rms_users` | `/users` |
| `teltonika_rms_company`, `teltonika_rms_companies` | `/companies` |
| `teltonika_rms_tag`, `teltonika_rms_tags` | `/tags` |
| `teltonika_rms_device`, `teltonika_rms_devices` | `/devices` |
| `teltonika_rms_devices_monitoring` | `/devices/monitoring` |
| `teltonika_rms_device_statistics` | `/devices/statistics` |
| `teltonika_rms_alert`, `teltonika_rms_alerts` | `/alerts` |
| `teltonika_rms_alert_configuration`, `teltonika_rms_alert_configurations` | `/alerts-configurations` |
| `teltonika_rms_email_configuration`, `teltonika_rms_email_configurations` | `/email-configurations` |
| `teltonika_rms_role`, `teltonika_rms_roles` | `/roles` |
| `teltonika_rms_role_permissions` | `/roles/permissions` |
| `teltonika_rms_automation`, `teltonika_rms_automations` | `/automations` |
| `teltonika_rms_vpn_hub`, `teltonika_rms_vpn_hubs` | `/vpn/hubs` |
| `teltonika_rms_credits_summary` | `/credits/summary` |
| `teltonika_rms_files` | `/files` |
| `teltonika_rms_hotspots` | `/hotspots` |
| `teltonika_rms_data_collect_configs` | `/data-collect/configs` |
| `teltonika_rms_task_group`, `teltonika_rms_task_groups` | `/devices/tasks/groups` (+ nested tasks via `/devices/tasks?group_id=…`) |

`teltonika_rms_device` and `teltonika_rms_devices_monitoring` expose the full
RMS response as a `raw` JSON attribute in addition to the typed fields, so
model-specific attributes that vary between routers remain accessible.

### Install

```sh
git clone https://github.com/lombare/Terraform-Provider-Teltonika-RMS.git
cd Terraform-Provider-Teltonika-RMS
make install
```

`make install` drops the plugin under
`~/.terraform.d/plugins/registry.terraform.io/lombare/teltonika-rms/1.0.0/<os_arch>/`
so Terraform resolves the source `registry.terraform.io/lombare/teltonika-rms`
locally without any registry publication.

### Scope & known limitations

- Covers the management-shaped surface of the RMS API — everything creatable,
  updatable, or deletable is a resource; useful read endpoints are data sources.
  Endpoints that model **actions** rather than state (device commands, remote
  access sessions, credit moves, report generation, quick-hub bootstrapping,
  configurator apply, wireless/data-usage detail, tasks, images) are not yet
  wired; the pattern for adding them is documented in the README.
- `teltonika_rms_user_invitation` and `teltonika_rms_device_tag_assignment`
  trust local state on read: RMS does not expose a get-single endpoint for
  invitations, and no single endpoint returns just the tag set for a device.
  Drift is corrected on the next apply rather than surfaced in `terraform plan`.
- `teltonika_rms_configurator_template` requires replacement on payload change —
  RMS has no update verb for templates.
- `teltonika_rms_task_group_task` is a **state-only helper resource** — it
  performs no RMS API calls at any lifecycle stage. Tasks are created by the
  RMS API only as a side effect of the task-group POST/PUT, so a task record
  only makes sense when spliced into a group's `tasks` list. The pattern lets
  users name and reuse task definitions; inline objects work too.

### Security

- All dependencies are on versions that patch the currently disclosed advisories
  (`golang.org/x/net` v0.59.0, `google.golang.org/grpc` v1.84.0).
- Sensitive attributes (`token`, `password`) are marked so Terraform redacts them
  from plan output and state diffs.

### Compatibility

- Terraform ≥ 1.0.
- Go 1.24+ to build from source.
- Tested against RMS API v3 (BETA), OpenAPI dated the release date above.
