# Static unit tests for install.ps1 helper functions.
$ErrorActionPreference = 'Stop'

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
. (Join-Path $scriptDir 'install.ps1')

$pass = 0
$fail = 0

function Assert-Equal {
    param(
        [string]$Description,
        [object]$Expected,
        [object]$Actual
    )

    if ("$Expected" -eq "$Actual") {
        Write-Host "  PASS: $Description"
        $script:pass++
    } else {
        Write-Host "  FAIL: $Description"
        Write-Host "    expected: '$Expected'"
        Write-Host "    actual:   '$Actual'"
        $script:fail++
    }
}

Write-Host '=== Get-ReleaseArchiveCandidates ==='
$candidates = Get-ReleaseArchiveCandidates -Version 'v0.29.2-periscope.2' -Arch 'amd64'
Assert-Equal 'prefers periscope archive' 'periscope_0.29.2-periscope.2_windows_amd64.zip' $candidates[0]
Assert-Equal 'includes legacy agentsview archive' 'agentsview_v0.29.2-periscope.2_windows_amd64.zip' $candidates[3]

Write-Host ''
Write-Host '=== Test-SkipChecksum ==='
$env:PERISCOPE_SKIP_CHECKSUM = '0'
$env:AGENTSVIEW_SKIP_CHECKSUM = '0'
Assert-Equal 'default off' 'False' (Test-SkipChecksum)

$env:PERISCOPE_SKIP_CHECKSUM = '1'
$env:AGENTSVIEW_SKIP_CHECKSUM = '0'
Assert-Equal 'periscope env' 'True' (Test-SkipChecksum)

$env:PERISCOPE_SKIP_CHECKSUM = '0'
$env:AGENTSVIEW_SKIP_CHECKSUM = '1'
Assert-Equal 'legacy agentsview env' 'True' (Test-SkipChecksum)

Remove-Item Env:PERISCOPE_SKIP_CHECKSUM -ErrorAction SilentlyContinue
Remove-Item Env:AGENTSVIEW_SKIP_CHECKSUM -ErrorAction SilentlyContinue

Write-Host ''
Write-Host '=== Get-InstallDir ==='
$env:PERISCOPE_INSTALL_DIR = 'C:\periscope\bin'
$env:AGENTSVIEW_INSTALL_DIR = 'C:\agentsview\bin'
Assert-Equal 'prefers periscope install dir' 'C:\periscope\bin' (Get-InstallDir)
Remove-Item Env:PERISCOPE_INSTALL_DIR -ErrorAction SilentlyContinue
Assert-Equal 'falls back to agentsview install dir' 'C:\agentsview\bin' (Get-InstallDir)
Remove-Item Env:AGENTSVIEW_INSTALL_DIR -ErrorAction SilentlyContinue
Assert-Equal 'default periscope path' (Join-Path $env:USERPROFILE '.periscope\bin') (Get-InstallDir)

Write-Host ''
Write-Host "Results: $pass passed, $fail failed"
if ($fail -ne 0) {
    exit 1
}
