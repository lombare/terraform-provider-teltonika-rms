data "teltonika_rms_current_user" "me" {}

resource "teltonika_rms_tag" "zero_touch_ok" {
  name        = "zero-touch-ok"
  description = "Devices that have completed zero-touch provisioning."
  color       = "#4287f5"
  company_id  = data.teltonika_rms_current_user.me.company_id
}

resource "teltonika_rms_tag" "zero_touch_failed" {
  name        = "zero-touch-failed"
  description = "Devices where zero-touch provisioning failed at some step."
  color       = "#e53935"
  company_id  = data.teltonika_rms_current_user.me.company_id
}

# Two reusable task definitions declared as state-only helpers. Neither of these
# resources talks to RMS — they hold typed task records so the group below can
# splice them into its `tasks` list. See the resource docs page.

resource "teltonika_rms_task_group_task" "update" {
  name            = "Update firmware"
  order           = 1
  type            = "command"
  timeout         = 120
  stop_on_failure = true

  data = {
    command = "echo 'pretend firmware update'"
  }
}

resource "teltonika_rms_task_group_task" "configure" {
  name            = "Push configuration"
  order           = 2
  type            = "command"
  timeout         = 60
  stop_on_failure = true

  data = {
    command          = file("${path.module}/configure.sh")
    acceptable_codes = [0, 2]
  }
}

resource "teltonika_rms_task_group" "zero_touch" {
  name       = "ZeroTouch"
  company_id = data.teltonika_rms_current_user.me.company_id

  success_tag_ids = [teltonika_rms_tag.zero_touch_ok.id]
  failed_tag_ids  = [teltonika_rms_tag.zero_touch_failed.id]

  # Terraform accepts either a resource reference (each state-only record above
  # is treated as an object with matching attributes) or an inline literal.
  tasks = [
    teltonika_rms_task_group_task.update,
    teltonika_rms_task_group_task.configure,
  ]
}
