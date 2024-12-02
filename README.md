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

Run the setup script to initialize the database and start the servers.

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

### Health Endpoint

- **Check health status**

  - **Endpoint**: `/health`
  - **Description**: Check the health status of the subscriber
  - **Example**: `curl http://localhost:8080/health`
  - **Output**:

    ```bash
    {
      "details": {
        "database": "reachable",
        "grpc": "reachable"
      },
      "status": "ok"
    }
    ```

#### Genesis Data Endpoint

- **Get Genesis Data**

  - **Endpoint**: `/genesis`
  - **Description**: Retrieve the genesis data.
  - **Example**: `curl http://localhost:8080/genesis`
  - **Output**:

    ```bash
    {
      "customAllocation": [
        {
          "address": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "balance": 853000000000000000
        }
      ],
      "emissionBalancer": {
        "emissionAddress": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
        "maxSupply": 1e+19
      },
      "initialRules": {
        "baseUnits": 1,
        "chainID": "11111111111111111111111111111111LpoYY",
        "maxActionsPerTx": 16,
        "maxBlockUnits": {
          "bandwidth": 1800000,
          "compute": 18446744073709552000,
          "storageAllocate": 18446744073709552000,
          "storageRead": 18446744073709552000,
          "storageWrite": 18446744073709552000
        },
        "maxOutputsPerAction": 1,
        "minBlockGap": 250,
        "minEmptyBlockGap": 750,
        "minUnitPrice": {
          "bandwidth": 100,
          "compute": 100,
          "storageAllocate": 100,
          "storageRead": 100,
          "storageWrite": 100
        },
        "networkID": 0,
        "sponsorStateKeysMaxChunks": [
          1
        ],
        "storageKeyAllocateUnits": 20,
        "storageKeyReadUnits": 5,
        "storageKeyWriteUnits": 10,
        "storageValueAllocateUnits": 5,
        "storageValueReadUnits": 2,
        "storageValueWriteUnits": 3,
        "unitPriceChangeDenominator": {
          "bandwidth": 48,
          "compute": 48,
          "storageAllocate": 48,
          "storageRead": 48,
          "storageWrite": 48
        },
        "validityWindow": 60000,
        "windowTargetUnits": {
          "bandwidth": 18446744073709552000,
          "compute": 18446744073709552000,
          "storageAllocate": 18446744073709552000,
          "storageRead": 18446744073709552000,
          "storageWrite": 18446744073709552000
        }
      },
      "stateBranchFactor": 16
    }
    ```

#### Block Endpoints

- **Get Block by Height or Hash**

  - **Endpoint**: `/blocks/:identifier`
  - **Description**: Retrieve a block by its height or hash.
  - **Example**: `curl http://localhost:8080/blocks/29` or `curl http://localhost:8080/blocks/block_hash_here`
  - **Output**:

    ```bash
      {
        "ID": 832,
        "BlockHeight": 416,
        "BlockHash": "2b38U36vP4esbZbjAXu54pPALaJ32kaf32tDU1qC6Ek97aRtH1",
        "ParentBlock": "2AHQ7qbehvuYJ2sDB8Wgwz5PMyJx2BejHhJZa292KmbXw3Te9R",
        "StateRoot": "6jGeT1cD6SopVAcc13yyJ8VMLFdiPh6upL3BAoBv1HGPUUuUZ",
        "Timestamp": "2024-12-02T10:22:50Z",
        "UnitPrices": "(Bandwidth=100, Compute=100, Storage(Read)=100, Storage(Allocate)=100, Storage(Write)=100)"
      }
    ```

- **Get All Blocks**

  - **Endpoint**: `/blocks`
  - **Parameters**:
    - `limit`: Number of blocks to return (default: 10).
    - `offset`: Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/blocks?limit=2&offset=0"`
  - **Output**:

    ```bash
    [
      {
        "ID": 1988,
        "BlockHeight": 994,
        "BlockHash": "iRex7gTFqNQKfhMn3Voqex2vLaJXfqQ5xg3xhxqpuc9Rn6LUy",
        "ParentBlock": "NSsynENZfcbXges2JuFqSgj6SDGMTq7jCgFqXnjqj9i74YAwa",
        "StateRoot": "taSygsQcSb8y1EMwwEXkUvM9Dz5SFDkf3Rz6YFtcwAxQbZmUk",
        "Timestamp": "2024-12-02T10:30:11Z",
        "UnitPrices": "(Bandwidth=100, Compute=100, Storage(Read)=100, Storage(Allocate)=100, Storage(Write)=100)"
      },
      {
        "ID": 1986,
        "BlockHeight": 993,
        "BlockHash": "NSsynENZfcbXges2JuFqSgj6SDGMTq7jCgFqXnjqj9i74YAwa",
        "ParentBlock": "ZNwNGUx9DL4hoSx76wwS2RAUnMj773bGg51bd42oaqgWdxiMg",
        "StateRoot": "2RizGvY7zGDgNkiXWVvuEhVkGAtydGoB3irSzgFq1qUV41ysLP",
        "Timestamp": "2024-12-02T10:30:10Z",
        "UnitPrices": "(Bandwidth=100, Compute=100, Storage(Read)=100, Storage(Allocate)=100, Storage(Write)=100)"
      }
    ]
    ```

#### Transaction Endpoints

- **Get Transaction by Hash**

  - **Endpoint**: `/transactions/:tx_hash`
  - **Example**: `curl http://localhost:8080/transactions/tx_hash_here`
  - **Output**:

    ```bash
    {
      "ID": 2,
      "TxHash": "WPfzKZZAeug9wakxdzpQyp2qzJ27Kbi8BEfEY3KQDAKfiBbDp",
      "BlockHash": "2b38U36vP4esbZbjAXu54pPALaJ32kaf32tDU1qC6Ek97aRtH1",
      "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "MaxFee": 53000,
      "Success": true,
      "Fee": 48500,
      "Outputs": {
        "receiver_balance": 100000000000,
        "sender_balance": 852999899999951500
      },
      "Timestamp": "2024-12-02T10:22:50Z"
    }
    ```

- **Get All Transactions**

  - **Endpoint**: `/transactions`
  - **Parameters**:
    - `limit`: Number of transactions to return (default: 10).
    - `offset`: Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/transactions?limit=2&offset=0"`
  - **Output**:

    ```bash
    [
      {
        "ID": 4,
        "TxHash": "57KRbZ34ypofg839mYuiwRAsN2XdzsoLW7UXoTpdYGHLrMJNX",
        "BlockHash": "CZP3tAWjARyoyNV18qD9QPfiarNZUbSxqPSEP1rF7ESoBJRWU",
        "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
        "MaxFee": 53000,
        "Success": true,
        "Fee": 48500,
        "Outputs": {
          "receiver_balance": 105000000000,
          "sender_balance": 852999894999903000
        },
        "Timestamp": "2024-12-02T10:32:22Z"
      },
      {
        "ID": 2,
        "TxHash": "WPfzKZZAeug9wakxdzpQyp2qzJ27Kbi8BEfEY3KQDAKfiBbDp",
        "BlockHash": "2b38U36vP4esbZbjAXu54pPALaJ32kaf32tDU1qC6Ek97aRtH1",
        "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
        "MaxFee": 53000,
        "Success": true,
        "Fee": 48500,
        "Outputs": {
          "receiver_balance": 100000000000,
          "sender_balance": 852999899999951500
        },
        "Timestamp": "2024-12-02T10:22:50Z"
      }
    ]
    ```

- **Get Transactions by Block**

  - **Endpoint**: `/transactions/block/:identifier`
  - **Description**: Retrieve transactions associated with a block, specified by either block height or hash.
  - **Example**: `curl http://localhost:8080/transactions/block/29`
  - **Output**:

    ```bash
    [
      {
        "ID": 2,
        "TxHash": "WPfzKZZAeug9wakxdzpQyp2qzJ27Kbi8BEfEY3KQDAKfiBbDp",
        "BlockHash": "2b38U36vP4esbZbjAXu54pPALaJ32kaf32tDU1qC6Ek97aRtH1",
        "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
        "MaxFee": 53000,
        "Success": true,
        "Fee": 48500,
        "Outputs": {
          "receiver_balance": 100000000000,
          "sender_balance": 852999899999951500
        },
        "Timestamp": "2024-12-02T10:22:50Z"
      }
    ]
    ```

#### Action Endpoints

- **Get Actions by Transaction Hash**

  - **Endpoint**: `/actions/:tx_hash`
  - **Description**: Retrieve actions associated with a specific transaction hash.
  - **Example**: `curl http://localhost:8080/actions/tx_hash_here`
  - **Output**:

    ```bash
    [
      {
        "ID": 2,
        "TxHash": "WPfzKZZAeug9wakxdzpQyp2qzJ27Kbi8BEfEY3KQDAKfiBbDp",
        "ActionType": 0,
        "ActionDetails": {
          "AssetAddress": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
          "Memo": "",
          "To": "0x00f570339dce77fb2694edac17c9e6f36c0945959813b99b0b1a18849a7d622237",
          "Value": 100000000000
        },
        "Timestamp": "2024-12-02T10:22:50Z"
      }
    ]
    ```

- **Get All Actions**

  - **Endpoint**: `/actions`
  - **Parameters**:
    - `limit`: Number of actions to return (default: 10).
    - `offset`: Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/actions?limit=2&offset=0"`
  - **Output**:

    ```bash
    [
      {
        "ID": 4,
        "TxHash": "57KRbZ34ypofg839mYuiwRAsN2XdzsoLW7UXoTpdYGHLrMJNX",
        "ActionType": 0,
        "ActionDetails": {
          "AssetAddress": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
          "Memo": "",
          "To": "0x00f570339dce77fb2694edac17c9e6f36c0945959813b99b0b1a18849a7d622237",
          "Value": 5000000000
        },
        "Timestamp": "2024-12-02T10:32:22Z"
      },
      {
        "ID": 2,
        "TxHash": "WPfzKZZAeug9wakxdzpQyp2qzJ27Kbi8BEfEY3KQDAKfiBbDp",
        "ActionType": 0,
        "ActionDetails": {
          "AssetAddress": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
          "Memo": "",
          "To": "0x00f570339dce77fb2694edac17c9e6f36c0945959813b99b0b1a18849a7d622237",
          "Value": 100000000000
        },
        "Timestamp": "2024-12-02T10:22:50Z"
      }
    ]
    ```

- **Get Actions by Block**

  - **Endpoint**: `/actions/block/:identifier`
  - **Description**: Retrieve actions within a block, specified by either block height or hash.
  - **Example**: `curl http://localhost:8080/actions/block/29`
  - **Output**:

    ```bash
    [
      {
        "ID": 2,
        "TxHash": "WPfzKZZAeug9wakxdzpQyp2qzJ27Kbi8BEfEY3KQDAKfiBbDp",
        "ActionType": 0,
        "ActionDetails": {
          "AssetAddress": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
          "Memo": "",
          "To": "0x00f570339dce77fb2694edac17c9e6f36c0945959813b99b0b1a18849a7d622237",
          "Value": 100000000000
        },
        "Timestamp": "2024-12-02T10:22:50Z"
      }
    ]
    ```

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
