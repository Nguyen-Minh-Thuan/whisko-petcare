#!/bin/bash

# Update Vendor Bank Account
# This script updates the bank account information for vendor John's Pet Services

VENDOR_ID="69f50647-4055-4189-8b54-71932e929f88"
API_URL="https://api.whisko.shop/api/vendors/$VENDOR_ID/bank-account"

echo "=================================="
echo "Updating Vendor Bank Account"
echo "=================================="
echo "Vendor ID: $VENDOR_ID"
echo "Bank: TPBank (BIN: 970423)"
echo "Account: 31404112004"
echo "Name: NGUYEN QUY HUNG"
echo "Branch: Ho Chi Minh"
echo ""
echo "Sending request to: $API_URL"
echo ""

# Make the API call
curl -X PUT "$API_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "bank_name": "TPBank",
    "account_number": "31404112004",
    "account_name": "NGUYEN QUY HUNG",
    "bank_branch": "Ho Chi Minh"
  }'

echo ""
echo ""
echo "=================================="
echo ""
echo "Note: This endpoint does NOT require authentication"
echo "The bank account is now configured for automatic payouts"
echo ""
echo "Bank Code Mapping:"
echo "  TPBank → BIN: 970423"
echo ""
echo "When a user pays for this vendor's service:"
echo "  1. Payment confirmed"
echo "  2. Payout automatically created"
echo "  3. Money transferred to: TPBank - 31404112004"
echo "  4. Account holder: NGUYEN QUY HUNG"
echo "=================================="
