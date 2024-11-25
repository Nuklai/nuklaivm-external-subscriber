# Build stage
FROM golang:1.22.5 AS builder

# Install required dependencies
RUN apt-get update && apt-get install -y \
    gcc libc-dev bash make git

# Set the working directory
WORKDIR /go/src/app

# Copy the Go application source code
COPY . .

# Build the application
RUN go mod tidy && go build -o build/subscriber main.go

# Final stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y bash libc6 libgcc1 libstdc++6 && apt-get clean

# Create a new user and group
RUN groupadd -r nuklai && useradd --no-log-init -r -g nuklai nuklai

# Copy the build artifacts from the builder stage
COPY --from=builder --chown=nuklai:nuklai /go/src/app/build /app

# Set permissions and ownership
USER nuklai

# Set the working directory and entry point
WORKDIR /app
CMD [ "/app/subscriber" ]

# Expose necessary ports
EXPOSE 8080
EXPOSE 50051

# Metadata
LABEL Name=subscriber
