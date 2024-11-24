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

Set up environment variables, particularly `DATABASE_URL` for PostgreSQL access. By default, this is configured in `config/config.go` to:

```sh
DATABASE_URL="postgres://postgres:postgres@localhost:5432/blockchain?sslmode=disable"
```

### Step 3: Run the Program

Copy the .env.example file:

```sh
cp .env.example .env
```

Note that if you modify the values of `DB_USER`, `DB_PASSWORD` or `DB_NAME`, make sure to also update your docker-compose.yml file accordingly under the `environment` and `entrypoint` section.

Run the setup script to initialize the database and start the servers:

```sh
./scripts/run.sh
```

This script:

- Starts a PostgreSQL container with TimescaleDB using Docker Compose.
- Sets up database tables and indexes.
- Builds and runs the Go application, starting both the gRPC and REST API servers.

For a fresh start (removing old data), use:

```sh
./scripts/run.sh -n
```

### Step 4: Access the REST API

The REST API is available at `http://localhost:8080`.

## Usage

### REST API Endpoints

#### Block Endpoints

- **Get Block by Height or Hash**

  - **Endpoint**: `/blocks/:identifier`
  - **Description**: Retrieve a block by its height or hash.
  - **Example**: `curl http://localhost:8080/blocks/29` or `curl http://localhost:8080/blocks/block_hash_here`

- **Get All Blocks**
  - **Endpoint**: `/blocks`
  - **Parameters**:
    - `limit`: Number of blocks to return (default: 10).
    - `offset`: Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/blocks?limit=5&offset=0"`

#### Transaction Endpoints

- **Get Transaction by Hash**

  - **Endpoint**: `/transactions/:tx_hash`
  - **Example**: `curl http://localhost:8080/transactions/tx_hash_here`

- **Get All Transactions**

  - **Endpoint**: `/transactions`
  - **Parameters**:
    - `limit`: Number of transactions to return (default: 10).
    - `offset`: Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/transactions?limit=5&offset=0"`

- **Get Transactions by Block**
  - **Endpoint**: `/transactions/block/:identifier`
  - **Description**: Retrieve transactions associated with a block, specified by either block height or hash.
  - **Example**: `curl http://localhost:8080/transactions/block/29`

#### Action Endpoints

- **Get All Actions**

  - **Endpoint**: `/actions`
  - **Parameters**:
    - `limit`: Number of actions to return (default: 10).
    - `offset`: Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/actions?limit=5&offset=0"`

- **Get Actions by Transaction Hash**

  - **Endpoint**: `/actions/:tx_hash`
  - **Description**: Retrieve actions associated with a specific transaction hash.
  - **Example**: `curl http://localhost:8080/actions/tx_hash_here`

- **Get Actions by Block**
  - **Endpoint**: `/actions/block/:identifier`
  - **Description**: Retrieve actions within a block, specified by either block height or hash.
  - **Example**: `curl http://localhost:8080/actions/block/29`

#### Genesis Data Endpoint

- **Get Genesis Data**
  - **Endpoint**: `/genesis`
  - **Description**: Retrieve the genesis data.
  - **Example**: `curl http://localhost:8080/genesis`

### gRPC Server

The gRPC server listens on port `50051` and implements methods defined in the `ExternalSubscriber` service:

- **Initialize**: Receives the genesis data and saves it to the database.
- **AcceptBlock**: Receives block information and saves block, transaction, and action data.

## Database Schema

The database schema includes the following tables:

- **`blocks`**: Stores block information (height, hash, parent hash, state root, timestamp, and unit prices).
- **`transactions`**: Stores transaction details (hash, block association, sponsor, fee, success status, outputs).
- **`actions`**: Stores actions within transactions, including action type and details.
- **`genesis_data`**: Stores the genesis data received during initialization.

### Indexes

Indexes are created on critical fields to optimize query performance:

- `block_height`, `block_hash`, and `timestamp` in the `blocks` table.
- `tx_hash`, `block_hash`, and `sponsor` in the `transactions` table.
- `tx_hash` and `action_type` in the `actions` table.

## Running Tests

Unit tests are located in each package, and the command below will run all tests:

```sh
go test ./...
```

Mocking is used to simulate database connections for accurate testing without requiring a live database.

## Logging

The subscriber logs important events and errors to `block_subscriber.log`. For structured logging, it’s recommended to use packages like `logrus` or `zap` for better observability.

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository.
2. Create a new feature branch (`git checkout -b feature-name`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-name`).
5. Create a Pull Request.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
