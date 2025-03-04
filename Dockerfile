# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required dependencies
RUN apk add --no-cache git make

# Copy the Go application source code
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o indexer ./cmd/indexer

# Use a minimal alpine image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/indexer .

# Copy config directory
COPY --from=builder /app/config ./config

# Create a new user and group
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of the application files
RUN chown -R appuser:appgroup /app

# Set permissions and ownership
USER nuklai

# Expose necessary ports
EXPOSE 8080

# Run app
ENTRYPOINT ["./indexer"]
CMD ["--config", "./config/config.yaml"]
