terraform {
  required_version = ">= 1.5.0"
  required_providers {
    datadog = { source = "DataDog/datadog", version = "~> 4.0" }
  }
}
provider "datadog" {}
variable "accounts" {
  type = map(object({
    account_tag         = string
    login_metric        = string
    failed_login_metric = string
    credit_metric       = string
  }))
}
variable "sre_role_ids" {
  type        = list(string)
  default     = []
  description = "Datadog role IDs used for dashboard editing restrictions"
}
locals {
  monitor_specs = merge([
    for account, metrics in var.accounts : {
      for kind, spec in {
        logins        = { metric = metrics.login_metric, threshold = 1000 }
        failed_logins = { metric = metrics.failed_login_metric, threshold = 5 }
        credits       = { metric = metrics.credit_metric, threshold = 10 }
        } : "${account}_${kind}" => merge(spec, {
          account     = account
          account_tag = metrics.account_tag
          kind        = kind
      })
      if spec.metric != ""
    }
  ]...)
}
resource "datadog_monitor" "snowflake" {
  for_each = local.monitor_specs
  name     = "[SF-${upper(each.value.account)}] ${replace(each.value.kind, "_", " ")}"
  type     = "query alert"
  query    = "sum(last_1h):sum:${each.value.metric}{snowflake_account:${each.value.account_tag}}.as_count() > ${each.value.threshold}"
  message  = "Snowflake ${each.value.kind} threshold exceeded for ${each.value.account}. Investigate account activity."
  tags     = ["service:snowflake", "snowflake_account:${each.value.account}", "team:sre"]
  monitor_thresholds { critical = each.value.threshold }
  notify_no_data      = false
  require_full_window = false
}
resource "datadog_dashboard" "snowflake" {
  title       = "Enterprise Snowflake Monitoring"
  description = "Multi-account Snowflake monitor overview, provisioned by Terraform."
  layout_type = "ordered"
  dynamic "widget" {
    for_each = datadog_monitor.snowflake
    content {
      alert_value_definition {
        alert_id = widget.value.id
        title    = widget.value.name
      }
    }
  }
  # Role restrictions require validated Datadog role UUIDs and permissions.
  restricted_roles = length(var.sre_role_ids) > 0 ? var.sre_role_ids : null
}
output "dashboard_url" { value = datadog_dashboard.snowflake.url }
