# Snowflake + Datadog multi-account starter bundle

This is a **starter** for repeatable provisioning and monitoring, not an unattended end-to-end deployment. Snowflake integration registration must be completed with the Datadog Snowflake Integration API or UI, after checking the current account API payload and credential handling. Custom SQL monitors are not provisioned by this starter; metric monitor examples only.

## Prerequisites
- Terraform, OpenSSL, Snowflake admin SQL access, Datadog API + application keys
- Three Snowflake accounts integrated into the same Datadog organization
- Verified metric names and a `snowflake_account` tag in Metrics Explorer

## Steps
1. In each Snowflake account, run `scripts/generate-key.ps1` in a safe local folder. Keep private key private. Extract public key body (remove BEGIN/END and newlines) and substitute `PUBLIC_KEY_VALUE` in `snowflake/provision.sql`.
2. Run SQL as ACCOUNTADMIN separately in DEV, UAT, PROD. Review grants against Datadog's current setup wizard. Avoid granting more than needed.
3. Register each account in Datadog Integrations > Snowflake, or implement API registration after validating current API request schema. Supply private key securely, not via repository.
4. Verify metrics are arriving in Datadog Metrics Explorer, including account tags. Note that existing custom SQL monitor results may not be usable as ordinary tagged metrics.
5. Copy `config/terraform.tfvars.example` to `terraform/terraform.tfvars` and edit the account metric names and thresholds. Change the default example thresholds in main.tf to your measured baseline.
6. Set `DD_API_KEY`, `DD_APP_KEY`, and `DD_SITE` in your shell from your secret store. Then run `terraform -chdir=terraform init`, `terraform -chdir=terraform validate`, `terraform -chdir=terraform plan`, and `terraform -chdir=terraform apply`.
7. Set dashboard viewing/editing access in Datadog RBAC. `restricted_roles` controls dashboard editing, not an automatic SRE viewer grant. Confirm access with an SRE test user.

## Important limitations
- Dashboard includes individual alert-value widgets for each account. Static monitor widgets do not dynamically disappear when a dashboard template variable changes. To achieve truly account-filterable visualizations, add timeseries widgets using verified metric tags and Datadog template variables.
- Metrics and thresholds in the example are placeholders and must be validated.
- Account registration, custom SQL monitors, SRE access policy, and advanced template variables are follow-up stages requiring tenant-specific identifiers and validated API schemas.
- The provider manages monitor/dashboard objects but does not provision Snowflake credentials.
