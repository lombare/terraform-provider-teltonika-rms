data "teltonika_rms_task_groups" "all" {}

output "group_count" {
  value = length(data.teltonika_rms_task_groups.all.task_groups)
}
