# Monitor Webhook Deployment

$serverIP = "15.134.38.118"
$webhookUrl = "http://${serverIP}:8080/payments/webhook"
$maxAttempts = 12  # 12 attempts = 2 minutes (10 seconds each)
$attemptDelay = 10  # seconds

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Monitoring Webhook Deployment" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "GitHub Actions is deploying to: $serverIP" -ForegroundColor White
Write-Host "This usually takes 2-3 minutes..." -ForegroundColor Gray
Write-Host ""
Write-Host "Testing webhook endpoint every $attemptDelay seconds..." -ForegroundColor Yellow
Write-Host "Press Ctrl+C to stop monitoring" -ForegroundColor Gray
Write-Host ""

for ($i = 1; $i -le $maxAttempts; $i++) {
    Write-Host "Attempt $i of ${maxAttempts}:" -NoNewline -ForegroundColor Cyan
    Write-Host " Testing..." -ForegroundColor Gray
    
    try {
        $response = Invoke-WebRequest -Uri $webhookUrl `
            -Method Post `
            -ContentType "application/json" `
            -Headers @{"x-signature" = "test"} `
            -Body '{"test": "data"}' `
            -UseBasicParsing `
            -TimeoutSec 5 `
            -ErrorAction Stop
        
        Write-Host "  [OK] Webhook responded with: $($response.StatusCode)" -ForegroundColor Green
        if ($response.StatusCode -eq 200) {
            Write-Host ""
            Write-Host "[SUCCESS] Webhook is deployed and working!" -ForegroundColor Green
            break
        }
        
    } catch {
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            
            if ($statusCode -eq 400) {
                Write-Host ""
                Write-Host "============================================" -ForegroundColor Green
                Write-Host "  SUCCESS! Webhook Deployed!" -ForegroundColor Green
                Write-Host "============================================" -ForegroundColor Green
                Write-Host ""
                Write-Host "Webhook endpoint is working correctly!" -ForegroundColor Green
                Write-Host "400 Bad Request is expected for test payloads" -ForegroundColor Gray
                Write-Host ""
                Write-Host "Next Step: Configure in PayOS Dashboard" -ForegroundColor Yellow
                Write-Host "  1. Go to: https://payos.vn/" -ForegroundColor White
                Write-Host "  2. Login to your account" -ForegroundColor White
                Write-Host "  3. Settings -> Webhook Configuration" -ForegroundColor White
                Write-Host "  4. Webhook URL: $webhookUrl" -ForegroundColor Cyan
                Write-Host "  5. Save" -ForegroundColor White
                Write-Host ""
                Write-Host "After configuring PayOS:" -ForegroundColor Yellow
                Write-Host "  - Create a payment" -ForegroundColor White
                Write-Host "  - Complete payment on PayOS" -ForegroundColor White
                Write-Host "  - Schedules will be created automatically!" -ForegroundColor Green
                Write-Host ""
                return
            } elseif ($statusCode -eq 404) {
                Write-Host "  [WAIT] Endpoint not found yet (404) - Deployment in progress..." -ForegroundColor Yellow
            } else {
                Write-Host "  [INFO] Status: $statusCode" -ForegroundColor Yellow
            }
        } else {
            $errorMsg = $_.Exception.Message
            if ($errorMsg -like "*closed unexpectedly*") {
                Write-Host "  [WAIT] Connection closed - Old code still running..." -ForegroundColor Yellow
            } elseif ($errorMsg -like "*timed out*") {
                Write-Host "  [WAIT] Timeout - Server might be restarting..." -ForegroundColor Yellow
            } else {
                Write-Host "  [WAIT] $errorMsg" -ForegroundColor Yellow
            }
        }
    }
    
    if ($i -lt $maxAttempts) {
        Write-Host "  Waiting $attemptDelay seconds..." -ForegroundColor Gray
        Start-Sleep -Seconds $attemptDelay
    }
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Yellow
Write-Host "  Deployment Taking Longer Than Expected" -ForegroundColor Yellow
Write-Host "============================================" -ForegroundColor Yellow
Write-Host ""
Write-Host "Options:" -ForegroundColor White
Write-Host ""
Write-Host "1. Keep waiting - deployment might take a bit longer" -ForegroundColor Gray
Write-Host "   Run this script again in 1-2 minutes" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Check GitHub Actions status:" -ForegroundColor Gray
Write-Host "   https://github.com/Nguyen-Minh-Thuan/whisko-petcare/actions" -ForegroundColor Cyan
Write-Host ""
Write-Host "3. Check server logs manually:" -ForegroundColor Gray
Write-Host "   ssh ubuntu@$serverIP" -ForegroundColor Cyan
Write-Host "   cd /home/ubuntu/whisko-petcare/deployments" -ForegroundColor Cyan
Write-Host "   docker-compose logs -f app" -ForegroundColor Cyan
Write-Host ""
Write-Host "4. Manual deployment:" -ForegroundColor Gray
Write-Host "   ssh ubuntu@$serverIP" -ForegroundColor Cyan
Write-Host "   cd /home/ubuntu/whisko-petcare && git pull origin main" -ForegroundColor Cyan
Write-Host "   cd deployments && docker-compose up -d --build" -ForegroundColor Cyan
Write-Host ""
