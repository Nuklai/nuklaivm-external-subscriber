#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
  export $(cat .env | xargs)
fi

# Parse command-line arguments
DOCKER_MODE=false

if [ "$1" == "-d" ]; then
  DOCKER_MODE=true
elif [ "$1" != "" ]; then
  echo "Usage: $0 [-d (run everything in Docker mode)]"
  exit 1
fi

# Stop and clean up previous containers and resources
echo "Stopping and cleaning up previous containers..."
docker-compose down -v --remove-orphans
docker container rm -f nuklaivm-postgres >/dev/null 2>&1
docker container rm -f nuklaivm-subscriber >/dev/null 2>&1

ensure_network() {
  if ! docker network inspect nuklaivm-network >/dev/null 2>&1; then
    echo "Creating network nuklaivm-network..."
    docker network create nuklaivm-network
  fi
}

ensure_network

if [ "$DOCKER_MODE" = true ]; then
  # Run everything in Docker
  echo "Running everything in Docker mode..."
  docker-compose up --build -d
  echo "All services are running in Docker."
  docker container logs -f nuklaivm-subscriber
else
  # Run PostgreSQL in Docker and subscriber locally
  echo "Starting PostgreSQL in Docker..."
  docker-compose up -d nuklaivm-postgres

  echo "Waiting for PostgreSQL to become ready..."
  # Wait until PostgreSQL is ready to accept connections
  until docker exec -i nuklaivm-postgres pg_isready -U postgres > /dev/null 2>&1; do
    sleep 2
  done

  echo "Building the subscriber program locally..."
  go build -o bin/nuklaivm-subscriber main.go

  echo "Starting the subscriber program locally..."
  ./bin/nuklaivm-subscriber
  echo "Subscriber program is running locally."
fi
