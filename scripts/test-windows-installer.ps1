#Requires -Version 5.1
<#
Exercises the generated Windows installer offline: the release download and
version lookup are replaced by local functions, the installer is run against a
locally built archive, and the installed binary, download URL, architecture
selection, checksum verification and temporary directory cleanup are checked.
On Windows the user PATH update is checked too (the user and process
environment are restored afterwards), and a run holds a file in the temporary
directory open to confirm a cleanup failure does not fail the install.
#>
[CmdletBinding()]
param(
    [string]$InstallerPath = (Join-Path $PSScriptRoot '..\zSDKs\sdk-cli\scripts\install.ps1')
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$InstallerPath = (Resolve-Path -LiteralPath $InstallerPath).Path
$isWindowsHost = [System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows)

function Assert-True {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) {
        throw "Assertion failed: $Message"
    }
    Write-Host "ok: $Message"
}

function New-InstallerSandbox {
    $root = Join-Path ([System.IO.Path]::GetTempPath()) "installer-test-$([guid]::NewGuid())"
    foreach ($name in 'temp', 'localappdata', 'install', 'payload') {
        New-Item -ItemType Directory -Path (Join-Path $root $name) -Force | Out-Null
    }
    $binary = Join-Path (Join-Path $root 'payload') 'cli.exe'
    Set-Content -LiteralPath $binary -Value "stub binary $([guid]::NewGuid())" -NoNewline -Encoding ascii
    $archive = Join-Path $root 'cli_Windows.zip'
    Compress-Archive -LiteralPath $binary -DestinationPath $archive
    [pscustomobject]@{
        Root       = $root
        Temp       = Join-Path $root 'temp'
        InstallDir = Join-Path $root 'install'
        Binary     = $binary
        Archive    = $archive
    }
}

function Invoke-Installer {
    param($Sandbox, [string]$Architecture = 'AMD64', [string]$NativeArchitecture = '')
    $env:TEMP = $Sandbox.Temp
    $env:LOCALAPPDATA = Join-Path $Sandbox.Root 'localappdata'
    $env:CLI_INSTALL_DIR = $Sandbox.InstallDir
    $env:CLI_VERSION = 'v0.0.0-test'
    $env:PROCESSOR_ARCHITECTURE = $Architecture
    if ($NativeArchitecture) {
        $env:PROCESSOR_ARCHITEW6432 = $NativeArchitecture
    }
    else {
        Remove-Item Env:PROCESSOR_ARCHITEW6432 -ErrorAction SilentlyContinue
    }
    $global:InstallerTestArchive = $Sandbox.Archive
    $global:InstallerTestDownloads = @()
    $global:InstallerTestOutput = @()
    & $InstallerPath *>&1 | ForEach-Object {
        $line = "$_"
        $global:InstallerTestOutput += $line
        $line
    }
}

function global:Invoke-RestMethod {
    throw 'the installer must not look up the latest release when CLI_VERSION is set'
}

$global:InstallerTestChecksum = $null
$global:InstallerTestArchiveName = $null
function global:Invoke-WebRequest {
    param([string]$Uri, [string]$OutFile, [switch]$UseBasicParsing)
    $global:InstallerTestDownloads += $Uri
    if ($Uri -like '*/checksums.txt') {
        $hash = $global:InstallerTestChecksum
        if (-not $hash) {
            $hash = (Get-FileHash -LiteralPath $global:InstallerTestArchive -Algorithm SHA256).Hash.ToLowerInvariant()
        }
        Set-Content -LiteralPath $OutFile -Value "$hash  $($global:InstallerTestArchiveName)" -Encoding ascii
        return
    }
    $global:InstallerTestArchiveName = ($Uri -split '/')[-1]
    Copy-Item -LiteralPath $global:InstallerTestArchive -Destination $OutFile
}

Import-Module Microsoft.PowerShell.Archive

$global:InstallerTestOpenHandle = $null
function global:Expand-Archive {
    param([string]$Path, [string]$DestinationPath, [switch]$Force)
    Microsoft.PowerShell.Archive\Expand-Archive -Path $Path -DestinationPath $DestinationPath -Force:$Force
    if ($global:InstallerTestHoldExtractedBinary) {
        $extracted = Join-Path $DestinationPath 'cli.exe'
        $global:InstallerTestOpenHandle = [System.IO.File]::Open($extracted, 'Open', 'Read', 'Read')
    }
}

function Get-UserPathEntries {
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (-not $userPath) {
        return @()
    }
    @($userPath -split ';' | Where-Object { $_ })
}

$savedEnv = @{}
foreach ($name in 'TEMP', 'LOCALAPPDATA', 'CLI_INSTALL_DIR', 'CLI_VERSION', 'PROCESSOR_ARCHITECTURE', 'PROCESSOR_ARCHITEW6432', 'Path') {
    $savedEnv[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}
$savedUserPath = if ($isWindowsHost) { [Environment]::GetEnvironmentVariable('Path', 'User') } else { $null }

$global:InstallerTestOutput = @()
try {
    Write-Host '--- install succeeds and cleans up'
    $global:InstallerTestHoldExtractedBinary = $false
    $sandbox = New-InstallerSandbox
    $output = Invoke-Installer $sandbox
    $installed = Join-Path $sandbox.InstallDir 'cli.exe'
    Assert-True (Test-Path -LiteralPath $installed) 'binary is copied into CLI_INSTALL_DIR'
    Assert-True ((Get-Content -LiteralPath $installed -Raw) -eq (Get-Content -LiteralPath $sandbox.Binary -Raw)) 'installed binary matches the archived one'
    Assert-True ($global:InstallerTestDownloads.Count -eq 2) 'the archive and checksums.txt are downloaded'
    Assert-True ($global:InstallerTestDownloads[0] -like '*/releases/download/v0.0.0-test/cli_Windows_x86_64.zip') "download URL names the pinned version and architecture: $($global:InstallerTestDownloads[0])"
    Assert-True ($global:InstallerTestDownloads[1] -like '*/releases/download/v0.0.0-test/checksums.txt') "checksums come from the same release: $($global:InstallerTestDownloads[1])"
    Assert-True (@(Get-ChildItem -LiteralPath $sandbox.Temp -Force).Count -eq 0) 'temporary directory is removed'
    Assert-True (($output -join "`n") -match 'Checksum verified') 'installer verifies the archive checksum'
    Assert-True (($output -join "`n") -match 'Installation successful') 'installer reports success'

    Write-Host '--- ARM64 selects the arm64 archive'
    $sandbox = New-InstallerSandbox
    Invoke-Installer $sandbox -Architecture 'ARM64' | Out-Null
    Assert-True ($global:InstallerTestDownloads[0] -like '*/cli_Windows_arm64.zip') "ARM64 downloads the arm64 archive: $($global:InstallerTestDownloads[0])"

    Write-Host '--- 32-bit PowerShell on 64-bit Windows selects the native archive'
    $sandbox = New-InstallerSandbox
    Invoke-Installer $sandbox -Architecture 'x86' -NativeArchitecture 'AMD64' | Out-Null
    Assert-True ($global:InstallerTestDownloads[0] -like '*/cli_Windows_x86_64.zip') "PROCESSOR_ARCHITEW6432 wins over x86: $($global:InstallerTestDownloads[0])"
    Assert-True (Test-Path -LiteralPath (Join-Path $sandbox.InstallDir 'cli.exe')) 'binary is installed for the native architecture'

    Write-Host '--- install fails when the archive does not match checksums.txt'
    $global:InstallerTestChecksum = '0' * 64
    $sandbox = New-InstallerSandbox
    $mismatch = $null
    try {
        Invoke-Installer $sandbox | Out-Null
    }
    catch {
        $mismatch = $_.Exception.Message
    }
    $global:InstallerTestChecksum = $null
    Assert-True ($mismatch -match 'Checksum mismatch') "checksum mismatch stops the install: $mismatch"
    Assert-True (-not (Test-Path -LiteralPath (Join-Path $sandbox.InstallDir 'cli.exe'))) 'nothing is installed on a checksum mismatch'
    Assert-True (@(Get-ChildItem -LiteralPath $sandbox.Temp -Force).Count -eq 0) 'temporary directory is removed after a checksum mismatch'

    if ($isWindowsHost) {
        Write-Host '--- a user PATH entry that only shares a prefix does not count as installed'
        $sandbox = New-InstallerSandbox
        [Environment]::SetEnvironmentVariable('Path', "$($sandbox.InstallDir)-old", 'User')
        Invoke-Installer $sandbox | Out-Null
        $entries = @(Get-UserPathEntries)
        Assert-True ($entries -contains $sandbox.InstallDir) "install directory is appended next to the prefixed entry: $($entries -join ';')"
        Assert-True ($entries -contains "$($sandbox.InstallDir)-old") 'the existing entry is kept'

        Write-Host '--- a user PATH entry with a trailing backslash counts as installed'
        $sandbox = New-InstallerSandbox
        [Environment]::SetEnvironmentVariable('Path', "$($sandbox.InstallDir)\", 'User')
        Invoke-Installer $sandbox | Out-Null
        $entries = @(Get-UserPathEntries)
        Assert-True ($entries.Count -eq 1 -and $entries[0] -eq "$($sandbox.InstallDir)\") "user PATH is left unchanged: $($entries -join ';')"
    }
    else {
        Write-Host '--- skipping user PATH cases: the user environment scope is Windows-only'
    }

    if ($isWindowsHost) {
        Write-Host '--- install succeeds when the temporary directory cannot be removed'
        $global:InstallerTestHoldExtractedBinary = $true
        $sandbox = New-InstallerSandbox
        $output = Invoke-Installer $sandbox
        $installed = Join-Path $sandbox.InstallDir 'cli.exe'
        Assert-True (Test-Path -LiteralPath $installed) 'binary is copied despite the held temporary file'
        Assert-True (($output -join "`n") -match 'Warning: could not remove temporary directory') 'installer warns about the leftover temporary directory'
        Assert-True (($output -join "`n") -match 'Installation successful') 'installer still reports success'
        Assert-True (@(Get-ChildItem -LiteralPath $sandbox.Temp -Force).Count -eq 1) 'temporary directory is left in place'
        $global:InstallerTestOpenHandle.Dispose()
    }
    else {
        Write-Host '--- skipping held-file cleanup case: open files can be deleted on this platform'
    }
}
catch {
    Write-Host '--- installer output'
    $global:InstallerTestOutput | ForEach-Object { Write-Host "  $_" }
    throw
}
finally {
    if ($global:InstallerTestOpenHandle) {
        $global:InstallerTestOpenHandle.Dispose()
    }
    if ($isWindowsHost) {
        [Environment]::SetEnvironmentVariable('Path', $savedUserPath, 'User')
    }
    foreach ($name in $savedEnv.Keys) {
        [Environment]::SetEnvironmentVariable($name, $savedEnv[$name], 'Process')
    }
}
