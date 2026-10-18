
# Snowflake resources for Datadog monitoring.
# Add the Snowflake provider to the existing required_providers block
# in main.tf; do not create a second terraform {} required_providers block.


locals {
  datadog_public_key = replace(
    replace(
      replace(
        replace(
          trimspace(file("${path.module}/../secrets/datadog_rsa_key.pub")),
          "-----BEGIN PUBLIC KEY-----",
          ""
        ),
        "-----END PUBLIC KEY-----",
        ""
      ),
      "\n",
      ""
    ),
    "\r",
    ""
  )
}


variable "snowflake_organization" {
  type = string
}

variable "snowflake_account" {
  type = string
}

variable "snowflake_admin_user" {
  type = string
}

variable "snowflake_admin_private_key_path" {
  type = string
}

provider "snowflake" {
  organization_name = var.snowflake_organization
  account_name      = var.snowflake_account
  user              = var.snowflake_admin_user
  role              = "ACCOUNTADMIN"
  authenticator     = "SNOWFLAKE_JWT"
  private_key       = file(var.snowflake_admin_private_key_path)
}

resource "snowflake_account_role" "datadog" {
  name    = "DATADOG_ROLE"
  comment = "Dedicated read-only Snowflake monitoring role"
}

resource "snowflake_warehouse" "datadog" {
  name                = "DATADOG_WH"
  warehouse_size      = "XSMALL"
  auto_suspend        = 60
  auto_resume         = true
  initially_suspended = true
  comment             = "Warehouse for Datadog monitoring queries"
}

resource "snowflake_user" "datadog" {
  name              = "DATADOG_USER"
  login_name        = "DATADOG_USER"
  comment           = "Dedicated Datadog monitoring service user"
  default_role      = snowflake_account_role.datadog.name
  default_warehouse = snowflake_warehouse.datadog.name
  rsa_public_key    = local.datadog_public_key
}

resource "snowflake_grant_account_role" "datadog_user" {
  role_name = snowflake_account_role.datadog.name
  user_name = snowflake_user.datadog.name
}

resource "snowflake_grant_privileges_to_account_role" "warehouse_usage" {
  account_role_name = snowflake_account_role.datadog.name
  privileges        = ["USAGE"]

  on_account_object {
    object_type = "WAREHOUSE"
    object_name = snowflake_warehouse.datadog.name
  }
}

resource "snowflake_grant_database_role" "security_viewer" {
  database_role_name = "\"SNOWFLAKE\".\"SECURITY_VIEWER\""
  parent_role_name   = snowflake_account_role.datadog.name
}

resource "snowflake_grant_database_role" "usage_viewer" {
  database_role_name = "\"SNOWFLAKE\".\"USAGE_VIEWER\""
  parent_role_name   = snowflake_account_role.datadog.name
}