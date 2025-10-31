# Whisko


## Overview

(Putting this here since idk what to write)
- **CQRS**: This pattern separates the read and write operations of the application. Commands are responsible for changing the state, while queries are responsible for reading the state.
- **Event Sourcing**: Instead of storing just the current state of an entity, this pattern stores a sequence of events that led to the current state. This allows for better traceability and the ability to reconstruct past states.

## 🎯 What's This?

This is not a complete application...yet. For now, it provides:

- ✅ **CQRS Architecture**: Separate command and query responsibilities
- ✅ **Event Sourcing**: All changes stored as events for complete audit trail
- ¿? **MongoDB Integration**: Event store and read projections
- ✅ **Basic User Entity**: Simple example to demonstrate the pattern

## 🏗️ Architecture

```
┌────────────────────────────┐    ┌────────────────────────────┐
│        Commands            │    │          Queries           │
│      (Write Side)          │    │        (Read Side)         │
├────────────────────────────┤    ├────────────────────────────┤
│ • Create User              │    │ • Get User                 │
│ • Update User Profile      │    │ • List Users               │
│ • Update User Contact      │    │ • Search Users             │
│ • Delete User              │    │                            │
└──────────────┬─────────────┘    └──────────────┬─────────────┘
               │                                  │
               ▼                                  ▼
      ┌──────────────────────┐           ┌──────────────────────┐
      │   Event Store        │           │   Projection/Read    │
      │ (MongoDB/In-Memory)  │           │   Repository         │
      └──────────────────────┘           │ (MongoDB/In-Memory)  │
               │                         └──────────────────────┘
               ▼
      ┌──────────────────────┐
      │   Aggregates         │
      │   (User, etc.)       │
      └──────────────────────┘
```

## 📁 Project Structure

```
whisko-petcare/
├── cmd/
│   └── api/
│       └── main.go                  # Application entry point (HTTP server)
├── internal/
│   ├── application/
│   │   ├── command/                 # Command handlers (write operations)
│   │   ├── query/                   # Query handlers (read operations)
│   │   └── services/                # Application services (orchestrators)
│   ├── domain/
│   │   ├── aggregate/               # Aggregates (User, etc.)
│   │   ├── event/                   # Domain events
│   │   └── repository/              # Repository interfaces
│   ├── infrastructure/
│   │   ├── bus/                     # Event bus implementation
│   │   ├── eventstore/              # Event store (in-memory/MongoDB)
│   │   ├── http/                    # HTTP controllers
│   │   ├── mongo/                   # MongoDB read model repo
│   │   └── projection/              # Read model projections
├── pkg/
│   ├── errors/                      # Shared error types
│   └── middleware/                  # HTTP middleware
├── examples/                        # Example/demo scripts
├── deployments/                     # Docker, k8s, CI/CD files
├── go.mod
├── go.sum
└── README.md

```

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- MongoDB Atlas account (or local MongoDB)
- **Cloudinary account** (for image uploads) - [Sign up free](https://cloudinary.com/)
- PayOS account (for payments) - [Sign up](https://payos.vn/)

### Run the Application

1. **Clone and setup**:
```bash
git clone <your-repo>
cd whisko-petcare
go mod tidy
```

2. **Configure environment**:
```bash
cp .env.example .env
# Edit .env with your credentials:
# - MongoDB connection string
# - Cloudinary credentials (Cloud Name, API Key, API Secret)
# - PayOS credentials
# - JWT secret key
```

3. **Start the application**:
```bash
go run cmd/main.go
```

4. **Or with Docker**:
```bash
docker-compose up
```

The server starts on `http://localhost:8080`

## 📚 API Examples

### Create a User (with JSON)
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Markus Marcaroni", "email": "nononon@example.com"}'
```

### Create a User with Image (Multipart)
```bash
curl -X POST http://localhost:8080/users \
  -F "name=John Doe" \
  -F "email=john@example.com" \
  -F "phone=+1234567890" \
  -F "image=@/path/to/avatar.jpg"
```
**Response:**
```json
{
  "message": "User created successfully",
  "user_id": "user_1234567890",
  "image_url": "https://res.cloudinary.com/dys4wwi0j/image/upload/v.../WhiskoImages/avatars/user_1234567890.jpg"
}
```

### Create a Pet with Image
```bash
curl -X POST http://localhost:8080/pets \
  -F "user_id=user_123" \
  -F "name=Fluffy" \
  -F "species=cat" \
  -F "breed=Persian" \
  -F "age=3" \
  -F "weight=4.5" \
  -F "image=@pet-photo.jpg"
```

### Create a Vendor with Logo
```bash
curl -X POST http://localhost:8080/vendors \
  -F "name=Pet Care Clinic" \
  -F "email=contact@petcare.com" \
  -F "phone=+1234567890" \
  -F "image=@logo.png"
```

### Create a Service with Image
```bash
curl -X POST http://localhost:8080/services \
  -F "vendor_id=vendor_123" \
  -F "name=Dog Grooming" \
  -F "description=Full service grooming" \
  -F "price=5000" \
  -F "duration_minutes=60" \
  -F "tags=grooming,bathing" \
  -F "image=@service-photo.jpg"
```

**📖 For complete API documentation, see [SINGLE_CALL_IMAGE_UPLOAD.md](docs/SINGLE_CALL_IMAGE_UPLOAD.md)**

### 🧪 Postman Collection for Testing

For faster API testing with automated variable management, use the Postman collection:

- **[Import Collection & Environment](postman/)** - Includes 20+ pre-configured requests
- **Features**: Auto-saves IDs, realistic test data, automated tests
- **Quick Start**: Import both JSON files → Select environment → Run collection

See [postman/README.md](postman/README.md) for detailed usage instructions.

## 💡 Key Concepts

- **Commands**: Change state and emit events
- **Events**: Immutable facts about what happened
- **Projections**: Optimized read models built from events
- **Aggregates**: Consistency boundaries (User in this example)

## 🔧 Configuration

Key environment variables:

```bash
# Server
PORT=8080

# MongoDB
MONGO_URI=mongodb+srv://...
MONGO_DATABASE=cqrs_eventsourcing

# JWT
JWT_SECRET_KEY=your-secret-key-min-32-chars
JWT_TOKEN_DURATION=24h

# Cloudinary (REQUIRED for image uploads)
CLOUDINARY_CLOUD_NAME=your-cloud-name
CLOUDINARY_API_KEY=your-api-key
CLOUDINARY_API_SECRET=your-api-secret
CLOUDINARY_FOLDER=WhiskoImages

# PayOS (for payments)
PAYOS_CLIENT_ID=your-client-id
PAYOS_API_KEY=your-api-key
PAYOS_CHECKSUM_KEY=your-checksum-key
PORT=8080                              # Server port
```

## ✨ Features

- ✅ **CQRS + Event Sourcing**: Complete implementation with MongoDB
- ✅ **Image Upload**: Single-call entity creation with images via Cloudinary
- ✅ **Payment Integration**: PayOS payment gateway support
- ✅ **Authentication**: JWT-based auth system
- ✅ **Multi-Entity Support**: Users, Pets, Vendors, Services, Schedules, Vendor Staff
- ✅ **Docker Ready**: Full containerization with docker-compose
- ✅ **API Documentation**: Comprehensive endpoint documentation

## 📖 Documentation

- **[Single-Call Image Upload](docs/SINGLE_CALL_IMAGE_UPLOAD.md)** - Image upload API guide
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Complete deployment instructions
- **[Deployment Checklist](DEPLOYMENT_CHECKLIST.md)** - Step-by-step deployment checklist
- **[API Response System](docs/API_RESPONSE_SYSTEM.md)** - Response format documentation
- **[PayOS Integration](docs/PAYOS_INTEGRATION.md)** - Payment integration guide

## 🚢 Deployment

See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for complete deployment instructions.

**Quick Deploy:**
```bash
# 1. Set up environment
cp .env.example .env
# Edit .env with your Cloudinary, MongoDB, and PayOS credentials

# 2. Deploy with Docker
cd deployments
docker-compose up -d --build

# 3. Verify
curl http://localhost:8080/health
```

**Important:** Cloudinary credentials are **REQUIRED** for image upload functionality. Get them from https://cloudinary.com/
- More complex queries and projections
- Event handlers for side effects
- Testing

## 🧪 Testing

```bash
go test ./...
```

## 🐳 Docker (aint done yet)

```bash
# Development
docker-compose up

# Production build
docker build -t whisko-petcare .
docker run -p 8080:8080 whisko-petcare
```

## 📄 License

MIT License                   # Documentation for the project
```