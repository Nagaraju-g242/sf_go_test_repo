
param(
    [long]$SourceMonitorId = 325478312,
    [string]$MonitorName = "[AUTO-TEST] Snowflake Login History",
    [switch]$Create
)

$ErrorActionPreference = "Stop"

if (-not $env:DD_API_KEY -or -not $env:DD_APP_KEY) {
    throw "Set DD_API_KEY and DD_APP_KEY before running this script."
}

$site = if ($env:DD_SITE) { $env:DD_SITE } else { "datadoghq.com" }
$baseUrl = "https://api.$site/api/v1/monitor"

$headers = @{
    "DD-API-KEY"         = $env:DD_API_KEY
    "DD-APPLICATION-KEY" = $env:DD_APP_KEY
    "Content-Type"      = "application/json"
}

Write-Host "Fetching existing monitor $SourceMonitorId..."

$source = Invoke-RestMethod `
    -Method Get `
    -Uri "$baseUrl/$SourceMonitorId" `
    -Headers $headers

if ($source.type -ne "data-quality alert") {
    throw "Source monitor is not a data-quality alert."
}

if (-not $source.options.variables -or
    -not $source.options.variables[0].monitor_options.custom_sql) {
    throw "Source monitor has no Custom SQL configuration."
}

# Build a new request using only monitor creation fields.
# Do not send source monitor ID, timestamps, or status fields.
$payload = @{
    name    = $MonitorName
    type    = $source.type
    query   = $source.query
    message = "Automated Snowflake Custom SQL monitor test."
    tags    = @("service:snowflake", "automation:test", "team:sre")
    options = @{
        thresholds = @{
            critical = $source.options.thresholds.critical
            warning  = $source.options.thresholds.warning
        }
        variables = @(
            @{
                name           = $source.options.variables[0].name
                data_source    = $source.options.variables[0].data_source
                schema_version = $source.options.variables[0].schema_version
                filter         = $source.options.variables[0].filter
                group_by       = @($source.options.variables[0].group_by)
                measure        = $source.options.variables[0].measure
                monitor_options = @{
                    custom_sql = $source.options.variables[0].monitor_options.custom_sql
                    model_type_override = $source.options.variables[0].monitor_options.model_type_override
                    crontab_override = $source.options.variables[0].monitor_options.crontab_override
                }
            }
        )
        include_tags    = $true
        on_missing_data = $source.options.on_missing_data
        notify_audit    = $source.options.notify_audit
    }
}

$json = $payload | ConvertTo-Json -Depth 30

if (-not $Create) {
    Write-Host ""
    Write-Host "DRY RUN: No monitor created." -ForegroundColor Yellow
    Write-Host "Name: $($payload.name)"
    Write-Host "Type: $($payload.type)"
    Write-Host "Query: $($payload.query)"
    Write-Host "Data source: $($payload.options.variables[0].data_source)"
    Write-Host "Entity filter: $($payload.options.variables[0].filter)"
    Write-Host "SQL schedule: $($payload.options.variables[0].monitor_options.crontab_override)"
    Write-Host ""
    Write-Host "Run with -Create to create the test monitor."
    return
}

# Avoid creating duplicate test monitors.
$existing = Invoke-RestMethod `
    -Method Get `
    -Uri $baseUrl `
    -Headers $headers

$duplicate = @($existing) | Where-Object { $_.name -eq $MonitorName }

if ($duplicate) {
    throw "A monitor named '$MonitorName' already exists (ID: $($duplicate[0].id))."
}

Write-Host "Creating test monitor..."

try {
    $created = Invoke-RestMethod `
        -Method Post `
        -Uri $baseUrl `
        -Headers $headers `
        -Body $json

    Write-Host "Monitor created successfully!" -ForegroundColor Green
    Write-Host "Monitor ID: $($created.id)"
    Write-Host "Monitor Name: $($created.name)"
    Write-Host "Monitor Type: $($created.type)"
    Write-Host "Monitor URL: https://app.$site/monitors/$($created.id)"
}
catch {
    Write-Host "Monitor creation failed." -ForegroundColor Red

    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
    else {
        Write-Host $_.Exception.Message
    }

    throw
}
