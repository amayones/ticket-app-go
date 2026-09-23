$ErrorActionPreference = "Stop"
$root = (Resolve-Path "$PSScriptRoot\..").Path
Write-Host "==> Building frontend..." -ForegroundColor Cyan
Push-Location "$root\frontend"
if (!(Test-Path "node_modules")) { npm install }
npm run build
if ($LASTEXITCODE -ne 0) { throw "frontend build failed" }
Pop-Location
Write-Host "==> Building Go binaries..." -ForegroundColor Cyan
Push-Location $root
go vet ./...
if ($LASTEXITCODE -ne 0) { throw "go vet failed" }
go build -o app.exe .
if ($LASTEXITCODE -ne 0) { throw "go build app.exe failed" }
go build -o stop.exe ./cmd/stop
if ($LASTEXITCODE -ne 0) { throw "go build stop.exe failed" }
Pop-Location
$s1 = (Get-Item "$root\app.exe").Length
$s2 = (Get-Item "$root\stop.exe").Length
Write-Host "==> Done:" -ForegroundColor Green
Write-Host "    app.exe  ($([math]::Round($s1/1MB,2)) MB) - backend + frontend embed :1067" -ForegroundColor Green
Write-Host "    stop.exe ($([math]::Round($s2/1KB,1)) KB) - stop helper" -ForegroundColor Green
Write-Host "    Root tetap bersih: .env .gitignore README.md app.exe stop.exe" -ForegroundColor DarkGray
Write-Host "    Run: .\app.exe              (console)" -ForegroundColor Yellow
Write-Host "         .\app.exe --hide       (hide background, PID %TEMP%\golang-backend.pid)" -ForegroundColor Yellow
Write-Host "         .\stop.exe             (stop hidden app.exe)" -ForegroundColor Yellow
Write-Host "         .\application\watch.ps1 (auto rebuild on change)" -ForegroundColor Yellow
