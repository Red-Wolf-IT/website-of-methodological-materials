$ErrorActionPreference = "Stop"

$Base = if ($env:API_BASE) { $env:API_BASE } else { "http://localhost:8080" }
$Admin = if ($env:ADMIN_TOKEN) { $env:ADMIN_TOKEN } else { "dev-admin-secret" }
$Failed = 0

function Test-Status {
    param(
        [string]$Name,
        [int]$Expected,
        [string[]]$CurlArgs
    )

    $code = & curl.exe -s -o NUL -w "%{http_code}" @CurlArgs
    if ([int]$code -ne $Expected) {
        Write-Host "FAIL $Name (expected $Expected, got $code)"
        $script:Failed++
    } else {
        Write-Host "OK   $Name"
    }
}

Test-Status "health" 200 @("$Base/health")
Test-Status "list manuals" 200 @("$Base/manuals")
Test-Status "get manual" 200 @("$Base/manuals/a1000000-0000-4000-8000-000000000001")
Test-Status "filter by tag" 200 @("$Base/manuals?tag_id=1&sort=popular")
Test-Status "search" 200 @("$Base/manuals?q=Go")
Test-Status "invalid uuid" 400 @("$Base/manuals/not-a-uuid")
Test-Status "not found" 404 @("$Base/manuals/00000000-0000-0000-0000-000000000099")
Test-Status "admin denied" 401 @("-X", "DELETE", "$Base/manuals/a1000000-0000-4000-8000-000000000001")

if ($Failed -gt 0) {
    Write-Host "$Failed test(s) failed"
    exit 1
}

Write-Host "All smoke tests passed"
exit 0
