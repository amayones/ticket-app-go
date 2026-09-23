Write-Host "Dev mode: 2 processes (Vite HMR + Go API)" -ForegroundColor Cyan
Write-Host "  [1] http://localhost:5173 (frontend, proxy /api -> :1067)" -ForegroundColor DarkGray
Write-Host "  [2] http://localhost:1067/api (backend via go run .)" -ForegroundColor DarkGray
$root = (Resolve-Path "$PSScriptRoot\..").Path
Start-Process powershell -ArgumentList "-NoExit","-Command","cd '$root\frontend'; npm run dev"
Start-Process powershell -ArgumentList "-NoExit","-Command","cd '$root'; go run ."
