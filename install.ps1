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

# Get latest release (including pre-releases)
Write-Host "Fetching latest release..." -ForegroundColor Yellow
try {
    # Try latest release first
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/postacksol/flux-relay-cli/releases/latest" -ErrorAction Stop
    $version = $release.tag_name
    Write-Host "Latest version: $version" -ForegroundColor Green
} catch {
    # If no latest release, try getting all releases
    try {
        $allReleases = Invoke-RestMethod -Uri "https://api.github.com/repos/postacksol/flux-relay-cli/releases" -ErrorAction Stop
        if ($allReleases.Count -gt 0) {
            $release = $allReleases[0]
            $version = $release.tag_name
            Write-Host "Using release: $version" -ForegroundColor Yellow
        } else {
            throw "No releases found"
        }
    } catch {
        Write-Host "Error: Could not fetch release. Please build from source or wait for binaries." -ForegroundColor Red
        Write-Host "Build from source: git clone https://github.com/postacksol/flux-relay-cli.git && cd flux-relay-cli && go build -o flux-relay.exe ." -ForegroundColor Yellow
        exit 1
    }
}

# Find Windows binary
$asset = $release.assets | Where-Object { 
    ($_.name -like "*windows*$arch*" -or $_.name -like "*windows*.exe" -or $_.name -like "*Windows*$arch*") -and
    $_.name -notlike "*.zip" -and $_.name -notlike "*.tar.gz"
} | Select-Object -First 1

if (-not $asset) {
    Write-Host "Error: Could not find Windows binary for this architecture" -ForegroundColor Red
    Write-Host "Available assets:" -ForegroundColor Yellow
    $release.assets | ForEach-Object { Write-Host "  - $($_.name)" -ForegroundColor Gray }
    Write-Host "`nThe release may not have binaries yet. The GitHub Actions workflow should build them automatically." -ForegroundColor Yellow
    Write-Host "Please check: https://github.com/postacksol/flux-relay-cli/actions" -ForegroundColor Cyan
    Write-Host "`nAlternatively, build from source:" -ForegroundColor Yellow
    Write-Host "  git clone https://github.com/postacksol/flux-relay-cli.git" -ForegroundColor Gray
    Write-Host "  cd flux-relay-cli" -ForegroundColor Gray
    Write-Host "  go build -o flux-relay.exe ." -ForegroundColor Gray
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
