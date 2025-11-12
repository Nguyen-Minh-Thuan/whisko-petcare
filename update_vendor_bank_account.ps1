# Update Vendor Bank Account
# This script updates the bank account information for vendor John's Pet Services

# Vendor ID
$vendorId = "69f50647-4055-4189-8b54-71932e929f88"

# API endpoint
$apiUrl = "https://api.whisko.shop/api/vendors/$vendorId/bank-account"

# Bank account information (corrected)
$bankAccountData = @{
    bank_name = "TPBank"
    account_number = "31404112004"
    account_name = "NGUYEN QUY HUNG"
    bank_branch = "Ho Chi Minh"  # Fixed typo: "branh" -> "Ho Chi Minh"
} | ConvertTo-Json

Write-Host "=================================="
Write-Host "Updating Vendor Bank Account"
Write-Host "=================================="
Write-Host "Vendor ID: $vendorId"
Write-Host "Bank: TPBank (BIN: 970423)"
Write-Host "Account: 31404112004"
Write-Host "Name: NGUYEN QUY HUNG"
Write-Host "Branch: Ho Chi Minh"
Write-Host ""

# Make the API call
Write-Host "Sending request to: $apiUrl"
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri $apiUrl -Method PUT -Body $bankAccountData -ContentType "application/json"
    
    Write-Host "✅ SUCCESS!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Response:"
    $response | ConvertTo-Json -Depth 10
} catch {
    Write-Host "❌ ERROR!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Error Details:"
    Write-Host $_.Exception.Message
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host ""
        Write-Host "Response Body:"
        Write-Host $responseBody
    }
}

Write-Host ""
Write-Host "=================================="
Write-Host ""
Write-Host "Note: This endpoint does NOT require authentication"
Write-Host "The bank account is now configured for automatic payouts"
Write-Host ""
Write-Host "Bank Code Mapping:"
Write-Host "  TPBank → BIN: 970423"
Write-Host ""
Write-Host "When a user pays for this vendor's service:"
Write-Host "  1. Payment confirmed"
Write-Host "  2. Payout automatically created"
Write-Host "  3. Money transferred to: TPBank - 31404112004"
Write-Host "  4. Account holder: NGUYEN QUY HUNG"
Write-Host "=================================="
