# 🔧 Webhook Local Testing Setup Script

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  PayOS Webhook Local Testing Setup" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Check if ngrok is installed
$ngrokInstalled = Get-Command ngrok -ErrorAction SilentlyContinue

if (-not $ngrokInstalled) {
    Write-Host "⚠️  ngrok is not installed" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "ngrok is required to expose your local server to the internet for webhook testing." -ForegroundColor White
    Write-Host "Without ngrok, PayOS cannot send webhooks to your localhost." -ForegroundColor White
    Write-Host ""
    Write-Host "Options to install ngrok:" -ForegroundColor Cyan
    Write-Host "  1. Using Chocolatey (Recommended):" -ForegroundColor White
    Write-Host "     choco install ngrok" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  2. Using Scoop:" -ForegroundColor White
    Write-Host "     scoop install ngrok" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  3. Manual Download:" -ForegroundColor White
    Write-Host "     Download from: https://ngrok.com/download" -ForegroundColor Gray
    Write-Host "     Extract and add to PATH" -ForegroundColor Gray
    Write-Host ""
    
    $install = Read-Host "Do you want to install ngrok via Chocolatey now? (Y/N)"
    
    if ($install -eq "Y" -or $install -eq "y") {
        Write-Host "Installing ngrok via Chocolatey..." -ForegroundColor Green
        
        # Check if Chocolatey is installed
        $chocoInstalled = Get-Command choco -ErrorAction SilentlyContinue
        
        if (-not $chocoInstalled) {
            Write-Host "❌ Chocolatey is not installed" -ForegroundColor Red
            Write-Host "Please install Chocolatey first from: https://chocolatey.org/install" -ForegroundColor Yellow
            Write-Host ""
            Write-Host "Or install ngrok manually from: https://ngrok.com/download" -ForegroundColor Yellow
            exit 1
        }
        
        # Install ngrok
        choco install ngrok -y
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ ngrok installed successfully!" -ForegroundColor Green
            Write-Host "Please restart your terminal and run this script again." -ForegroundColor Yellow
            exit 0
        } else {
            Write-Host "❌ Failed to install ngrok" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "❌ Cannot continue without ngrok" -ForegroundColor Red
        Write-Host "Please install ngrok manually and run this script again." -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "✅ ngrok is installed" -ForegroundColor Green
Write-Host ""

# Check if server is running
Write-Host "Checking if API server is running on port 8080..." -ForegroundColor Cyan
$serverRunning = Test-NetConnection -ComputerName localhost -Port 8080 -InformationLevel Quiet -WarningAction SilentlyContinue

if (-not $serverRunning) {
    Write-Host "⚠️  API server is not running on port 8080" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please start your server in another terminal first:" -ForegroundColor White
    Write-Host "  go run cmd/api/main.go" -ForegroundColor Gray
    Write-Host ""
    $continue = Read-Host "Press Enter when server is running, or type 'exit' to quit"
    if ($continue -eq "exit") {
        exit 0
    }
    
    # Check again
    $serverRunning = Test-NetConnection -ComputerName localhost -Port 8080 -InformationLevel Quiet -WarningAction SilentlyContinue
    if (-not $serverRunning) {
        Write-Host "❌ Server still not running. Exiting." -ForegroundColor Red
        exit 1
    }
}

Write-Host "✅ API server is running on port 8080" -ForegroundColor Green
Write-Host ""

# Start ngrok
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Starting ngrok tunnel..." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "🚀 Starting ngrok on port 8080..." -ForegroundColor Green
Write-Host ""
Write-Host "IMPORTANT INSTRUCTIONS:" -ForegroundColor Yellow
Write-Host "1. Copy the ngrok HTTPS URL (e.g., https://abc123.ngrok.io)" -ForegroundColor White
Write-Host "2. Add '/payments/webhook' to the URL" -ForegroundColor White
Write-Host "3. Configure this URL in your PayOS Dashboard:" -ForegroundColor White
Write-Host "   → Go to https://payos.vn/" -ForegroundColor Gray
Write-Host "   → Settings → Webhook Configuration" -ForegroundColor Gray
Write-Host "   → Set: https://YOUR-NGROK-URL.ngrok.io/payments/webhook" -ForegroundColor Gray
Write-Host "4. Test with a payment!" -ForegroundColor White
Write-Host ""
Write-Host "Press Ctrl+C to stop ngrok when done testing" -ForegroundColor Yellow
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan

# Start ngrok (this will keep running)
ngrok http 8080
