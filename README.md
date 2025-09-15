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
- Go 1.18+
- MongoDB (optional - has in-memory fallback)

### Run the Application

1. **Clone and setup**:
```bash
git clone <your-repo>
cd whisko-petcare
go mod tidy
```

2. **Configure (optional)**:
```bash
cp .env.example .env
# Edit .env if using MongoDB
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

### Create a User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Markus Marcaroni", "email": "nononon@example.com"}'
```

### Get a User
```bash
curl http://localhost:8080/api/v1/users/{user-id}
```

### List Users
```bash
curl http://localhost:8080/api/v1/users
```

### Update a User
```bash
curl -X PUT http://localhost:8080/api/v1/users/{user-id} \
  -H "Content-Type: application/json" \
  -d '{"name": "The World", "email": "you@example.com"}'
```

## 💡 Key Concepts

- **Commands**: Change state and emit events
- **Events**: Immutable facts about what happened
- **Projections**: Optimized read models built from events
- **Aggregates**: Consistency boundaries (User in this example)

## 🔧 Configuration

Environment variables:

```bash
MONGODB_URI=mongodb://localhost:27017  # MongoDB connection
DATABASE_NAME=whisko                   # Database name
PORT=8080                              # Server port
```

## 📈 What's Next?

This base provides:
- ✅ CQRS pattern implementation
- ✅ Event sourcing foundation
- ✅ MongoDB integration (kinda :D)
- ✅ Basic API structure
- ✅ Docker setup

**To-do**:
- Your specific domain models (Pet, Booking, Service, etc.)
- Business logic and validation
- Authentication & authorization
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