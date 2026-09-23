$ErrorActionPreference = "Stop"
$root = (Resolve-Path "$PSScriptRoot\..").Path
$version = (& git rev-parse --short HEAD 2>$null)
if (-not $version) { $version = "dev" }

Write-Host "==> Building frontend..." -ForegroundColor Cyan
Push-Location "$root\frontend"
try {
  if (!(Test-Path "node_modules")) { npm ci --no-audit --no-fund }
  npm run build
  if ($LASTEXITCODE -ne 0) { throw "frontend build failed" }
  # vite emptyOutDir menghapus placeholder embed; buat ulang agar
  # //go:embed all:frontend/dist tetap compile di fresh clone.
  Set-Content -Path "$root\frontend\dist\.gitignore" -Value "*`n!.gitignore`n" -NoNewline:$false
} finally { Pop-Location }

Write-Host "==> go vet ./..." -ForegroundColor Cyan
Push-Location $root
try {
  go vet ./...
  if ($LASTEXITCODE -ne 0) { throw "go vet failed" }
  Write-Host "==> Building Go binaries (version $version)..." -ForegroundColor Cyan
  go build -trimpath -ldflags "-s -w -X main.version=$version" -o app.exe .
  if ($LASTEXITCODE -ne 0) { throw "go build app.exe failed" }
  go build -trimpath -ldflags "-s -w" -o stop.exe ./cmd/stop
  if ($LASTEXITCODE -ne 0) { throw "go build stop.exe failed" }
} finally { Pop-Location }

$s1 = (Get-Item "$root\app.exe").Length
$s2 = (Get-Item "$root\stop.exe").Length
Write-Host "==> Done:" -ForegroundColor Green
Write-Host ("    app.exe  ({0} MB) - backend + frontend embed" -f [math]::Round($s1/1MB,2)) -ForegroundColor Green
Write-Host ("    stop.exe ({0} KB) - stop helper" -f [math]::Round($s2/1KB,1)) -ForegroundColor Green
Write-Host "    Run: .\app.exe | .\app.exe --hide | .\stop.exe | task watch" -ForegroundColor Yellow
