# footprints uninstaller for Windows (PowerShell)
#   irm https://raw.githubusercontent.com/MahendraDani/footprints/main/uninstall.ps1 | iex
# Removes the binary, the PATH entry, and the ~/.footprints workspace.
# To keep the workspace:
#   $env:FOOTPRINTS_KEEP_DATA = 1; irm https://raw.githubusercontent.com/MahendraDani/footprints/main/uninstall.ps1 | iex
$ErrorActionPreference = 'Stop'

$InstallDir = if ($env:FOOTPRINTS_INSTALL_DIR) { $env:FOOTPRINTS_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\footprints' }
$exe = Join-Path $InstallDir 'footprints.exe'

Write-Host "[1/3] Removing the footprints binary"
if (Test-Path $exe) {
  Remove-Item $exe -Force
  Write-Host "  removed $exe"
} else {
  Write-Host "  no footprints.exe found in $InstallDir"
}
# Remove the folder only if it's now empty, so a shared custom directory is never wiped
if ((Test-Path $InstallDir) -and -not (Get-ChildItem $InstallDir -Force)) { Remove-Item $InstallDir -Force }

Write-Host "[2/3] Cleaning PATH"
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath) {
  $kept = $userPath -split ';' | Where-Object { $_ -and ($_ -ne $InstallDir) }
  [Environment]::SetEnvironmentVariable('Path', ($kept -join ';'), 'User')
  Write-Host "  removed $InstallDir from your PATH"
}

Write-Host "[3/3] Workspace"
$data = Join-Path $env:USERPROFILE '.footprints'
if ($env:FOOTPRINTS_KEEP_DATA) {
  if (Test-Path $data) { Write-Host "  kept $data (delete it yourself to remove everything)" }
} elseif (Test-Path $data) {
  Remove-Item $data -Recurse -Force
  Write-Host "  deleted $data"
} else {
  Write-Host "  nothing to delete"
}
Write-Host "Done. Open a new terminal."