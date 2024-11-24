#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
  export $(cat .env | xargs)
fi

# Parse command line arguments for nuking data
NUKE=false
while getopts "n" opt; do
  case ${opt} in
    n )
      NUKE=true
      ;;
    * )
      echo "Usage: $0 [-n (nuke database)]"
      exit 1
      ;;
  esac
done

# Stop the previous instance of the application
./scripts/stop.sh

rm -f block_subscriber.log

if [ "$NUKE" = true ]; then
  # Stop and remove old Docker container if running
  echo "Stopping and removing old PostgreSQL container if exists..."
  docker-compose down -v --remove-orphans
else
  # Stop and remove old Docker container if running, keeping volumes intact
  echo "Stopping old PostgreSQL container if exists..."
  docker-compose down
fi

# Start Docker container
echo "Starting PostgreSQL..."
docker-compose --env-file .env up -d

echo "Waiting for PostgreSQL to become ready..."
# Wait until PostgreSQL is ready to accept connections
until docker exec -i postgres pg_isready -U ${DB_USER:-postgres} > /dev/null 2>&1; do
  sleep 2
done

# Build and start the Go application
echo "Building the Go application..."
go build -o bin/nuklaivm-subscriber main.go

echo "Starting the application..."
nohup ./bin/nuklaivm-subscriber &> block_subscriber.log &

echo "Setup complete. Application is running."
