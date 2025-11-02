# Quick checks for why webhook didn't work with real payment

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Real Payment Troubleshooting Checklist" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "POSSIBLE CAUSES:" -ForegroundColor Yellow
Write-Host ""

Write-Host "1. WEBHOOK URL NOT CONFIGURED IN PAYOS" -ForegroundColor Red
Write-Host "   - You must configure webhook in PayOS dashboard BEFORE payment" -ForegroundColor White
Write-Host "   - Go to: https://payos.vn/" -ForegroundColor Gray
Write-Host "   - Settings -> Webhook Configuration" -ForegroundColor Gray
Write-Host "   - Add: http://15.134.38.118:8080/payments/webhook" -ForegroundColor Gray
Write-Host "   - IF THIS WASN'T DONE, PayOS won't call your webhook!" -ForegroundColor Red
Write-Host ""

Write-Host "2. SERVER WAS DOWN" -ForegroundColor Red
Write-Host "   - Check if server was running when payment was made" -ForegroundColor White
Write-Host "   - Test now:" -ForegroundColor Gray
$testNow = Read-Host "   Test webhook now? (y/n)"
if ($testNow -eq "y") {
    try {
        $response = Invoke-WebRequest -Uri "http://15.134.38.118:8080/health" -Method GET -UseBasicParsing -TimeoutSec 5
        Write-Host "   ✅ Server is UP now (Status: $($response.StatusCode))" -ForegroundColor Green
    } catch {
        Write-Host "   ❌ Server is DOWN now!" -ForegroundColor Red
        Write-Host "   It might have been down during payment!" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "3. LATEST CODE NOT DEPLOYED" -ForegroundColor Red
Write-Host "   - Webhook handler might not be on server" -ForegroundColor White
Write-Host "   - Need to deploy latest code" -ForegroundColor White
Write-Host ""

Write-Host "4. WEBHOOK CALLED BUT FAILED" -ForegroundColor Red
Write-Host "   - Check server logs to see actual error" -ForegroundColor White
Write-Host "   - Run: check_real_payment_logs.ps1" -ForegroundColor Gray
Write-Host ""

Write-Host "5. PAYMENT DATA MISSING REQUIRED FIELDS" -ForegroundColor Red
Write-Host "   - Schedule creation needs:" -ForegroundColor White
Write-Host "     * user_id, vendor_id, pet_id" -ForegroundColor Gray
Write-Host "     * service_ids (array)" -ForegroundColor Gray
Write-Host "     * start_time, end_time" -ForegroundColor Gray
Write-Host "   - If payment missing these, schedule creation will fail" -ForegroundColor White
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  IMMEDIATE ACTIONS" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "ACTION 1: Check PayOS Dashboard" -ForegroundColor Green
Write-Host "  - Go to https://payos.vn/" -ForegroundColor White
Write-Host "  - Login" -ForegroundColor White
Write-Host "  - Check if webhook URL is configured" -ForegroundColor White
Write-Host "  - Check webhook call history (if available)" -ForegroundColor White
Write-Host ""

Write-Host "ACTION 2: Check Server Logs" -ForegroundColor Green
Write-Host "  - SSH: ssh ubuntu@15.134.38.118" -ForegroundColor White
Write-Host "  - Run: cd /home/ubuntu/whisko-petcare/deployments" -ForegroundColor White
Write-Host "  - Run: docker-compose logs api --tail=200" -ForegroundColor White
Write-Host "  - Look for: 'WEBHOOK RECEIVED'" -ForegroundColor White
Write-Host ""

Write-Host "ACTION 3: Update Server Code" -ForegroundColor Green
Write-Host "  - SSH: ssh ubuntu@15.134.38.118" -ForegroundColor White
Write-Host "  - Run: cd /home/ubuntu/whisko-petcare && git pull origin main" -ForegroundColor White
Write-Host "  - Run: cd deployments && docker-compose restart api" -ForegroundColor White
Write-Host ""

Write-Host "ACTION 4: Manual Check Payment Status" -ForegroundColor Green
Write-Host "  - What's your order code from the payment?" -ForegroundColor White
$orderCode = Read-Host "  Enter order code (or press Enter to skip)"
if ($orderCode) {
    Write-Host "  Call this endpoint:" -ForegroundColor White
    Write-Host "  GET http://15.134.38.118:8080/payments/status/$orderCode" -ForegroundColor Gray
    Write-Host "  Authorization: Bearer YOUR_TOKEN" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Then manually trigger schedule creation:" -ForegroundColor White
    Write-Host "  GET http://15.134.38.118:8080/payments/check/$orderCode" -ForegroundColor Gray
    Write-Host "  Authorization: Bearer YOUR_TOKEN" -ForegroundColor Gray
}
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Please do ACTION 1 and ACTION 2 first!" -ForegroundColor Yellow
Write-Host "Then tell me what you find in the logs!" -ForegroundColor Yellow
Write-Host "============================================" -ForegroundColor Cyan
