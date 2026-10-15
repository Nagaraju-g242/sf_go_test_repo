
param(
    [Parameter(Mandatory = $true)]
    [string]$EntityId,

    [string]$MonitorName = "[AUTO] Snowflake Login History",

    [int]$Critical = 100,
    [int]$Warning = 50,

    [switch]$Create
)

$ErrorActionPreference = "Stop"

$site = if ($env:DD_SITE) { $env:DD_SITE } else { "datadoghq.com" }

if (-not $env:DD_API_KEY -or -not $env:DD_APP_KEY) {
    throw "Datadog API and application keys are required."
}

$headers = @{
    "DD-API-KEY"         = $env:DD_API_KEY
    "DD-APPLICATION-KEY" = $env:DD_APP_KEY
    "Content-Type"      = "application/json"
}

$sql = @"
SELECT
    USER_NAME,
    COUNT(*) AS dd_value
FROM SNOWFLAKE.ACCOUNT_USAGE.LOGIN_HISTORY
WHERE EVENT_TIMESTAMP >= DATEADD('HOUR', -24, CURRENT_TIMESTAMP())
GROUP BY USER_NAME
ORDER BY dd_value DESC
"@

$variable = @{
    name           = "query1"
    data_source    = "data_quality_metrics"
    schema_version = "0.0.1"
    filter         = 'search for database_table where `entity_id:' + $EntityId + '`'
    group_by       = @("entity_id", "table", "schema", "database")
    measure        = "custom"
    monitor_options = @{
        custom_sql           = $sql
        model_type_override  = "any"
        crontab_override     = "0 * * * *"
    }
}

$payload = @{
    name    = $MonitorName
    type    = "data-quality alert"
    query   = "formula(`"query1`").last(`"2h`") >= $Critical"
    message = "Automated Snowflake login history monitor."
    tags    = @("service:snowflake", "automation:powershell")
    options = @{
        thresholds = @{
            critical = $Critical
            warning  = $Warning
        }
        variables       = @($variable)
        include_tags    = $true
        on_missing_data = "show_and_notify_no_data"
        notify_audit    = $true
    }
}

Write-Host "Monitor: $MonitorName"
Write-Host "Entity ID: $EntityId"
Write-Host "Type: $($payload.type)"
Write-Host "SQL schedule: 0 * * * *"

if (-not $Create) {
    Write-Host "DRY RUN: No monitor was created." -ForegroundColor Yellow
    return
}

$baseUrl = "https://api.$site/api/v1/monitor"

# Check whether a monitor with the same name already exists.
$existing = Invoke-RestMethod -Method Get -Uri $baseUrl -Headers $headers

if (@($existing | Where-Object { $_.name -eq $MonitorName }).Count -gt 0) {
    throw "Monitor name already exists. Choose a different name."
}

$json = $payload | ConvertTo-Json -Depth 30

try {
    $created = Invoke-RestMethod `
        -Method Post `
        -Uri $baseUrl `
        -Headers $headers `
        -Body $json

    Write-Host "Monitor created successfully!" -ForegroundColor Green
    Write-Host "Monitor ID: $($created.id)"
    Write-Host "Monitor URL: https://app.$site/monitors/$($created.id)"
}
catch {
    Write-Host "Monitor creation failed." -ForegroundColor Red

    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    } else {
        Write-Host $_.Exception.Message
    }

    throw
}
