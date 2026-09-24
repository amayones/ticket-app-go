# Satu terminal untuk pengembangan harian: backend (go run .) jalan sebagai
# background job, Vite HMR jalan di foreground. Ctrl+C menghentikan keduanya.
#   pwsh ./scripts/start.ps1   (atau: task start)
$ErrorActionPreference = "Stop"
$root = (Resolve-Path "$PSScriptRoot\..").Path

$envPort = if ($env:APP_PORT) { $env:APP_PORT } else { "1067" }
if (Test-Path "$root\.env") {
  foreach ($line in Get-Content "$root\.env") {
    if ($line -match '^\s*APP_PORT\s*=\s*(.+?)\s*$') { $envPort = $Matches[1] }
  }
}

Write-Host "==> Backend (go run .) starting in background..." -ForegroundColor Cyan
$job = Start-Job -ScriptBlock {
  Set-Location $using:root
  go run .
}

function Stop-Backend {
  if ($job.State -eq 'Running') { Stop-Job $job | Out-Null }
  Remove-Job $job -Force -ErrorAction SilentlyContinue | Out-Null
  Write-Host "==> Backend stopped." -ForegroundColor DarkGray
}

# Tunggu backend sehat (maks 30 detik) sebelum menyalakan Vite.
$ready = $false
for ($i = 0; $i -lt 30; $i++) {
  Start-Sleep -Seconds 1
  if ($job.State -ne 'Running') {
    Receive-Job $job
    Stop-Backend
    throw "backend gagal start (lihat error di atas)"
  }
  try {
    $res = Invoke-WebRequest -UseBasicParsing -TimeoutSec 2 "http://localhost:$envPort/healthz"
    if ($res.StatusCode -eq 200) { $ready = $true; break }
  } catch { }
}
if (-not $ready) { Stop-Backend; throw "backend tidak sehat setelah 30 detik" }
Write-Host "==> Backend OK di http://localhost:$envPort" -ForegroundColor Green

Push-Location "$root\frontend"
try {
  if (!(Test-Path "node_modules")) {
    Write-Host "==> npm ci (pertama kali)..." -ForegroundColor Cyan
    npm ci --no-audit --no-fund
  }
  Write-Host "==> Frontend (Vite HMR) di http://localhost:5173 - Ctrl+C untuk berhenti" -ForegroundColor Cyan
  npm run dev
} finally {
  Pop-Location
  Stop-Backend
}
