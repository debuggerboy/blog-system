# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy source code
COPY . .

# Generate templ files
RUN templ generate ./...

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o blog-server cmd/server/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

# Create a non-root user
RUN adduser -D -u 1000 appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/blog-server .

# Copy migrations and create data directory
RUN mkdir -p /app/data /app/migrations
COPY --from=builder /app/migrations /app/migrations

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./blog-server"]
