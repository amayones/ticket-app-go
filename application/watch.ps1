$ErrorActionPreference = "SilentlyContinue"
Write-Host "Watching for changes... (Ctrl+C to stop)" -ForegroundColor Cyan
Write-Host "Go + frontend/src -> auto build.ps1 + restart app.exe" -ForegroundColor DarkGray

$root = (Resolve-Path "$PSScriptRoot\..").Path

function Invoke-Build {
  Write-Host "`n[watch] change detected, rebuilding..." -ForegroundColor Yellow
  & "$root\application\build.ps1"
  if ($LASTEXITCODE -ne 0) { Write-Host "[watch] build failed, will retry on next change" -ForegroundColor Red; return }
  Write-Host "[watch] restarting app.exe..." -ForegroundColor Cyan
  Get-Process app -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
  Start-Sleep -Milliseconds 500
  $exe = Join-Path $root "app.exe"
  if (Test-Path $exe) { Start-Process $exe -WorkingDirectory $root | Out-Null; Write-Host "[watch] app.exe started :1067" -ForegroundColor Green }
}

$watcher = New-Object IO.FileSystemWatcher $root, "*.*"
$watcher.IncludeSubdirectories = $true
$watcher.NotifyFilter = [IO.NotifyFilters]::LastWrite -bor [IO.NotifyFilters]::FileName -bor [IO.NotifyFilters]::DirectoryName

$action = {
  $path = $Event.SourceEventArgs.FullPath
  if ($path -match "\\\.git\\|\\frontend\\node_modules\\|\\frontend\\dist\\|app\.exe$|\.gitkeep$|app\.pid") { return }
  if ($path -match "\.(go|js|jsx|ts|tsx|css|html|json)$|go\.mod$") {
    Start-Sleep -Milliseconds 400
    Invoke-Build
  }
}

$handlers = @()
$handlers += Register-ObjectEvent $watcher Created -Action $action
$handlers += Register-ObjectEvent $watcher Changed -Action $action
$handlers += Register-ObjectEvent $watcher Renamed -Action $action
$watcher.EnableRaisingEvents = $true

Write-Host "[watch] watching $root (go, frontend/src, go.mod)" -ForegroundColor DarkGray
try { while ($true) { Start-Sleep -Seconds 1 } }
finally {
  $watcher.EnableRaisingEvents = $false
  $handlers | Unregister-Event -ErrorAction SilentlyContinue
  $watcher.Dispose()
}
