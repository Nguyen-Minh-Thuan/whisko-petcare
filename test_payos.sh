#!/bin/bash

# PayOS Integration Test Script
# This script demonstrates how to test the PayOS payment integration

API_BASE="http://localhost:8080"

echo "=== PayOS Payment Integration Test ==="
echo ""

# Test 1: Health Check
echo "1. Testing health endpoint..."
curl -s "${API_BASE}/health" | jq '.'
echo ""

# Test 2: Create a test payment
echo "2. Creating a test payment..."
PAYMENT_RESPONSE=$(curl -s -X POST "${API_BASE}/payments" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user-123",
    "amount": 100000,
    "description": "Test payment for pet supplies",
    "items": [
      {
        "name": "Premium Dog Food",
        "quantity": 1,
        "price": 75000
      },
      {
        "name": "Dog Toy",
        "quantity": 1,
        "price": 25000
      }
    ]
  }')

echo "Payment Creation Response:"
echo "$PAYMENT_RESPONSE" | jq '.'
echo ""

# Extract payment ID and order code for further tests
PAYMENT_ID=$(echo "$PAYMENT_RESPONSE" | jq -r '.data.payment_id // empty')
ORDER_CODE=$(echo "$PAYMENT_RESPONSE" | jq -r '.data.order_code // empty')

if [ -n "$PAYMENT_ID" ]; then
  echo "Created payment ID: $PAYMENT_ID"
  echo "Order code: $ORDER_CODE"
  echo ""

  # Test 3: Get payment by ID
  echo "3. Retrieving payment by ID..."
  curl -s "${API_BASE}/payments/${PAYMENT_ID}" | jq '.'
  echo ""

  # Test 4: Get payment by order code
  if [ -n "$ORDER_CODE" ]; then
    echo "4. Retrieving payment by order code..."
    curl -s "${API_BASE}/payments/order/${ORDER_CODE}" | jq '.'
    echo ""
  fi

  # Test 5: List user payments
  echo "5. Listing payments for user..."
  curl -s "${API_BASE}/payments/user/test-user-123?limit=5" | jq '.'
  echo ""

  # Test 6: Cancel payment
  echo "6. Cancelling payment..."
  curl -s -X PUT "${API_BASE}/payments/${PAYMENT_ID}/cancel" \
    -H "Content-Type: application/json" \
    -d '{"reason": "Test cancellation"}' | jq '.'
  echo ""

  # Test 7: Check payment status after cancellation
  echo "7. Checking payment status after cancellation..."
  curl -s "${API_BASE}/payments/${PAYMENT_ID}" | jq '.'
  echo ""

else
  echo "❌ Failed to create payment. Check your configuration."
  echo "Make sure:"
  echo "- MongoDB is running"
  echo "- PayOS credentials are configured in .env"
  echo "- Application is running on port 8080"
fi

echo "=== Test completed ==="
echo ""
echo "Next steps:"
echo "1. Configure PayOS credentials in .env file"
echo "2. Test with real PayOS sandbox environment"
echo "3. Test webhook integration with ngrok or similar"