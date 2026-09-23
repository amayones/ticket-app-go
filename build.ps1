$ErrorActionPreference = "Stop"
Write-Host "==> Building frontend..." -ForegroundColor Cyan
Push-Location frontend
if (!(Test-Path "node_modules")) { npm install }
npm run build
if ($LASTEXITCODE -ne 0) { throw "frontend build failed" }
Pop-Location
Write-Host "==> Building Go binary (embedded frontend/dist)..." -ForegroundColor Cyan
go vet ./...
if ($LASTEXITCODE -ne 0) { throw "go vet failed" }
go build -o app.exe .
if ($LASTEXITCODE -ne 0) { throw "go build failed" }
$s = (Get-Item app.exe).Length
Write-Host "==> Done: app.exe ($([math]::Round($s/1MB, 2)) MB) - single binary, embedded frontend" -ForegroundColor Green
Write-Host "    Run: .\app.exe  (frontend + API di satu port)" -ForegroundColor Yellow
