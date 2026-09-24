data "teltonika_rms_task_group" "zero_touch" {
  id = "41"
}

output "task_names" {
  value = [for t in data.teltonika_rms_task_group.zero_touch.tasks : t.name]
}
