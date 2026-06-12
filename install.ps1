# speeder installer for Windows
# Usage: irm https://raw.githubusercontent.com/mhdiiilham/speeder/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo  = "mhdiiilham/speeder"
$bin   = "speeder.exe"
$dir   = "$env:LOCALAPPDATA\speeder"

# Detect architecture
$arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITEW6432 -eq "ARM64" -or $env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        "arm64"
    } else {
        "amd64"
    }
} else {
    Write-Error "speeder requires a 64-bit Windows system."
    exit 1
}

Write-Host "Installing speeder ($arch)..." -ForegroundColor Cyan

# Resolve latest release download URL
$apiUrl  = "https://api.github.com/repos/$repo/releases/latest"
$headers = @{ "User-Agent" = "speeder-installer" }

try {
    $release = Invoke-RestMethod -Uri $apiUrl -Headers $headers
} catch {
    Write-Error "Failed to fetch release info from GitHub: $_"
    exit 1
}

$assetName = "speeder-windows-$arch.exe"
$asset     = $release.assets | Where-Object { $_.name -eq $assetName } | Select-Object -First 1

if (-not $asset) {
    Write-Error "Could not find asset '$assetName' in the latest release."
    exit 1
}

$downloadUrl = $asset.browser_download_url
$version     = $release.tag_name

# Create install directory
if (-not (Test-Path $dir)) {
    New-Item -ItemType Directory -Path $dir | Out-Null
}

# Download binary
$dest = Join-Path $dir $bin
Write-Host "Downloading $version from $downloadUrl ..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $dest -UseBasicParsing

# Add install directory to user PATH (persistent, no admin required)
$userPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$dir*") {
    [System.Environment]::SetEnvironmentVariable("PATH", "$userPath;$dir", "User")
    Write-Host "Added $dir to your PATH." -ForegroundColor Green
    Write-Host ""
    Write-Host "NOTE: Restart your terminal (or open a new one) for 'speeder' to be available." -ForegroundColor Yellow
} else {
    Write-Host "$dir is already in your PATH." -ForegroundColor Green
}

Write-Host ""
Write-Host "speeder $version installed successfully!" -ForegroundColor Green
Write-Host "Run: speeder --version"
