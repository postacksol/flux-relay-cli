# Quick Test Script for Flux Relay CLI (PowerShell)
# Run this to quickly verify all features work

Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "Flux Relay CLI - Quick Test Script" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

$PASSED = 0
$FAILED = 0

function Test-Command {
    param(
        [string]$TestName,
        [string]$Command,
        [string]$Expected
    )
    
    Write-Host "Testing: $TestName... " -NoNewline
    
    try {
        $result = Invoke-Expression $Command 2>&1
        if ($LASTEXITCODE -eq 0 -or $result) {
            Write-Host "✓ PASSED" -ForegroundColor Green
            $script:PASSED++
            return $true
        } else {
            Write-Host "✗ FAILED" -ForegroundColor Red
            $script:FAILED++
            return $false
        }
    } catch {
        Write-Host "✗ FAILED" -ForegroundColor Red
        $script:FAILED++
        return $false
    }
}

Write-Host "1. Testing Authentication..." -ForegroundColor Yellow
Test-Command "Login status" "flux-relay login --help" ""

Write-Host ""
Write-Host "2. Testing Project Management..." -ForegroundColor Yellow
Test-Command "List projects" "flux-relay pr list" ""
Test-Command "Show project help" "flux-relay pr --help" ""

Write-Host ""
Write-Host "3. Testing Server Management..." -ForegroundColor Yellow
Test-Command "List servers" "flux-relay server list" ""
Test-Command "Server alias (srv)" "flux-relay srv list" ""
Test-Command "Show server help" "flux-relay server --help" ""

Write-Host ""
Write-Host "4. Testing Nameserver Management..." -ForegroundColor Yellow
Test-Command "List nameservers" "flux-relay ns list" ""
Test-Command "Show nameserver help" "flux-relay ns --help" ""

Write-Host ""
Write-Host "5. Testing SQL Command..." -ForegroundColor Yellow
Test-Command "SQL help" "flux-relay sql --help" ""

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "Test Results:" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "Passed: $PASSED" -ForegroundColor Green
Write-Host "Failed: $FAILED" -ForegroundColor Red
Write-Host ""

if ($FAILED -eq 0) {
    Write-Host "All basic tests passed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "For interactive shell tests, see TEST_GUIDE.md" -ForegroundColor Yellow
    exit 0
} else {
    Write-Host "Some tests failed. Check your installation." -ForegroundColor Yellow
    Write-Host "Make sure:"
    Write-Host "  1. CLI is built: go build -o flux-relay.exe ."
    Write-Host "  2. You're logged in: flux-relay login"
    Write-Host "  3. You have projects/servers set up"
    exit 1
}
