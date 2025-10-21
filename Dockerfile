# Multi-stage Dockerfile for RabbitMQ Stream Viewer

# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY web/ ./

# Build frontend
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Build the publisher binary (for testing)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o publisher ./cmd/publisher

# Stage 3: Final minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binaries from builder
COPY --from=backend-builder /app/server .
COPY --from=backend-builder /app/publisher .

# Copy built frontend to the static directory
COPY --from=frontend-builder /app/web/dist ./web/dist

# Expose server port
EXPOSE 8080

# Run the server
CMD ["./server", "-config", "/config/config.yaml"]

