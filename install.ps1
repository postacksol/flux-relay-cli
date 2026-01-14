# Flux Relay CLI Installer for Windows
# Usage: 
#   Option 1: .\install.ps1
#   Option 2: Invoke-WebRequest -Uri https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 -OutFile install.ps1; .\install.ps1

$ErrorActionPreference = "Stop"

Write-Host "Flux Relay CLI Installer" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan
Write-Host ""

# Detect architecture
$arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
}

# Get latest release
Write-Host "Fetching latest release..." -ForegroundColor Yellow
try {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/postacksol/flux-relay-cli/releases/latest" -ErrorAction Stop
    $version = $release.tag_name
    Write-Host "Latest version: $version" -ForegroundColor Green
} catch {
    Write-Host "Error: Could not fetch latest release. Using v1.0.0" -ForegroundColor Red
    $version = "v1.0.0"
}

# Find Windows binary
$asset = $release.assets | Where-Object { $_.name -like "*windows*$arch*" -or $_.name -like "*windows*.exe" } | Select-Object -First 1

if (-not $asset) {
    Write-Host "Error: Could not find Windows binary for this architecture" -ForegroundColor Red
    exit 1
}

# Determine install location
$installDir = "$env:USERPROFILE\.flux-relay\bin"
$binPath = Join-Path $installDir "flux-relay.exe"

# Create install directory
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# Download binary
Write-Host "Downloading $($asset.name)..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $binPath -ErrorAction Stop
    Write-Host "Downloaded successfully!" -ForegroundColor Green
} catch {
    Write-Host "Error: Failed to download binary" -ForegroundColor Red
    exit 1
}

# Add to PATH (user-level)
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    Write-Host "Adding to PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to PATH" -ForegroundColor Green
    Write-Host "Note: You may need to restart your terminal for PATH changes to take effect" -ForegroundColor Yellow
} else {
    Write-Host "Already in PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host "Binary installed to: $binPath" -ForegroundColor Cyan
Write-Host ""
Write-Host "Try running: flux-relay --version" -ForegroundColor Yellow
