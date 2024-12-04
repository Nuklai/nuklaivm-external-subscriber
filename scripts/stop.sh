#!/bin/bash


# Kill existing process on port 50051 and 8080
# fuser -k 50051/tcp 8080/tcp

# Stop the Go application gracefully
pkill -f nuklaivm-subscriber

# Stop Docker containers without nuking volumes
docker-compose down

docker container rm -f nuklaivm-postgres >/dev/null 2>&1
docker container rm -f nuklaivm-subscriber >/dev/null 2>&1
