# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /backgammon-server .

# Runtime stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /backgammon-server .

# Copy static files
COPY --from=builder /app/static ./static

# Expose port
EXPOSE 8080

# Run the application
CMD ["./backgammon-server"]
