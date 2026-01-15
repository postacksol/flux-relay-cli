# Flux Relay CLI Installer for Windows
# Usage: 
#   Option 1 (One-liner): irm https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 | iex
#   Option 2: .\install.ps1
#   Option 3: Invoke-WebRequest -Uri https://raw.githubusercontent.com/postacksol/flux-relay-cli/main/install.ps1 -OutFile install.ps1; .\install.ps1

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

# Find Windows binary - try multiple patterns
$asset = $null
$patterns = @(
    "*windows*$arch*.exe",
    "*windows*.exe",
    "*Windows*$arch*.exe",
    "*Windows*.exe",
    "flux-relay-windows-$arch.exe",
    "flux-relay-windows-amd64.exe",
    "*.exe"
)

foreach ($pattern in $patterns) {
    $asset = $release.assets | Where-Object { 
        $_.name -like $pattern -and
        $_.name -notlike "*.zip" -and 
        $_.name -notlike "*.tar.gz" -and
        $_.name -notlike "*source*"
    } | Select-Object -First 1
    
    if ($asset) {
        Write-Host "Found binary: $($asset.name)" -ForegroundColor Green
        break
    }
}

if (-not $asset) {
    Write-Host "Warning: Could not find Windows binary for this architecture" -ForegroundColor Yellow
    Write-Host "Available assets:" -ForegroundColor Yellow
    if ($release.assets.Count -gt 0) {
        $release.assets | ForEach-Object { Write-Host "  - $($_.name)" -ForegroundColor Gray }
    } else {
        Write-Host "  (no assets found)" -ForegroundColor Gray
    }
    Write-Host "`nThe release may not have binaries yet. Building from source..." -ForegroundColor Yellow
    
    # Check if Go is installed
    $goInstalled = $false
    try {
        $goVersion = go version 2>&1
        if ($LASTEXITCODE -eq 0) {
            $goInstalled = $true
            Write-Host "Go found: $goVersion" -ForegroundColor Green
        }
    } catch {
        $goInstalled = $false
    }
    
    if (-not $goInstalled) {
        Write-Host "`nError: Go is not installed. Please install Go first:" -ForegroundColor Red
        Write-Host "  https://go.dev/dl/" -ForegroundColor Cyan
        Write-Host "`nOr download a pre-built binary from:" -ForegroundColor Yellow
        Write-Host "  https://github.com/postacksol/flux-relay-cli/releases" -ForegroundColor Cyan
        exit 1
    }
    
    # Create temporary directory for building (with timestamp to avoid cache issues)
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $tempDir = Join-Path $env:TEMP "flux-relay-cli-install-$timestamp"
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    
    # Check if Git is installed
    $gitInstalled = $false
    try {
        $gitVersion = git --version 2>&1
        if ($LASTEXITCODE -eq 0) {
            $gitInstalled = $true
            Write-Host "Git found: $gitVersion" -ForegroundColor Green
        }
    } catch {
        $gitInstalled = $false
    }
    
    if (-not $gitInstalled) {
        Write-Host "`nError: Git is not installed or not in PATH. Please install Git first:" -ForegroundColor Red
        Write-Host "  https://git-scm.com/download/win" -ForegroundColor Cyan
        Write-Host "`nOr download a pre-built binary from:" -ForegroundColor Yellow
        Write-Host "  https://github.com/postacksol/flux-relay-cli/releases" -ForegroundColor Cyan
        exit 1
    }
    
    Write-Host "Cloning repository..." -ForegroundColor Yellow
    # Suppress PowerShell error handling for git (it writes to stderr even on success)
    $ErrorActionPreference = "Continue"
    $gitOutput = git clone --depth 1 --branch main https://github.com/postacksol/flux-relay-cli.git $tempDir 2>&1
    $cloneExitCode = $LASTEXITCODE
    $ErrorActionPreference = "Stop"
    
    # Check if clone actually succeeded by verifying directory and key files exist
    if ($cloneExitCode -ne 0 -or -not (Test-Path $tempDir) -or -not (Test-Path (Join-Path $tempDir "go.mod"))) {
        Write-Host "Error: Failed to clone repository" -ForegroundColor Red
        if ($gitOutput) {
            # Filter out the "Cloning into..." message which is normal stderr output
            $errorOutput = $gitOutput | Where-Object { $_ -notlike "*Cloning into*" -and $_ -notlike "*done*" }
            if ($errorOutput) {
                Write-Host "Git error: $errorOutput" -ForegroundColor Gray
            }
        }
        Write-Host "Make sure Git is installed and you have internet connectivity." -ForegroundColor Yellow
        Write-Host "  https://git-scm.com/download/win" -ForegroundColor Cyan
        exit 1
    }
    
    Write-Host "Repository cloned successfully!" -ForegroundColor Green
    
    Write-Host "Building from source..." -ForegroundColor Yellow
    Push-Location $tempDir
    try {
        $buildOutput = go build -o flux-relay.exe . 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "Build failed: $buildOutput"
        }
        Write-Host "Build successful!" -ForegroundColor Green
    } catch {
        Write-Host "Error: Build failed" -ForegroundColor Red
        Write-Host $buildOutput -ForegroundColor Red
        Pop-Location
        exit 1
    }
    Pop-Location
    
    # Determine install location
    $installDir = "$env:USERPROFILE\.flux-relay\bin"
    $finalBinPath = Join-Path $installDir "flux-relay.exe"
    
    # Set the binary path to the built file
    $builtBinPath = Join-Path $tempDir "flux-relay.exe"
    if (-not (Test-Path $builtBinPath)) {
        Write-Host "Error: Built binary not found at $builtBinPath" -ForegroundColor Red
        exit 1
    }
    
    # Create install directory
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }
    
    # Copy built binary to install location
    Write-Host "Installing binary..." -ForegroundColor Yellow
    Copy-Item -Path $builtBinPath -Destination $finalBinPath -Force
    Write-Host "Installed successfully!" -ForegroundColor Green
    
    # Clean up temp directory
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    
    # Set variables for PATH section
    $binPath = $finalBinPath
} else {
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
