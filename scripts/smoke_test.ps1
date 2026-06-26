$ErrorActionPreference = "Stop"

$Base = if ($env:API_BASE) { $env:API_BASE } else { "http://localhost:8080" }
$Admin = if ($env:ADMIN_TOKEN) { $env:ADMIN_TOKEN } else { "dev-admin-secret" }
$Failed = 0
$TmpDir = Join-Path $PSScriptRoot "..\.tmp-smoke"
New-Item -ItemType Directory -Force -Path $TmpDir | Out-Null
$Utf8NoBom = New-Object System.Text.UTF8Encoding $false

function Write-JsonFile {
    param([string]$Path, [string]$Content)
    [System.IO.File]::WriteAllText($Path, $Content, $Utf8NoBom)
}

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

Write-Host "=== Smoke test: $Base ==="

# GET
Test-Status "health" 200 @("$Base/health")
Test-Status "list tags" 200 @("$Base/tags")
Test-Status "list manuals" 200 @("$Base/manuals")
Test-Status "get manual" 200 @("$Base/manuals/a1000000-0000-4000-8000-000000000001")
Test-Status "filter by tag" 200 @("$Base/manuals?tag_id=1&sort=popular")
Test-Status "search" 200 @("$Base/manuals?q=Go")
Test-Status "invalid uuid" 400 @("$Base/manuals/not-a-uuid")
Test-Status "not found" 404 @("$Base/manuals/00000000-0000-0000-0000-000000000099")

# POST validation
$emptyJson = Join-Path $TmpDir "empty.json"
Write-JsonFile $emptyJson '{"title":"","author":""}'
Test-Status "validation error" 400 @("-X", "POST", "$Base/manuals", "-H", "Content-Type: application/json", "--data-binary", "@$emptyJson")

# Admin
Test-Status "admin denied" 401 @("-X", "DELETE", "$Base/manuals/a1000000-0000-4000-8000-000000000001")
Test-Status "admin allowed check" 401 @("-X", "PUT", "$Base/manuals/a1000000-0000-4000-8000-000000000001", "-H", "Content-Type: application/json", "-d", "{}")

# Create + delete flow (admin)
$createJson = Join-Path $TmpDir "create.json"
Write-JsonFile $createJson '{"title":"Smoke Test","author":"Tester","content":"Temporary record for smoke test"}'
$createOut = Join-Path $TmpDir "create_out.json"
$createCode = & curl.exe -s -o $createOut -w "%{http_code}" -X POST "$Base/manuals" -H "Content-Type: application/json" --data-binary "@$createJson"
if ([int]$createCode -ne 201) {
    Write-Host "FAIL create manual (expected 201, got $createCode)"
    $Failed++
} else {
    Write-Host "OK   create manual"
    $body = Get-Content $createOut -Raw | ConvertFrom-Json
    $id = $body.data.id
    $delCode = & curl.exe -s -o NUL -w "%{http_code}" -X DELETE "$Base/manuals/$id" -H "X-Admin-Token: $Admin"
    if ([int]$delCode -ne 204) {
        Write-Host "FAIL delete manual (expected 204, got $delCode)"
        $Failed++
    } else {
        Write-Host "OK   delete manual (admin)"
    }
}

Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue

if ($Failed -gt 0) {
    Write-Host "$Failed test(s) failed"
    exit 1
}

Write-Host "All smoke tests passed"
exit 0
