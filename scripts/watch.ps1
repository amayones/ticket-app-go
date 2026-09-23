$ErrorActionPreference = "Continue"
Write-Host "Watching for changes... (Ctrl+C to stop)" -ForegroundColor Cyan
Write-Host "Go + frontend/src -> auto rebuild + restart via PID file" -ForegroundColor DarkGray

$root = (Resolve-Path "$PSScriptRoot\..").Path
$pidFile = Join-Path ([IO.Path]::GetTempPath()) "go-core.pid"
$legacyPid = Join-Path ([IO.Path]::GetTempPath()) "golang-backend.pid"

function Stop-App {
  foreach ($pf in @($pidFile, $legacyPid)) {
    if (Test-Path $pf) {
      $pidStr = (Get-Content $pf -Raw).Trim()
      $procId = 0
      if ([int]::TryParse($pidStr, [ref]$procId)) {
        Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
      }
      Remove-Item $pf -Force -ErrorAction SilentlyContinue
    }
  }
}

function Invoke-Build {
  Write-Host "`n[watch] change detected, rebuilding..." -ForegroundColor Yellow
  & powershell -NoProfile -ExecutionPolicy Bypass -File "$root\scripts\build.ps1"
  if ($LASTEXITCODE -ne 0) { Write-Host "[watch] build failed, will retry on next change" -ForegroundColor Red; return }
  Write-Host "[watch] restarting app.exe..." -ForegroundColor Cyan
  Stop-App
  Start-Sleep -Milliseconds 500
  $exe = Join-Path $root "app.exe"
  if (Test-Path $exe) {
    Start-Process $exe -WorkingDirectory $root | Out-Null
    Write-Host "[watch] app.exe started" -ForegroundColor Green
  }
}

$watcher = New-Object IO.FileSystemWatcher $root, "*.*"
$watcher.IncludeSubdirectories = $true
$watcher.NotifyFilter = [IO.NotifyFilters]::LastWrite -bor [IO.NotifyFilters]::FileName -bor [IO.NotifyFilters]::DirectoryName

$action = {
  $path = $Event.SourceEventArgs.FullPath
  if ($path -match "\\\.git\\|\\frontend\\node_modules\\|\\frontend\\dist\\|app\.exe$|stop\.exe$|\.gitkeep$|go-core\.pid$|golang-backend\.pid$") { return }
  if ($path -match "\.(go|js|jsx|ts|tsx|css|html)$|go\.mod$|go\.sum$") {
    Start-Sleep -Milliseconds 400
    Invoke-Build
  }
}

$handlers = @()
$handlers += Register-ObjectEvent $watcher Created -Action $action
$handlers += Register-ObjectEvent $watcher Changed -Action $action
$handlers += Register-ObjectEvent $watcher Renamed -Action $action
$watcher.EnableRaisingEvents = $true

Write-Host "[watch] watching $root" -ForegroundColor DarkGray
try { while ($true) { Start-Sleep -Seconds 1 } }
finally {
  $watcher.EnableRaisingEvents = $false
  $handlers | Unregister-Event -ErrorAction SilentlyContinue
  $watcher.Dispose()
}
