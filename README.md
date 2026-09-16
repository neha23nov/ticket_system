# Ticket System — Backend API

A REST API backend service built in **Go** for managing support tickets. Users can register, log in, create tickets, view their own tickets, and update ticket statuses.

## Features

- **User Authentication** — Register and login with JWT-based authentication.
- **Ticket Management** — Create, list, view, and update tickets.
- **Ownership Enforcement** — Users can only access their own tickets.
- **Status Workflow** — Tickets follow a strict lifecycle: `open → in_progress → closed`.
- **Secure Passwords** — All passwords are hashed using bcrypt (never stored as plain text).
- **Dockerized** — Ready for containerized deployment.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.22 |
| Router | [chi](https://github.com/go-chi/chi) |
| Database | SQLite via [GORM](https://gorm.io/) |
| Auth | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) |
| Passwords | bcrypt ([x/crypto](https://pkg.go.dev/golang.org/x/crypto/bcrypt)) |

## API Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|:------------:|
| `GET` | `/health` | Health check | ❌ |
| `POST` | `/auth/register` | Register a new user | ❌ |
| `POST` | `/auth/login` | Login and get JWT token | ❌ |
| `POST` | `/tickets` | Create a new ticket | ✅ |
| `GET` | `/tickets` | List your tickets | ✅ |
| `GET` | `/tickets/{id}` | Get a specific ticket | ✅ |
| `PATCH` | `/tickets/{id}/status` | Update ticket status | ✅ |

## Ticket Status Flow

```
open  →  in_progress  →  closed
```

- A newly created ticket starts with status `open`.
- Status can only move **forward** (open → in_progress → closed).
- A **closed** ticket **cannot** be reopened.

## Quick Start

### Run Locally (without Docker)

```bash
# Clone the repository
git clone <your-repo-url>
cd ticket-system

# Install dependencies
go mod download

# Run the server
go run ./cmd/server

# The server starts at http://localhost:8080
```

### Run with Docker

```bash
# Build the Docker image
docker build -t ticket-system .

# Run the container
docker run -p 8080:8080 ticket-system

# Verify it's running
curl http://localhost:8080/health
# Expected: {"status":"ok"}
```


## Project Structure

```
ticket-system/
├── cmd/server/main.go              # Entry point — config, routing, server start
├── internal/
│   ├── models/
│   │   ├── user.go                 # User database model
│   │   └── ticket.go               # Ticket model + status rules
│   ├── database/
│   │   └── database.go             # SQLite connection + migrations
│   ├── auth/
│   │   ├── password.go             # bcrypt password hashing
│   │   └── jwt.go                  # JWT token generation & validation
│   ├── middleware/
│   │   └── auth.go                 # JWT authentication middleware
│   └── handlers/
│       ├── health.go               # GET /health
│       ├── auth_handler.go         # POST /auth/register, /auth/login
│       └── ticket_handler.go       # Ticket CRUD endpoints
├── Dockerfile                      # Multi-stage Docker build
├── .env.example                    # Environment variable template
├── go.mod                          # Go module dependencies
└── README.md                       # This file
```

## Assumptions

1. SQLite is used as the database — data is persisted in a local file. For production, consider PostgreSQL.
2. JWT tokens expire after 24 hours.
3. No rate limiting is implemented (out of scope for this assignment).
4. No admin role or ticket assignment flow.
5. Email format validation is basic (presence check only).


