# Webhook Diagnostic Script
# This script helps debug why schedules aren't being created

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Webhook Schedule Creation Diagnostic" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "IMPORTANT: Run this on your SERVER via SSH" -ForegroundColor Yellow
Write-Host "ssh ubuntu@15.134.38.118" -ForegroundColor Gray
Write-Host ""

Write-Host "Then run these commands to check logs:" -ForegroundColor White
Write-Host ""

Write-Host "# 1. Check if webhook was called" -ForegroundColor Green
Write-Host "cd /home/ubuntu/whisko-petcare/deployments" -ForegroundColor Gray
Write-Host "docker-compose logs api | grep -i 'WEBHOOK RECEIVED'" -ForegroundColor Gray
Write-Host ""

Write-Host "# 2. Check if payment was confirmed" -ForegroundColor Green
Write-Host "docker-compose logs api | grep -i 'ConfirmPaymentHandler'" -ForegroundColor Gray
Write-Host ""

Write-Host "# 3. Check if schedule creation was attempted" -ForegroundColor Green
Write-Host "docker-compose logs api | grep -i 'Auto-creating schedule'" -ForegroundColor Gray
Write-Host ""

Write-Host "# 4. Check for any errors" -ForegroundColor Green
Write-Host "docker-compose logs api | grep -E '(ERROR|FAILED|❌)' | tail -20" -ForegroundColor Gray
Write-Host ""

Write-Host "# 5. Watch logs in real-time (do this before making a test payment)" -ForegroundColor Green
Write-Host "docker-compose logs -f api" -ForegroundColor Gray
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Common Issues & Solutions" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Issue 1: Webhook not being called" -ForegroundColor Yellow
Write-Host "  - Check if webhook URL is configured in PayOS dashboard" -ForegroundColor White
Write-Host "  - URL should be: http://15.134.38.118:8080/payments/webhook" -ForegroundColor Gray
Write-Host ""

Write-Host "Issue 2: Payment not found" -ForegroundColor Yellow
Write-Host "  - Make sure payment was created before PayOS webhook fires" -ForegroundColor White
Write-Host "  - Check if orderCode matches between payment and webhook" -ForegroundColor White
Write-Host ""

Write-Host "Issue 3: Schedule creation fails" -ForegroundColor Yellow
Write-Host "  - Check if all required fields exist in payment:" -ForegroundColor White
Write-Host "    * UserID, VendorID, PetID" -ForegroundColor Gray
Write-Host "    * ServiceIDs array" -ForegroundColor Gray
Write-Host "    * StartTime, EndTime" -ForegroundColor Gray
Write-Host ""

Write-Host "Issue 4: createScheduleHandler is nil" -ForegroundColor Yellow
Write-Host "  - This means the handler wasn't initialized properly" -ForegroundColor White
Write-Host "  - Check server startup logs" -ForegroundColor White
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Test Workflow" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "1. Start watching logs:" -ForegroundColor Green
Write-Host "   ssh ubuntu@15.134.38.118" -ForegroundColor Gray
Write-Host "   cd /home/ubuntu/whisko-petcare/deployments" -ForegroundColor Gray
Write-Host "   docker-compose logs -f api" -ForegroundColor Gray
Write-Host ""

Write-Host "2. Make a test payment from your app" -ForegroundColor Green
Write-Host ""

Write-Host "3. Complete the payment on PayOS checkout page" -ForegroundColor Green
Write-Host ""

Write-Host "4. Watch the logs - you should see:" -ForegroundColor Green
Write-Host "   ✅ 🔔 WEBHOOK RECEIVED from PayOS" -ForegroundColor White
Write-Host "   ✅ ✅ Signature present" -ForegroundColor White
Write-Host "   ✅ 🔔 ConfirmPaymentHandler: Processing order code" -ForegroundColor White
Write-Host "   ✅ ✅ Found payment" -ForegroundColor White
Write-Host "   ✅ 💰 PayOS Status: PAID" -ForegroundColor White
Write-Host "   ✅ 📅 Auto-creating schedule" -ForegroundColor White
Write-Host "   ✅ ✅ Successfully auto-created schedule!" -ForegroundColor White
Write-Host ""

Write-Host "5. If you see errors, note them and we'll fix them!" -ForegroundColor Yellow
Write-Host ""
