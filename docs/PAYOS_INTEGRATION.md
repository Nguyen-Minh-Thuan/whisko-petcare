# PayOS Payment Integration Guide

This guide explains how to configure and use the PayOS payment integration in your Whisko Pet Care application.

## Prerequisites

1. **PayOS Account**: You need a PayOS merchant account with API credentials
2. **MongoDB**: Running MongoDB instance for data persistence
3. **Go 1.18+**: Required for running the application

## Configuration

### 1. Environment Variables

Copy `.env.example` to `.env` and configure your PayOS credentials:

```bash
cp .env.example .env
```

Update the PayOS configuration in `.env`:

```env
# PayOS Configuration
PAYOS_CLIENT_ID=your_actual_client_id
PAYOS_API_KEY=your_actual_api_key
PAYOS_CHECKSUM_KEY=your_actual_checksum_key
PAYOS_PARTNER_CODE=your_actual_partner_code
PAYOS_BASE_URL=https://api-merchant.payos.vn
PAYOS_WEBHOOK_URL=http://your-domain.com/payments/webhook
PAYOS_RETURN_URL=http://your-domain.com/payments/return
PAYOS_CANCEL_URL=http://your-domain.com/payments/cancel
```

### 2. PayOS Account Setup

1. Log into your PayOS merchant dashboard
2. Navigate to API settings
3. Copy your credentials:
   - Client ID
   - API Key
   - Checksum Key
   - Partner Code

## API Endpoints

### Payment Creation

**POST** `/payments`

Create a new payment request:

```json
{
  "user_id": "user123",
  "amount": 100000,
  "description": "Pet food order",
  "items": [
    {
      "name": "Premium Dog Food",
      "quantity": 2,
      "price": 50000
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "payment_id": "uuid-payment-id",
    "order_code": 1234567890,
    "checkout_url": "https://pay.payos.vn/...",
    "qr_code": "data:image/png;base64,...",
    "amount": 100000,
    "status": "PENDING",
    "expired_at": "2024-01-01T15:15:00Z"
  }
}
```

### Payment Retrieval

**GET** `/payments/{payment_id}`

Get payment details by payment ID.

**GET** `/payments/order/{order_code}`

Get payment details by order code.

**GET** `/payments/user/{user_id}?offset=0&limit=10`

List payments for a specific user with pagination.

### Payment Management

**PUT** `/payments/{payment_id}/cancel`

Cancel a pending payment:

```json
{
  "reason": "Customer requested cancellation"
}
```

## Webhook Integration

PayOS will send webhook notifications to `/payments/webhook` when payment status changes.

### Webhook Security

The webhook handler verifies the signature sent by PayOS to ensure authenticity. Configure your webhook URL in the PayOS dashboard.

## Payment Flow

1. **Create Payment**: Client calls `POST /payments` with payment details
2. **Redirect to PayOS**: Client redirects user to the `checkout_url`
3. **User Payment**: User completes payment on PayOS platform
4. **Webhook Notification**: PayOS sends webhook to your server
5. **Payment Confirmation**: Your system updates payment status
6. **Return to App**: User is redirected to `return_url` or `cancel_url`

## Payment Statuses

- `PENDING`: Payment created, awaiting payment
- `PAID`: Payment completed successfully
- `CANCELLED`: Payment was cancelled
- `EXPIRED`: Payment expired (default 15 minutes)
- `FAILED`: Payment failed

## Return URLs

### Success Return
`/payments/return?orderCode=1234567890`

Displays a success page with payment details.

### Cancel Return
`/payments/cancel?orderCode=1234567890`

Displays a cancellation page and marks payment as cancelled.

## Testing

### 1. Start the Application

```bash
go run cmd/api/main.go
```

### 2. Create a Test Payment

```bash
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user",
    "amount": 50000,
    "description": "Test payment",
    "items": [
      {
        "name": "Test Item",
        "quantity": 1,
        "price": 50000
      }
    ]
  }'
```

### 3. Test Payment Retrieval

```bash
curl http://localhost:8080/payments/{payment_id}
```

## Error Handling

The API uses standardized error responses:

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "Amount must be greater than 0",
    "details": null
  },
  "timestamp": "2024-01-01T10:00:00Z"
}
```

## MongoDB Collections

The integration creates these MongoDB collections:

- `payments`: Payment aggregate data
- `payment_events`: Event sourcing events for payments

## Architecture

The PayOS integration follows Domain-Driven Design (DDD) principles:

- **Domain Layer**: Payment aggregate with business logic
- **Application Layer**: Command/Query handlers for payment operations
- **Infrastructure Layer**: PayOS HTTP client, MongoDB repository
- **Presentation Layer**: HTTP controllers and REST API

## Security Considerations

1. **API Keys**: Never expose PayOS credentials in client-side code
2. **Webhook Verification**: Always verify webhook signatures
3. **HTTPS**: Use HTTPS in production for all PayOS communications
4. **Environment Variables**: Store sensitive configuration in environment variables

## Production Deployment

1. Update webhook URLs to your production domain
2. Use HTTPS for all URLs
3. Configure proper MongoDB connection string
4. Set appropriate timeouts and retry policies
5. Implement proper logging and monitoring

## Support

For PayOS-specific issues, consult the [PayOS Documentation](https://docs.payos.vn).

For integration issues, check the application logs and ensure all configuration is correct.