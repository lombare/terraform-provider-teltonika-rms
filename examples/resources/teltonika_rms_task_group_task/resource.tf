# teltonika_rms_task_group_task is a state-only helper — it never talks to RMS by itself.
# Its purpose is to give a task definition a stable, named handle you can splice into one
# or more `teltonika_rms_task_group` resources.

resource "teltonika_rms_task_group_task" "reboot" {
  name            = "Reboot device"
  order           = 1
  type            = "command"
  timeout         = 30
  stop_on_failure = true

  data = {
    command = "reboot"
  }
}

# Reference it from a task group — the resource is treated as an object with matching
# attributes, so it plugs straight into the group's `tasks` list.
resource "teltonika_rms_task_group" "reboot_only" {
  name  = "Reboot only"
  tasks = [teltonika_rms_task_group_task.reboot]
}
