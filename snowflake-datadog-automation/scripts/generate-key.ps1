param([string]$OutputDirectory = ".\secrets")
New-Item -ItemType Directory -Force -Path $OutputDirectory | Out-Null
openssl genrsa -out "$OutputDirectory\datadog_rsa_key.pem" 2048
openssl pkcs8 -topk8 -inform PEM -in "$OutputDirectory\datadog_rsa_key.pem" -out "$OutputDirectory\datadog_rsa_key.p8" -nocrypt
openssl rsa -in "$OutputDirectory\datadog_rsa_key.pem" -pubout -out "$OutputDirectory\datadog_rsa_key.pub"
Write-Host "Register the base64 public key in Snowflake; never commit private keys."
