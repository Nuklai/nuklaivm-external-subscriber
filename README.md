# NuklaiVM External Subscriber

This project is an external subscriber for the NuklaiVM blockchain that captures and stores historical blockchain data. It runs a gRPC server to receive block and transaction data and a REST API server to expose that data for querying. The subscriber stores data in a PostgreSQL database with TimescaleDB enabled for efficient handling of time-series data.

## Features

- gRPC server to receive block and transaction data from the blockchain.
- REST API to expose the stored blockchain data for easy retrieval.
- Supports real-time data storage, including blocks, transactions, and actions.
- Full historical data storage beyond the default 1024-block limit of the in-memory indexer.
- Written in Go for efficient data processing.

## Requirements

- [Go 1.22.5+](https://golang.org/dl/)
- [PostgreSQL with TimescaleDB](https://www.timescale.com/)
- Docker (for setting up PostgreSQL using Docker Compose)
- gRPC tools

## Running Instructions

### Step 1: Clone the Repository

```sh
git clone https://github.com/Nuklai/nuklaivm-external-subscriber
cd nuklaivm-external-subscriber
```

### Step 2: Run the Program

```sh
./scripts/run.sh
```

This will do the following:

- Use Docker Compose to set up a PostgreSQL instance with TimescaleDB.
- Start a PostgreSQL container with the necessary tables to store blocks, transactions, actions, and genesis data.

If you would like to delete the postgres data volume, you can do:

```sh
./scripts/run.sh -n
```

### Step 3: Access the API

The REST API will be available at `http://localhost:8080`.

## API Documentation

### REST API Endpoints

The following REST API endpoints are available:

- **Get Block by Height or Hash**

  - **Endpoint**: `/blocks/:identifier`
  - **Description**: Retrieve a block by its height or hash. The identifier can be either the block height or block hash.

- **Get All Blocks**

  - **Endpoint**: `/blocks`
  - **Description**: Retrieve all blocks, with optional pagination support.
  - **Parameters**:
    - `limit`: Number of blocks to return (default: 10).
    - `offset`: Offset for pagination (default: 0).

- **Get Transaction by Hash**

  - **Endpoint**: `/transactions/:tx_hash`
  - **Description**: Retrieve a transaction by its hash.

- **Get All Transactions**

  - **Endpoint**: `/transactions`
  - **Description**: Retrieve all transactions, with optional pagination support.
  - **Parameters**:
    - `limit`: Number of transactions to return (default: 10).
    - `offset`: Offset for pagination (default: 0).

- **Get Transactions by Block Height or Hash**

  - **Endpoint**: `/transactions/block/:identifier`
  - **Description**: Retrieve all transactions associated with a block, specified by either block height or hash.

- **Get Actions by Transaction Hash**

  - **Endpoint**: `/actions/:tx_hash`
  - **Description**: Retrieve actions associated with a specific transaction hash.

- **Get All Actions**

  - **Endpoint**: `/actions`
  - **Description**: Retrieve all actions, with optional pagination support.
  - **Parameters**:
    - `limit`: Number of actions to return (default: 10).
    - `offset`: Offset for pagination (default: 0).

- **Get Actions by Block Height or Hash**

  - **Endpoint**: `/actions/block/:identifier`
  - **Description**: Retrieve all actions within a block, specified by either block height or hash.

- **Get Genesis Data**
  - **Endpoint**: `/genesis`
  - **Description**: Retrieve the genesis data.

### Example Queries

Here are some example queries that can be made to the REST API endpoints:

1. **Retrieve a Block by Height or Hash**

   ```sh
   curl http://localhost:8080/blocks/10
   ```

   or

   ```sh
   curl http://localhost:8080/blocks/abcd1234
   ```

2. **Retrieve All Blocks**

   ```sh
   curl "http://localhost:8080/blocks?limit=5&offset=0"
   ```

3. **Retrieve a Transaction by Hash**

   ```sh
   curl http://localhost:8080/transactions/abcd1234
   ```

4. **Retrieve All Transactions**

   ```sh
   curl "http://localhost:8080/transactions?limit=5&offset=0"
   ```

5. **Retrieve Transactions by Block Height or Hash**

   ```sh
   curl http://localhost:8080/transactions/block/10
   ```

   or

   ```sh
   curl http://localhost:8080/transactions/block/abcd1234
   ```

6. **Retrieve Actions by Transaction Hash**

   ```sh
   curl http://localhost:8080/actions/txid1234
   ```

7. **Retrieve All Actions**

   ```sh
   curl "http://localhost:8080/actions?limit=5&offset=0"
   ```

8. **Retrieve Actions by Block Height or Hash**

   ```sh
   curl http://localhost:8080/actions/block/100
   ```

   or

   ```sh
   curl http://localhost:8080/actions/block/abcd1234
   ```

9. **Retrieve Genesis Data**

   ```sh
   curl http://localhost:8080/genesis
   ```

## gRPC Server

The gRPC server listens on port `50051` and supports the following methods:

- **Initialize**: Receives the genesis data and saves it to the database.
- **AcceptBlock**: Receives information about a new block and saves it along with transaction data.

### Proto Definition

The gRPC server is implemented using the `ExternalSubscriber` service defined in the `hypersdk` proto.

## Code Structure

- **`main.go`**: Contains the main logic for the gRPC server, REST API, and database interactions.
- **`scripts/run.sh`**: Script for setting up the PostgreSQL database, including schema creation and running the program.
- **`docker-compose.yml`**: Docker Compose file to set up a PostgreSQL instance with TimescaleDB.

## Database Schema

The database schema consists of the following tables:

- **`blocks`**: Stores block data, including height, hash, parent block, state root, timestamp, and unit prices.
- **`transactions`**: Stores transaction data, including hash, associated block hash, sponsor, fee, success status, and outputs.
- **`actions`**: Stores actions within transactions, including action type and details.
- **`genesis_data`**: Stores the genesis data received during initialization.

### Sample Schema Setup

The schema setup is included in `scripts/run.sh` and executed using Docker.

## Running Tests

To run tests, make sure to have the testing environment configured properly. You can use the following command to run unit tests:

```sh
go test ./...
```

## Contributions

Feel free to submit pull requests or issues on the GitHub repository. All contributions are welcome to improve the project.
