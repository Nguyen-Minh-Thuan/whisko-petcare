# Deploy Webhook Fix to Server
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Deploying Webhook Fix" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Step 1: Commit and push code" -ForegroundColor Yellow
Write-Host "Running: git add, commit, push..." -ForegroundColor Gray
git add internal/infrastructure/http/http-payment-controller.go
git commit -m "Fix: Process PayOS webhooks without signature - schedules will now be created"
git push origin main

Write-Host ""
Write-Host "✅ Code pushed to GitHub!" -ForegroundColor Green
Write-Host ""

Write-Host "Step 2: Deploy to server" -ForegroundColor Yellow
Write-Host "Now SSH into server and run these commands:" -ForegroundColor White
Write-Host ""
Write-Host "  ssh ubuntu@15.134.38.118" -ForegroundColor Gray
Write-Host "  cd /home/ubuntu/whisko-petcare && git pull origin main" -ForegroundColor Gray
Write-Host "  cd deployments && docker-compose down && docker-compose up -d --build" -ForegroundColor Gray
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "What this fix does:" -ForegroundColor Yellow
Write-Host "- Processes webhooks EVEN WITHOUT signature" -ForegroundColor White
Write-Host "- Extracts orderCode from payload" -ForegroundColor White
Write-Host "- Confirms payment with PayOS" -ForegroundColor White
Write-Host "- AUTO-CREATES SCHEDULES! ✅" -ForegroundColor Green
Write-Host ""
Write-Host "Next payment will automatically create schedules!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Cyan
