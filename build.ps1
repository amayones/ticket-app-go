$ErrorActionPreference = "Stop"
Write-Host "==> Building frontend..." -ForegroundColor Cyan
Push-Location frontend
if (!(Test-Path "node_modules")) { npm install }
npm run build
if ($LASTEXITCODE -ne 0) { throw "frontend build failed" }
Pop-Location
Write-Host "==> Building Go binaries..." -ForegroundColor Cyan
go vet ./...
if ($LASTEXITCODE -ne 0) { throw "go vet failed" }
go build -o app.exe .
if ($LASTEXITCODE -ne 0) { throw "go build app.exe failed" }
go build -o stop.exe ./cmd/stop
if ($LASTEXITCODE -ne 0) { throw "go build stop.exe failed" }
$s1 = (Get-Item app.exe).Length
$s2 = (Get-Item stop.exe).Length
Write-Host "==> Done:" -ForegroundColor Green
Write-Host "    app.exe  ($([math]::Round($s1/1MB,2)) MB) - backend + frontend embed :1067" -ForegroundColor Green
Write-Host "    stop.exe ($([math]::Round($s2/1KB,1)) KB) - stop helper" -ForegroundColor Green
Write-Host "    Run: .\app.exe              (console)" -ForegroundColor Yellow
Write-Host "         .\app.exe --hide       (hide to tray/background, PID file %TEMP%\golang-backend.pid)" -ForegroundColor Yellow
Write-Host "         .\stop.exe             (stop hidden app.exe)" -ForegroundColor Yellow
Write-Host "         .\watch.ps1            (auto rebuild on Go/frontend change)" -ForegroundColor Yellow
