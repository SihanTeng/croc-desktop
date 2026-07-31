# croc-desktop one-line installer (Windows)
#
#   irm https://raw.githubusercontent.com/SihanTeng/croc-desktop/main/install.ps1 | iex
#
# Optional (set before piping to iex):
#   $env:VERSION = "v0.2.0"   # pin a release tag
#   $env:REPO = "SihanTeng/croc-desktop"
#   $env:GITHUB_TOKEN = "..." # higher API rate limit
#
# Linux / macOS:
#   curl -fsSL https://raw.githubusercontent.com/SihanTeng/croc-desktop/main/install.sh | bash

$ErrorActionPreference = "Stop"

# Avoid naming a parameter $Args — breaks `irm | iex` on PowerShell 7+.
$Repo = if ($env:REPO) { $env:REPO } else { "SihanTeng/croc-desktop" }
$AppName = "croc-desktop"
$GithubApi = "https://api.github.com"
$GithubDl = "https://github.com"

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Err {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

function Get-GhHeaders {
    $headers = @{
        "Accept"               = "application/vnd.github+json"
        "X-GitHub-Api-Version" = "2022-11-28"
        "User-Agent"           = "croc-desktop-install"
    }
    if ($env:GITHUB_TOKEN) {
        $headers["Authorization"] = "Bearer $($env:GITHUB_TOKEN)"
    }
    return $headers
}

function Get-LatestVersion {
    if ($env:VERSION) {
        $v = $env:VERSION
        if ($v -notmatch '^v') { $v = "v$v" }
        Write-Info "Using requested version: $v"
        return $v
    }

    Write-Info "Fetching latest release…"
    try {
        $response = Invoke-RestMethod -Uri "$GithubApi/repos/$Repo/releases/latest" -Headers (Get-GhHeaders)
        $version = $response.tag_name
        if (-not $version) { Write-Err "Could not parse latest tag from GitHub API" }
        Write-Info "Latest version: $version"
        return $version
    }
    catch {
        Write-Err "Failed to query GitHub releases for ${Repo}: $_"
    }
}

function Get-AssetName {
    param([string]$Version)
    # Matches .github/workflows/release.yml — currently windows-amd64 MSI only.
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    if ($arch -ne [System.Runtime.InteropServices.Architecture]::X64) {
        Write-Err "No prebuilt Windows package for architecture $arch (need x64). See: https://github.com/$Repo/releases"
    }
    return "${AppName}_${Version}_windows-amd64.msi"
}

function Download-Msi {
    param(
        [string]$Version,
        [string]$AssetName
    )

    $url = "$GithubDl/$Repo/releases/download/$Version/$AssetName"
    $tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("croc-desktop-install-" + [guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Force -Path $tempDir | Out-Null
    $outPath = Join-Path $tempDir $AssetName

    Write-Info "Downloading $AssetName…"
    Write-Info "  $url"
    try {
        Invoke-WebRequest -Uri $url -OutFile $outPath -UseBasicParsing -Headers (Get-GhHeaders)
    }
    catch {
        # Some hosts reject custom headers on release asset CDN; retry plain.
        try {
            Invoke-WebRequest -Uri $url -OutFile $outPath -UseBasicParsing
        }
        catch {
            Write-Err "Download failed. Check https://github.com/$Repo/releases/tag/$Version — $_"
        }
    }

    if (-not (Test-Path $outPath)) {
        Write-Err "Download produced no file"
    }
    $sizeMb = [math]::Round((Get-Item $outPath).Length / 1MB, 1)
    Write-Info "Download complete (${sizeMb} MB)"
    return $outPath
}

function Install-Msi {
    param([string]$MsiPath)

    Write-Info "Installing MSI (may prompt for administrator approval)…"
    # Per-machine WiX package → elevation. /passive shows progress, no full UI wizard.
    $msiArgs = "/i `"$MsiPath`" /passive /norestart"
    try {
        $p = Start-Process -FilePath "msiexec.exe" -ArgumentList $msiArgs -Wait -PassThru -Verb RunAs
    }
    catch {
        Write-Warn "Elevation failed or was cancelled; retrying without RunAs…"
        $p = Start-Process -FilePath "msiexec.exe" -ArgumentList $msiArgs -Wait -PassThru
    }

    if ($null -eq $p) {
        Write-Err "msiexec did not start"
    }
    # 0 = success, 3010 = success reboot required
    if ($p.ExitCode -ne 0 -and $p.ExitCode -ne 3010) {
        Write-Err "msiexec failed with exit code $($p.ExitCode)"
    }
    if ($p.ExitCode -eq 3010) {
        Write-Warn "Install succeeded; a reboot may be required."
    }
    Write-Info "MSI install finished"
}

function Main {
    Write-Host ""
    Write-Host "croc-desktop — install" -ForegroundColor Cyan
    Write-Host "Repo: https://github.com/$Repo"
    Write-Host ""

    if (-not ($IsWindows -or $env:OS -match "Windows")) {
        Write-Err "This script is for Windows. On Linux/macOS run: curl -fsSL https://raw.githubusercontent.com/$Repo/main/install.sh | bash"
    }

    $version = Get-LatestVersion
    $asset = Get-AssetName -Version $version
    $msi = Download-Msi -Version $version -AssetName $asset

    try {
        Install-Msi -MsiPath $msi
    }
    finally {
        $dir = Split-Path $msi -Parent
        if (Test-Path $dir) {
            Remove-Item -Path $dir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }

    Write-Host ""
    Write-Info "Done. Launch croc-desktop from the Start menu."
    Write-Host ""
}

Main
