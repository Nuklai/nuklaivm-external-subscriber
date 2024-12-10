# NuklaiVM External Subscriber

The `NuklaiVM External Subscriber` is an external data collector for the NuklaiVM blockchain. This application listens to blockchain events and saves block, transaction, and action data to a PostgreSQL database with TimescaleDB for efficient time-series data handling. It provides both a gRPC server to receive blockchain data in real time and a REST API for querying historical blockchain data.

## Key Features

- **gRPC Server**: Receives block and transaction data from the blockchain.
- **REST API**: Exposes blockchain data, including blocks, transactions, actions, and genesis data, with pagination and filtering capabilities.
- **Data Storage**: Stores data in PostgreSQL with TimescaleDB for efficient storage and querying of time-series data.
- **Historical Data Storage**: Provides full historical data beyond the default 1024-block in-memory limit of the NuklaiVM indexer.
- **Modular Configurations**: Includes customizable configurations for server addresses and database settings.

## Prerequisites

- **Go**: Version 1.22.5 or higher
- **PostgreSQL**: With TimescaleDB enabled
- **Docker**: For setting up PostgreSQL with TimescaleDB
- **gRPC tools**: For managing the gRPC server

## Installation and Setup

### Step 1: Clone the Repository

```sh
git clone https://github.com/Nuklai/nuklaivm-external-subscriber
cd nuklaivm-external-subscriber
```

### Step 2: Environment Setup

Set up environment variables:

```sh
cp .env.example .env
```

Modify any values within `.env` as per your own environment.

Note that if you modify the values of `DB_USER`, `DB_PASSWORD`, `DB_NAME` or `GRPC_WHITELISTED_BLOCKCHAIN_NODES`, make sure to also update your docker-compose.yml file accordingly under the `environment` and `entrypoint` section(if you plan on running the subscriber in docker).

By default, the subscriber rejects all connections. In order to add an IP address to a whitelist, you must put this in the `GRPC_WHITELISTED_BLOCKCHAIN_NODES` in your .env file and docker-compose.yml(if you plan on running with docker). If you're running the `nuklaivm` in docker, make sure to include the docker IP of that container in this environment variable as well.

Run the postgres in docker and subscriber natively:

```sh
./scripts/run.sh
```

Run the postgres in docker and subscriber in docker as well:

```sh
./scripts/run.sh -d
```

This script:

- Starts a PostgreSQL container using Docker Compose.
- Sets up database tables and indexes.
- Builds and runs the Go application, starting both the gRPC and REST API servers.

### Step 4: Access the REST API

The REST API is available at `http://localhost:8080`.

## Usage

### REST API Endpoints

- [Health APIs](./docs/rest_api/health.md)
- [Genesis APIs](./docs/rest_api/genesis.md)
- [Blocks APIs](./docs/rest_api/blocks.md)
- [Transactions APIs](./docs/rest_api/transactions.md)
- [Assets APIs](./docs/rest_api/assets.md)
- [Actions APIs](./docs/rest_api/actions.md)

### gRPC Server

The gRPC server listens on port `50051` and implements methods defined in the `ExternalSubscriber` service:

- **Initialize**: Receives the genesis data and saves it to the database.
- **AcceptBlock**: Receives block information and saves block, transaction, and action data.

## Database Schema

The database schema includes the following tables:

- **`blocks`**: Stores block information
- **`transactions`**: Stores transaction details
- **`assets`**: Stores assets details
- **`actions`**: Stores actions within transactions, including action type and details
- **`genesis_data`**: Stores the genesis data received during initialization

## Running Tests

TODO:

Unit tests are located in each package, and the command below will run all tests:

```sh
go test ./...
```

Mocking is used to simulate database connections for accurate testing without requiring a live database.

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository.
2. Create a new feature branch (`git checkout -b feature-name`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-name`).
5. Create a Pull Request.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
