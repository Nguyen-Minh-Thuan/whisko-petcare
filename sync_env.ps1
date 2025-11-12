# Sync .env file to deployments directory
# This ensures docker-compose can read environment variables

Write-Host "🔄 Syncing .env file to deployments directory..." -ForegroundColor Cyan

$rootEnv = Join-Path $PSScriptRoot ".env"
$deployEnv = Join-Path $PSScriptRoot "deployments\.env"

if (Test-Path $rootEnv) {
    Copy-Item $rootEnv $deployEnv -Force
    Write-Host "✅ Successfully copied .env to deployments/.env" -ForegroundColor Green
    Write-Host "📝 Files are now in sync" -ForegroundColor Green
} else {
    Write-Host "❌ Error: .env file not found in root directory" -ForegroundColor Red
    Write-Host "Please create a .env file first" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "💡 Tip: Run this script whenever you update your .env file" -ForegroundColor Yellow
