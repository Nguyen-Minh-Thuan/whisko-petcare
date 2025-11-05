# Emergency Webhook Debug Script
# Run this on SERVER to diagnose real payment issue

Write-Host "============================================" -ForegroundColor Red
Write-Host "  REAL PAYMENT DEBUG - Run on Server!" -ForegroundColor Red
Write-Host "============================================" -ForegroundColor Red
Write-Host ""

Write-Host "SSH into server first:" -ForegroundColor Yellow
Write-Host "  ssh ubuntu@15.134.38.118" -ForegroundColor Gray
Write-Host ""

Write-Host "Then run these commands ONE BY ONE:" -ForegroundColor Yellow
Write-Host ""

Write-Host "# 1. Go to deployments folder" -ForegroundColor Cyan
Write-Host "cd /home/ubuntu/whisko-petcare/deployments" -ForegroundColor White
Write-Host ""

Write-Host "# 2. Check if webhook was even called" -ForegroundColor Cyan
Write-Host "docker-compose logs api | grep 'WEBHOOK RECEIVED' | tail -10" -ForegroundColor White
Write-Host ""

Write-Host "# 3. If webhook WAS called, check what happened" -ForegroundColor Cyan
Write-Host "docker-compose logs api | grep -A 50 'WEBHOOK RECEIVED' | tail -100" -ForegroundColor White
Write-Host ""

Write-Host "# 4. Check for payment confirmation" -ForegroundColor Cyan
Write-Host "docker-compose logs api | grep -A 30 'ConfirmPaymentHandler' | tail -100" -ForegroundColor White
Write-Host ""

Write-Host "# 5. Check for schedule creation attempt" -ForegroundColor Cyan
Write-Host "docker-compose logs api | grep 'Auto-creating schedule' | tail -10" -ForegroundColor White
Write-Host ""

Write-Host "# 6. Check for ANY errors in last 100 lines" -ForegroundColor Cyan
Write-Host "docker-compose logs api --tail=100 | grep -E '(ERROR|FAILED|❌|Failed)'" -ForegroundColor White
Write-Host ""

Write-Host "# 7. Get ALL recent logs (last 200 lines)" -ForegroundColor Cyan
Write-Host "docker-compose logs api --tail=200" -ForegroundColor White
Write-Host ""

Write-Host "============================================" -ForegroundColor Red
Write-Host "  Copy the output and send it to me!" -ForegroundColor Red
Write-Host "============================================" -ForegroundColor Red
Write-Host ""

Write-Host "IMPORTANT: Look for these specific things:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Was webhook called?" -ForegroundColor White
Write-Host "   Search for: '🔔 WEBHOOK RECEIVED'" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Did it have signature?" -ForegroundColor White
Write-Host "   Search for: '✅ Signature present' OR '❌ Missing signature'" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Was payment found?" -ForegroundColor White
Write-Host "   Search for: '✅ Found payment' OR '❌ Payment not found'" -ForegroundColor Gray
Write-Host ""
Write-Host "4. What was PayOS status?" -ForegroundColor White
Write-Host "   Search for: '💰 PayOS Status:'" -ForegroundColor Gray
Write-Host ""
Write-Host "5. Did it try to create schedule?" -ForegroundColor White
Write-Host "   Search for: '📅 Auto-creating schedule'" -ForegroundColor Gray
Write-Host ""
Write-Host "6. Any error messages?" -ForegroundColor White
Write-Host "   Search for: 'ERROR', 'FAILED', '❌'" -ForegroundColor Gray
Write-Host ""
