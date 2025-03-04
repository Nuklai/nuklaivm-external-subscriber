.PHONY: build run clean test lint fmt init-config deps docker-build docker-run migrate-up migrate-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BINARY_NAME=indexer
MAIN_PATH=./cmd/indexer

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

run:
	$(GORUN) $(MAIN_PATH)

clean:
	rm -f $(BINARY_NAME)
	rm -rf ./bin

test:
	$(GOTEST) -v ./...

test-cover:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

lint:
	golangci-lint run

fmt:
	$(GOFMT) ./...

init-config:
	./$(BINARY_NAME) init

deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Docker
docker-build:
	docker build -t nuklai-indexer .

docker-run:
	docker run --name nuklai-indexer -p 8080:8080 nuklai-indexer

# Database migrations
migrate-up:
	migrate -path ./db/migrations -database "postgresql://postgres:postgres@localhost:5432/nuklai_indexer?sslmode=disable" up

migrate-down:
	migrate -path ./db/migrations -database "postgresql://postgres:postgres@localhost:5432/nuklai_indexer?sslmode=disable" down

# Create a new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./db/migrations -seq $$name
	