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

Note that if you modify the values of `DB_USER`, `DB_PASSWORD` or `DB_NAME`, make sure to also update your docker-compose.yml file accordingly under the `environment` and `entrypoint` section(if you plan on running the subscriber in docker).

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
      "BlockHeight": 73,
      "BlockHash": "p5jkYp2TfK1nH7WzXmeDypNKmtnMoRpNLadYCDNYoRyrTvxZS",
      "ParentBlockHash": "6bMS7Gi5u4SRfHUrboQcjhvpEjMNaNpFU4s7EAzdqi5Vqcyrj",
      "StateRoot": "oz2McvCZbS8gfp9sFjZfAFKawZW8TdrGjwwETfPpcEro26mTh",
      "BlockSize": 307,
      "TxCount": 1,
      "TotalFee": 48500,
      "AvgTxSize": 307,
      "UniqueParticipants": 1,
      "Timestamp": "2024-12-02T11:47:09Z"
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
    {
      "counter": 73,
      "items": [
        {
          "BlockHeight": 73,
          "BlockHash": "p5jkYp2TfK1nH7WzXmeDypNKmtnMoRpNLadYCDNYoRyrTvxZS",
          "ParentBlockHash": "6bMS7Gi5u4SRfHUrboQcjhvpEjMNaNpFU4s7EAzdqi5Vqcyrj",
          "StateRoot": "oz2McvCZbS8gfp9sFjZfAFKawZW8TdrGjwwETfPpcEro26mTh",
          "BlockSize": 307,
          "TxCount": 1,
          "TotalFee": 48500,
          "AvgTxSize": 307,
          "UniqueParticipants": 1,
          "Timestamp": "2024-12-02T11:47:09Z"
        },
        {
          "BlockHeight": 72,
          "BlockHash": "6bMS7Gi5u4SRfHUrboQcjhvpEjMNaNpFU4s7EAzdqi5Vqcyrj",
          "ParentBlockHash": "fStF9MQQwV7UKQK73Pp4reKBttRGuC1X4sQRoaiB1VTTT1eG3",
          "StateRoot": "2RuPbB1GejDpXG9pMxmmfUuP3XfLKrjmyNUXKbq8LSm6dALtL7",
          "BlockSize": 84,
          "TxCount": 0,
          "TotalFee": 0,
          "AvgTxSize": 0,
          "UniqueParticipants": 0,
          "Timestamp": "2024-12-02T11:48:12Z"
        }
      ]
    }
    ```

#### Transaction Endpoints

- **Get Transaction by Hash**

  - **Endpoint**: `/transactions/:tx_hash`
  - **Example**: `curl http://localhost:8080/transactions/tx_hash_here`
  - **Output**:

    ```bash
    {
      "ID": 1,
      "TxHash": "fPogR2TfRxHmDL2MXyZNJEo8NsFQvDk8Knzod4WzdHGudRZep",
      "BlockHash": "2shfGuUK1jcU4a8eB9EoCJthcG6dVHhtNJccnY5aMAqCzgGrbG",
      "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "MaxFee": 53000,
      "Success": true,
      "Fee": 48500,
      "Actions": [
        {
          "ActionType": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 5000000000
          },
          "Output": {
            "receiver_balance": 5000000000,
            "sender_balance": 852999994999951500
          }
        }
      ],
      "Timestamp": "2024-12-02T21:48:37Z"
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
    {
      "counter": 2,
      "items": [
        {
          "ID": 3,
          "TxHash": "21Es58ApKEK91Sq4jXsL35k5SME97kswSgwh9zeykATMneVMGi",
          "BlockHash": "2NRjSUBBw4KTQsJr939ArJLx5PihiuqQSaHm4c58T4DdMJD6Fs",
          "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "MaxFee": 53000,
          "Success": true,
          "Fee": 48500,
          "Actions": [
            {
              "ActionType": 0,
              "Input": {
                "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
                "memo": "",
                "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
                "value": 10000000000
              },
              "Output": {
                "receiver_balance": 15000000000,
                "sender_balance": 852999984999903000
              }
            }
          ],
          "Timestamp": "2024-12-02T21:50:15Z"
        },
        {
          "ID": 1,
          "TxHash": "fPogR2TfRxHmDL2MXyZNJEo8NsFQvDk8Knzod4WzdHGudRZep",
          "BlockHash": "2shfGuUK1jcU4a8eB9EoCJthcG6dVHhtNJccnY5aMAqCzgGrbG",
          "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "MaxFee": 53000,
          "Success": true,
          "Fee": 48500,
          "Actions": [
            {
              "ActionType": 0,
              "Input": {
                "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
                "memo": "",
                "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
                "value": 5000000000
              },
              "Output": {
                "receiver_balance": 5000000000,
                "sender_balance": 852999994999951500
              }
            }
          ],
          "Timestamp": "2024-12-02T21:48:37Z"
        }
      ]
    }
    ```

- **Get Transactions by Block**

  - **Endpoint**: `/transactions/block/:identifier`
  - **Description**: Retrieve transactions associated with a block, specified by either block height or hash.
  - **Example**: `curl http://localhost:8080/transactions/block/29`
  - **Output**:

    ```bash
    [
      {
        "ID": 1,
        "TxHash": "fPogR2TfRxHmDL2MXyZNJEo8NsFQvDk8Knzod4WzdHGudRZep",
        "BlockHash": "2shfGuUK1jcU4a8eB9EoCJthcG6dVHhtNJccnY5aMAqCzgGrbG",
        "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
        "MaxFee": 53000,
        "Success": true,
        "Fee": 48500,
        "Actions": [
          {
            "ActionType": 0,
            "Input": {
              "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
              "memo": "",
              "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
              "value": 5000000000
            },
            "Output": {
              "receiver_balance": 5000000000,
              "sender_balance": 852999994999951500
            }
          }
        ],
        "Timestamp": "2024-12-02T21:48:37Z"
      }
    ]
    ```

- **Get Transactions For a specific user**

  - **Endpoint**: `/transactions/user/:user`
  - **Description**: Retrieves all transactions associated with the specified user (sponsor). Supports both plain and 0x-prefixed identifiers.
  - **Parameters**:
    - Limit (optional): Number of transactions to return (default: 10).
    - offset (optional): Offset for pagination (default: 0).
  - **Example**: `curl http://localhost:8080/transactions/user/006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840`
  - **Output**:

    ```bash
    {
      "counter": 1,
      "items": [
        {
          "ID": 1,
          "TxHash": "UrQ1r61PUjQuq9DzGZLjomSeVYt1SGhiNTrdYRA98RrXNciSy",
          "BlockHash": "y7anf2mLCPeKhBtJ1PqXeTUhRaF3DnhWCiQkTwRPuxzrj8dHk",
          "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "MaxFee": 53000,
          "Success": true,
          "Fee": 48500,
          "Actions": [
            {
              "ActionType": 0,
              "Input": {
                "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
                "memo": "",
                "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
                "value": 5000000000
              },
              "Output": {
                "receiver_balance": 5000000000,
                "sender_balance": 852999994999951500
              }
            }
          ],
          "Timestamp": "2024-12-03T16:20:53Z"
        }
      ]
    }
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
        "ID": 1,
        "TxHash": "fPogR2TfRxHmDL2MXyZNJEo8NsFQvDk8Knzod4WzdHGudRZep",
        "ActionType": 0,
        "ActionIndex": 0,
        "Input": {
          "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
          "memo": "",
          "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
          "value": 5000000000
        },
        "Output": {
          "receiver_balance": 5000000000,
          "sender_balance": 852999994999951500
        },
        "Timestamp": "2024-12-02T21:48:37Z"
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
    {
      "counter": 2,
      "items": [
        {
          "ID": 3,
          "TxHash": "21Es58ApKEK91Sq4jXsL35k5SME97kswSgwh9zeykATMneVMGi",
          "ActionType": 0,
          "ActionIndex": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 10000000000
          },
          "Output": {
            "receiver_balance": 15000000000,
            "sender_balance": 852999984999903000
          },
          "Timestamp": "2024-12-02T21:50:15Z"
        },
        {
          "ID": 1,
          "TxHash": "fPogR2TfRxHmDL2MXyZNJEo8NsFQvDk8Knzod4WzdHGudRZep",
          "ActionType": 0,
          "ActionIndex": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 5000000000
          },
          "Output": {
            "receiver_balance": 5000000000,
            "sender_balance": 852999994999951500
          },
          "Timestamp": "2024-12-02T21:48:37Z"
        }
      ]
    }
    ```

- **Get Actions by Block**

  - **Endpoint**: `/actions/block/:identifier`
  - **Description**: Retrieve actions within a block, specified by either block height or hash.
  - **Example**: `curl http://localhost:8080/actions/block/29`
  - **Output**:

    ```bash
    [
      {
        "ID": 1,
        "TxHash": "fPogR2TfRxHmDL2MXyZNJEo8NsFQvDk8Knzod4WzdHGudRZep",
        "ActionType": 0,
        "ActionIndex": 0,
        "Input": {
          "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
          "memo": "",
          "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
          "value": 5000000000
        },
        "Output": {
          "receiver_balance": 5000000000,
          "sender_balance": 852999994999951500
        },
        "Timestamp": "2024-12-02T21:48:37Z"
      }
    ]
    ```

- **Get Actions For a specific user**

  - **Endpoint**: `/actions/user/:user`
  - **Description**: Retrieves all actions associated with the specified user. Supports both plain and 0x-prefixed identifiers.
  - **Parameters**:
    - Limit (optional): Number of actions to return (default: 10).
    - offset (optional): Offset for pagination (default: 0).
  - **Example**: `curl http://localhost:8080/actions/user/006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840`
  - **Output**:

    ```bash
    {
      "counter": 1,
      "items": [
        {
          "ID": 1,
          "TxHash": "UrQ1r61PUjQuq9DzGZLjomSeVYt1SGhiNTrdYRA98RrXNciSy",
          "ActionType": 0,
          "ActionIndex": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 5000000000
          },
          "Output": {
            "receiver_balance": 5000000000,
            "sender_balance": 852999994999951500
          },
          "Timestamp": "2024-12-03T16:20:53Z"
        }
      ]
    }
    ```

### gRPC Server

The gRPC server listens on port `50051` and implements methods defined in the `ExternalSubscriber` service:

- **Initialize**: Receives the genesis data and saves it to the database.
- **AcceptBlock**: Receives block information and saves block, transaction, and action data.

## Database Schema

The database schema includes the following tables:

- **`blocks`**: Stores block information
- **`transactions`**: Stores transaction details
- **`actions`**: Stores actions within transactions, including action type and details
- **`genesis_data`**: Stores the genesis data received during initialization

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
