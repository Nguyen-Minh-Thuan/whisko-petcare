# PayOS Integration Test Script (PowerShell)
# This script demonstrates how to test the PayOS payment integration

$API_BASE = "http://localhost:8080"

Write-Host "=== PayOS Payment Integration Test ===" -ForegroundColor Green
Write-Host ""

# Test 1: Health Check
Write-Host "1. Testing health endpoint..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$API_BASE/health" -Method Get
    $healthResponse | ConvertTo-Json -Depth 10
} catch {
    Write-Host "❌ Health check failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 2: Create a test payment
Write-Host "2. Creating a test payment..." -ForegroundColor Yellow
$paymentData = @{
    user_id = "test-user-123"
    amount = 100000
    description = "Test payment for pet supplies"
    items = @(
        @{
            name = "Premium Dog Food"
            quantity = 1
            price = 75000
        },
        @{
            name = "Dog Toy"
            quantity = 1
            price = 25000
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $paymentResponse = Invoke-RestMethod -Uri "$API_BASE/payments" -Method Post -Body $paymentData -ContentType "application/json"
    Write-Host "Payment Creation Response:" -ForegroundColor Green
    $paymentResponse | ConvertTo-Json -Depth 10
    
    $paymentId = $paymentResponse.data.payment_id
    $orderCode = $paymentResponse.data.order_code
    
    if ($paymentId) {
        Write-Host ""
        Write-Host "Created payment ID: $paymentId" -ForegroundColor Green
        Write-Host "Order code: $orderCode" -ForegroundColor Green
        Write-Host ""

        # Test 3: Get payment by ID
        Write-Host "3. Retrieving payment by ID..." -ForegroundColor Yellow
        $getPaymentResponse = Invoke-RestMethod -Uri "$API_BASE/payments/$paymentId" -Method Get
        $getPaymentResponse | ConvertTo-Json -Depth 10
        Write-Host ""

        # Test 4: Get payment by order code
        if ($orderCode) {
            Write-Host "4. Retrieving payment by order code..." -ForegroundColor Yellow
            $getByOrderResponse = Invoke-RestMethod -Uri "$API_BASE/payments/order/$orderCode" -Method Get
            $getByOrderResponse | ConvertTo-Json -Depth 10
            Write-Host ""
        }

        # Test 5: List user payments
        Write-Host "5. Listing payments for user..." -ForegroundColor Yellow
        $listPaymentsResponse = Invoke-RestMethod -Uri "$API_BASE/payments/user/test-user-123?limit=5" -Method Get
        $listPaymentsResponse | ConvertTo-Json -Depth 10
        Write-Host ""

        # Test 6: Cancel payment
        Write-Host "6. Cancelling payment..." -ForegroundColor Yellow
        $cancelData = @{
            reason = "Test cancellation"
        } | ConvertTo-Json
        
        $cancelResponse = Invoke-RestMethod -Uri "$API_BASE/payments/$paymentId/cancel" -Method Put -Body $cancelData -ContentType "application/json"
        $cancelResponse | ConvertTo-Json -Depth 10
        Write-Host ""

        # Test 7: Check payment status after cancellation
        Write-Host "7. Checking payment status after cancellation..." -ForegroundColor Yellow
        $finalStatusResponse = Invoke-RestMethod -Uri "$API_BASE/payments/$paymentId" -Method Get
        $finalStatusResponse | ConvertTo-Json -Depth 10
        Write-Host ""
    }
} catch {
    Write-Host "❌ Failed to create payment: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Make sure:" -ForegroundColor Yellow
    Write-Host "- MongoDB is running" -ForegroundColor Yellow
    Write-Host "- PayOS credentials are configured in .env" -ForegroundColor Yellow
    Write-Host "- Application is running on port 8080" -ForegroundColor Yellow
}

Write-Host "=== Test completed ===" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Configure PayOS credentials in .env file" -ForegroundColor White
Write-Host "2. Test with real PayOS sandbox environment" -ForegroundColor White
Write-Host "3. Test webhook integration with ngrok or similar" -ForegroundColor White