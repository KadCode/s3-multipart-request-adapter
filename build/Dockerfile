# ---- Builder stage ----
FROM golang:1.25-alpine AS builder

# Set working directory
WORKDIR /app

# Copy Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Go application
RUN go build -o s3-adapter main.go

# ---- Runtime stage ----
FROM alpine:3.18

# Install certificates for TLS
RUN apk add --no-cache ca-certificates

# Set working directory
WORKDIR /app

# Copy built binary
COPY --from=builder /app/s3-adapter .

# Copy certs and config files
COPY certs ./certs
COPY config/config.yaml ./config/config.yaml

# Expose HTTPS port
EXPOSE 8080

# Run the application
CMD ["./s3-adapter"]
