# ============================================================================
# Stage 1: Build Frontend
# ============================================================================
FROM node:22-alpine AS frontend-builder

WORKDIR /app/client

# Copy package files
COPY client/package*.json ./
RUN npm ci

# Copy frontend source
COPY client/ ./

# Create output directory for build
RUN mkdir -p /app/static

# Build frontend (outputs to /app/static/dist/)
RUN npm run build

# ============================================================================
# Stage 2: Build Backend
# ============================================================================
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /backgammon-server .

# ============================================================================
# Stage 3: Production Runtime
# ============================================================================
FROM alpine:latest

WORKDIR /root/

# Copy the Go binary from backend builder
COPY --from=backend-builder /backgammon-server .

# Copy the built frontend from frontend builder to static/
COPY --from=frontend-builder /app/static/dist ./static/dist

# Expose port
EXPOSE 8080

# Run the application
CMD ["./backgammon-server"]
