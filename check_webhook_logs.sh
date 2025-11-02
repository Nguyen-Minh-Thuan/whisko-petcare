#!/bin/bash
# Run this script on your SERVER (via SSH)
# ssh ubuntu@15.134.38.118

echo "============================================"
echo "  Checking Webhook Logs for Real Payment"
echo "============================================"
echo ""

cd /home/ubuntu/whisko-petcare/deployments

echo "1. Checking if webhook was called..."
echo "-------------------------------------------"
docker-compose logs api | grep "WEBHOOK RECEIVED" | tail -10
echo ""

echo "2. Checking last webhook attempt (with context)..."
echo "-------------------------------------------"
docker-compose logs api | grep -A 30 "WEBHOOK RECEIVED" | tail -50
echo ""

echo "3. Checking for payment confirmation..."
echo "-------------------------------------------"
docker-compose logs api | grep "ConfirmPaymentHandler" | tail -10
echo ""

echo "4. Checking for schedule creation attempts..."
echo "-------------------------------------------"
docker-compose logs api | grep -E "(Auto-creating schedule|Successfully auto-created schedule|Failed to auto-create schedule)" | tail -10
echo ""

echo "5. Checking for errors..."
echo "-------------------------------------------"
docker-compose logs api --tail=100 | grep -E "(ERROR|FAILED|❌|panic|fatal)"
echo ""

echo "6. Last 50 lines of logs..."
echo "-------------------------------------------"
docker-compose logs api --tail=50
echo ""

echo "============================================"
echo "  Analysis Complete!"
echo "============================================"
