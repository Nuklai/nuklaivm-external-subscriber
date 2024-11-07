#!/bin/bash

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
echo "Starting PostgreSQL with TimescaleDB..."
docker-compose up -d

# Wait for the database to be ready
echo "Waiting for PostgreSQL to start..."
sleep 10

# Create database schema
echo "Creating database schema..."
# Create schema if needed
docker exec -i timescaledb psql -U postgres -d blockchain < scripts/schema.sql

# Build and start the Go application
echo "Building the Go application..."
go build -o bin/nuklaivm-subscriber main.go

echo "Starting the application..."
nohup ./bin/nuklaivm-subscriber &> block_subscriber.log &

echo "Setup complete. Application is running."
