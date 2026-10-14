# footprints installer for Windows (PowerShell)
#   irm https://raw.githubusercontent.com/MahendraDani/footprints/main/install.ps1 | iex
#
# Env:
#   FOOTPRINTS_VERSION=v0.1.0         install a specific version (default: latest)
#   FOOTPRINTS_INSTALL_DIR=C:\some\dir  install somewhere specific
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Repo = 'MahendraDani/footprints'
$Bin = 'footprints'
$Version = if ($env:FOOTPRINTS_VERSION) { $env:FOOTPRINTS_VERSION } else { 'latest' }
$InstallDir = if ($env:FOOTPRINTS_INSTALL_DIR) { $env:FOOTPRINTS_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\footprints' }

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
  'AMD64' { 'amd64' }
  'ARM64' { 'arm64' }
  default { throw "unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$asset = "${Bin}_windows_${arch}.zip"
$base = if ($Version -eq 'latest') { "https://github.com/$Repo/releases/latest/download" }
        else { "https://github.com/$Repo/releases/download/$Version" }

$tmp = Join-Path ([IO.Path]::GetTempPath()) ("footprints-" + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
  Write-Host "Downloading $asset ($Version)..."
  Invoke-WebRequest "$base/$asset" -OutFile (Join-Path $tmp $asset)
  Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

  $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match [regex]::Escape($asset) + '$' } | Select-Object -First 1
  if (-not $line) { throw "no checksum listed for $asset" }
  $want = ($line -split '\s+')[0].ToLower()
  $got = (Get-FileHash (Join-Path $tmp $asset) -Algorithm SHA256).Hash.ToLower()
  if ($want -ne $got) { throw "checksum mismatch for $asset" }

  Expand-Archive (Join-Path $tmp $asset) -DestinationPath $tmp -Force
  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

  $dest = Join-Path $InstallDir "$Bin.exe"
  if (Test-Path $dest) { Move-Item $dest "$dest.old" -Force }  # a running .exe can be renamed but not overwritten
  Move-Item (Join-Path $tmp "$Bin.exe") $dest -Force
  Remove-Item "$dest.old" -Force -ErrorAction SilentlyContinue
} finally {
  Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

# Add to the user's PATH for future terminals...
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$parts = if ($userPath) { $userPath -split ';' | Where-Object { $_ } } else { @() }
if ($parts -notcontains $InstallDir) {
  [Environment]::SetEnvironmentVariable('Path', (($parts + $InstallDir) -join ';'), 'User')
}
# ...and for this window, so it works right now
if (($env:Path -split ';') -notcontains $InstallDir) { $env:Path = "$env:Path;$InstallDir" }

Write-Host "Installed $(& $dest --version) to $InstallDir"
Write-Host "Ready. Try:  $Bin graph"