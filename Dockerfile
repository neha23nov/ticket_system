# =============================================================================
# Stage 1: Build the Go application
# =============================================================================
# We use the Alpine variant of the Go image because it's smaller.
# We use the pure-Go SQLite driver, so no C compiler (CGo) is needed.
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency files first. Docker caches this layer, so dependencies
# are only re-downloaded when go.mod or go.sum change — not on every code change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build the Go binary. CGO_ENABLED=0 produces a fully static binary
# since we use a pure-Go SQLite driver (no C dependencies).
RUN CGO_ENABLED=0 go build -o ticket-system ./cmd/server

# =============================================================================
# Stage 2: Create the minimal runtime image
# =============================================================================
# We use a plain Alpine image (much smaller than the Go build image)
# because we only need the compiled binary to run.
FROM alpine:3.19

# Install CA certificates so the application can make HTTPS requests if needed.
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the compiled binary from the builder stage.
COPY --from=builder /app/ticket-system .

# Copy the frontend static files (HTML, CSS, JavaScript).
COPY --from=builder /app/web ./web

# Create the data directory where the SQLite database file will be stored.
RUN mkdir -p /app/data

# Document that the application listens on port 8080.
EXPOSE 8080

# Run the application when the container starts.
CMD ["./ticket-system"]
