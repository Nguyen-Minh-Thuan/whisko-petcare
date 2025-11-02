# Quick script to update webhook on server
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Updating Webhook on Server" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "The code has been pushed to GitHub!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. SSH into your server (or GitHub Actions will auto-deploy in 2-3 minutes)" -ForegroundColor White
Write-Host "   ssh ubuntu@15.134.38.118" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Update and restart the application:" -ForegroundColor White
Write-Host "   cd /home/ubuntu/whisko-petcare && git pull origin main && cd deployments && docker-compose restart api" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Wait 10 seconds, then test from here:" -ForegroundColor White
Write-Host "   Invoke-WebRequest -Uri 'http://15.134.38.118:8080/payments/webhook' -Method POST -UseBasicParsing" -ForegroundColor Gray
Write-Host ""
Write-Host "4. You should see 200 OK instead of 400!" -ForegroundColor Green
Write-Host ""
Write-Host "5. Then try adding the webhook URL in PayOS again!" -ForegroundColor Green
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan

# Offer to test automatically
Write-Host ""
$test = Read-Host "Do you want me to wait 3 minutes and test automatically? (y/n)"
if ($test -eq "y") {
    Write-Host ""
    Write-Host "Waiting for GitHub Actions deployment..." -ForegroundColor Yellow
    Write-Host "This will take about 2-3 minutes..." -ForegroundColor Gray
    
    for ($i = 1; $i -le 18; $i++) {
        Write-Host "." -NoNewline -ForegroundColor Gray
        Start-Sleep -Seconds 10
    }
    
    Write-Host ""
    Write-Host ""
    Write-Host "Testing webhook endpoint..." -ForegroundColor Yellow
    
    try {
        $response = Invoke-WebRequest -Uri "http://15.134.38.118:8080/payments/webhook" -Method POST -UseBasicParsing
        Write-Host "SUCCESS! Status Code: $($response.StatusCode)" -ForegroundColor Green
        Write-Host "Response: $($response.Content)" -ForegroundColor Green
        Write-Host ""
        Write-Host "✅ Webhook is ready! Go add it to PayOS now!" -ForegroundColor Green
    } catch {
        Write-Host "Status: $($_.Exception.Response.StatusCode.Value__)" -ForegroundColor Yellow
        if ($_.Exception.Response.StatusCode.Value__ -eq 400) {
            Write-Host "Still getting 400 - deployment might not be complete yet." -ForegroundColor Yellow
            Write-Host "Try manually updating on the server with the SSH command above." -ForegroundColor Yellow
        }
    }
}
