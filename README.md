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
      "BlockHeight": 19,
      "BlockHash": "jr4LHX4wBYrLWow2skK6uoCNrk8SkQVqwehNd8JFH8cJzt9FH",
      "ParentBlockHash": "DUU3CkV5aLWgAJtmpL3NZTYgPUK5jFWjgDAXpypK5UhE7Z7vD",
      "StateRoot": "1Aep1DqL5gd7z7riiSBGAyBu46nFzuwRtvvhamudFyuSoWFo5",
      "BlockSize": 391,
      "TxCount": 1,
      "TotalFee": 74800,
      "AvgTxSize": 391,
      "UniqueParticipants": 1,
      "Timestamp": "2024-12-03T21:17:28Z"
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
      "counter": 224,
      "items": [
        {
          "BlockHeight": 224,
          "BlockHash": "LF4GsBcYkRLFCcZ9MBwCFTdq74co4PvBdBbVbHGpQZgEZagmk",
          "ParentBlockHash": "ExuHmcdhUhEZHBueNzfhCf46VqyVDtFmm1HLCVYVtGASkd7s1",
          "StateRoot": "Pwkw7jyhsFyQsHznYF13RbUTWqRX9EysjT434dyvjmQkiDBof",
          "BlockSize": 84,
          "TxCount": 0,
          "TotalFee": 0,
          "AvgTxSize": 0,
          "UniqueParticipants": 0,
          "Timestamp": "2024-12-03T21:20:04Z"
        },
        {
          "BlockHeight": 223,
          "BlockHash": "ExuHmcdhUhEZHBueNzfhCf46VqyVDtFmm1HLCVYVtGASkd7s1",
          "ParentBlockHash": "17A97XSc8Y7Yvaz1Ze2qNTN72DmtCC4fjQHHknnyf3UXQGHjA",
          "StateRoot": "2tR8iWDVcJDDHCRXwE9kSzj5GxxTFKrBj2As3UvmXK8AqmEvdE",
          "BlockSize": 84,
          "TxCount": 0,
          "TotalFee": 0,
          "AvgTxSize": 0,
          "UniqueParticipants": 0,
          "Timestamp": "2024-12-03T21:20:03Z"
        }
      ]
    }
    ```

#### Transaction Endpoints

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
          "TxHash": "3cMm96hvrC4Pg6ffKgspodRUCiqqDztYzVumaJtGeD3772hmP",
          "BlockHash": "2MENB3iJJRJPkvq212sCCrJgAWaWite5CijugKbuf7zEVvVqe7",
          "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "MaxFee": 53000,
          "Success": true,
          "Fee": 48500,
          "Actions": [
            {
              "ActionType": "Transfer",
              "ActionTypeID": 0,
              "Input": {
                "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
                "memo": "",
                "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
                "value": 5000000000
              },
              "Output": {
                "receiver_balance": 5000000000,
                "sender_balance": 852999994999876700
              }
            }
          ],
          "Timestamp": "2024-12-03T21:18:36Z"
        },
        {
          "ID": 1,
          "TxHash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
          "BlockHash": "jr4LHX4wBYrLWow2skK6uoCNrk8SkQVqwehNd8JFH8cJzt9FH",
          "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "MaxFee": 58300,
          "Success": true,
          "Fee": 74800,
          "Actions": [
            {
              "ActionType": "CreateAsset",
              "ActionTypeID": 4,
              "Input": {
                "asset_type": 0,
                "decimals": 0,
                "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
                "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
                "max_supply": 0,
                "metadata": "test",
                "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
                "name": "Kiran",
                "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
                "symbol": "KP1"
              },
              "Output": {
                "asset_balance": 0,
                "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
                "nft_id": ""
              }
            }
          ],
          "Timestamp": "2024-12-03T21:17:28Z"
        }
      ]
    }
    ```

- **Get Transaction by Hash**

  - **Endpoint**: `/transactions/:tx_hash`
  - **Example**: `curl http://localhost:8080/transactions/tx_hash_here`
  - **Output**:

    ```bash
    {
      "ID": 3,
      "TxHash": "3cMm96hvrC4Pg6ffKgspodRUCiqqDztYzVumaJtGeD3772hmP",
      "BlockHash": "2MENB3iJJRJPkvq212sCCrJgAWaWite5CijugKbuf7zEVvVqe7",
      "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "MaxFee": 53000,
      "Success": true,
      "Fee": 48500,
      "Actions": [
        {
          "ActionType": "Transfer",
          "ActionTypeID": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 5000000000
          },
          "Output": {
            "receiver_balance": 5000000000,
            "sender_balance": 852999994999876700
          }
        }
      ],
      "Timestamp": "2024-12-03T21:18:36Z"
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
        "TxHash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
        "BlockHash": "jr4LHX4wBYrLWow2skK6uoCNrk8SkQVqwehNd8JFH8cJzt9FH",
        "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
        "MaxFee": 58300,
        "Success": true,
        "Fee": 74800,
        "Actions": [
          {
            "ActionType": "CreateAsset",
            "ActionTypeID": 4,
            "Input": {
              "asset_type": 0,
              "decimals": 0,
              "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
              "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
              "max_supply": 0,
              "metadata": "test",
              "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
              "name": "Kiran",
              "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
              "symbol": "KP1"
            },
            "Output": {
              "asset_balance": 0,
              "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
              "nft_id": ""
            }
          }
        ],
        "Timestamp": "2024-12-03T21:17:28Z"
      }
    ]
    ```

- **Get Transactions For a specific user**

  - **Endpoint**: `/transactions/user/:user`
  - **Description**: Retrieves all transactions associated with the specified user (sponsor). Supports both plain and 0x-prefixed identifiers.
  - **Parameters**:
    - Limit (optional): Number of transactions to return (default: 10).
    - offset (optional): Offset for pagination (default: 0).
  - **Example**: `curl http://localhost:8080/transactions/user/00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9`
  - **Output**:

    ```bash
    {
      "counter": 2,
      "items": [
        {
          "ID": 3,
          "TxHash": "3cMm96hvrC4Pg6ffKgspodRUCiqqDztYzVumaJtGeD3772hmP",
          "BlockHash": "2MENB3iJJRJPkvq212sCCrJgAWaWite5CijugKbuf7zEVvVqe7",
          "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "MaxFee": 53000,
          "Success": true,
          "Fee": 48500,
          "Actions": [
            {
              "ActionType": "Transfer",
              "ActionTypeID": 0,
              "Input": {
                "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
                "memo": "",
                "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
                "value": 5000000000
              },
              "Output": {
                "receiver_balance": 5000000000,
                "sender_balance": 852999994999876700
              }
            }
          ],
          "Timestamp": "2024-12-03T21:18:36Z"
        },
        {
          "ID": 1,
          "TxHash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
          "BlockHash": "jr4LHX4wBYrLWow2skK6uoCNrk8SkQVqwehNd8JFH8cJzt9FH",
          "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "MaxFee": 58300,
          "Success": true,
          "Fee": 74800,
          "Actions": [
            {
              "ActionType": "CreateAsset",
              "ActionTypeID": 4,
              "Input": {
                "asset_type": 0,
                "decimals": 0,
                "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
                "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
                "max_supply": 0,
                "metadata": "test",
                "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
                "name": "Kiran",
                "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
                "symbol": "KP1"
              },
              "Output": {
                "asset_balance": 0,
                "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
                "nft_id": ""
              }
            }
          ],
          "Timestamp": "2024-12-03T21:17:28Z"
        }
      ]
    }
    ```

- **Get Aggregated Estimated Fees for different Transactions**

  - **Endpoint**: `/transactions/estimated_fee`
  - **Description**: Retrieve the average, minimum, and maximum fees for all transaction types and names over a specified time interval.
  - **Query Parameters**:
    - `interval` (optional): Time interval for which the estimated fee is calculated (e.g., 1m, 1h, 1d). Default is `1m`.
  - **Example**: `curl http://localhost:8080/transactions/estimated_fee?interval=1h`
  - **Output**:

    ```bash
    {
      "avg_fee": 61650,
      "min_fee": 48500,
      "max_fee": 74800,
      "tx_count": 2,
      "fees": [
        {
          "action_name": "Transfer",
          "action_type": 0,
          "avg_fee": 48500,
          "max_fee": 48500,
          "min_fee": 48500,
          "tx_count": 1
        },
        {
          "action_name": "CreateAsset",
          "action_type": 4,
          "avg_fee": 74800,
          "max_fee": 74800,
          "min_fee": 74800,
          "tx_count": 1
        }
      ]
    }
    ```

- **Get Estimated Fee for different Transactions by action type**

  - **Endpoint**: `/transactions/estimated_fee/action_type/:action_type`
  - **Description**: Retrieve the average, minimum, and maximum fees for transactions of a specific action type.
  - **Path Parameters**:
    - `action_type`: The ID of the action type (e.g., 0 for "Transfer").
  - **Query Parameters**:
    - `interval` (optional): Time interval for which the estimated fee is calculated (e.g., 1m, 1h, 1d). Default is `1m`.
  - **Example**: `curl http://localhost:8080/transactions/estimated_fee/action_type/0?interval=1h`
  - **Output**:

    ```bash
    {
      "action_name": "Transfer",
      "action_type": 0,
      "avg_fee": 48500,
      "max_fee": 48500,
      "min_fee": 48500,
      "tx_count": 1
    }
    ```

- **Get Estimated Fee for different Transactions by action name**

  - **Endpoint**: `/transactions/estimated_fee/action_name/:action_name`
  - **Description**: Retrieve the average, minimum, and maximum fees for transactions of a specific action name.
  - **Path Parameters**:
    - `action_name`: The name of the action (e.g., "Transfer", "CreateAsset", etc). Case-insensitive.
  - **Query Parameters**:
    - `interval` (optional): Time interval for which the estimated fee is calculated (e.g., 1m, 1h, 1d). Default is `1m`.
  - **Example**: `curl http://localhost:8080/transactions/estimated_fee/action_name/createasset?interval=1h`
  - **Output**:

    ```bash
    {
      "action_name": "CreateAsset",
      "action_type": 4,
      "avg_fee": 74800,
      "max_fee": 74800,
      "min_fee": 74800,
      "tx_count": 1
    }
    ```

#### Action Endpoints

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
          "TxHash": "3cMm96hvrC4Pg6ffKgspodRUCiqqDztYzVumaJtGeD3772hmP",
          "ActionType": 0,
          "ActionName": "Transfer",
          "ActionIndex": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 5000000000
          },
          "Output": {
            "receiver_balance": 5000000000,
            "sender_balance": 852999994999876700
          },
          "Timestamp": "2024-12-03T21:18:36Z"
        },
        {
          "ID": 1,
          "TxHash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
          "ActionType": 4,
          "ActionName": "CreateAsset",
          "ActionIndex": 0,
          "Input": {
            "asset_type": 0,
            "decimals": 0,
            "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "max_supply": 0,
            "metadata": "test",
            "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "name": "Kiran",
            "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "symbol": "KP1"
          },
          "Output": {
            "asset_balance": 0,
            "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
            "nft_id": ""
          },
          "Timestamp": "2024-12-03T21:17:28Z"
        }
      ]
    }
    ```

- **Get Actions by Transaction Hash**

  - **Endpoint**: `/actions/:tx_hash`
  - **Description**: Retrieve actions associated with a specific transaction hash.
  - **Example**: `curl http://localhost:8080/actions/tx_hash_here`
  - **Output**:

    ```bash
    [
      {
        "ID": 3,
        "TxHash": "3cMm96hvrC4Pg6ffKgspodRUCiqqDztYzVumaJtGeD3772hmP",
        "ActionType": 0,
        "ActionName": "Transfer",
        "ActionIndex": 0,
        "Input": {
          "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
          "memo": "",
          "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
          "value": 5000000000
        },
        "Output": {
          "receiver_balance": 5000000000,
          "sender_balance": 852999994999876700
        },
        "Timestamp": "2024-12-03T21:18:36Z"
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
        "ID": 1,
        "TxHash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
        "ActionType": 4,
        "ActionName": "CreateAsset",
        "ActionIndex": 0,
        "Input": {
          "asset_type": 0,
          "decimals": 0,
          "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "max_supply": 0,
          "metadata": "test",
          "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "name": "Kiran",
          "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "symbol": "KP1"
        },
        "Output": {
          "asset_balance": 0,
          "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
          "nft_id": ""
        },
        "Timestamp": "2024-12-03T21:17:28Z"
      }
    ]
    ```

- **Get Actions by Action Type**

  - **Endpoint**: `/actions/type/:action_type`
  - **Parameters**:
    - `action_type`: Action type ID (e.g., 0 for "Transfer").
    - `limit`: Number of actions to return (default: 10).
    - `offset`: Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/actions/type/0?limit=5&offset=0"`
  - **Output**:

    ```bash
    {
      "counter": 2,
      "items": [
        {
          "ID": 3,
          "TxHash": "3cMm96hvrC4Pg6ffKgspodRUCiqqDztYzVumaJtGeD3772hmP",
          "ActionType": 0,
          "ActionName": "Transfer",
          "ActionIndex": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 5000000000
          },
          "Output": {
            "receiver_balance": 5000000000,
            "sender_balance": 852999994999876700
          },
          "Timestamp": "2024-12-03T21:18:36Z"
        }
      ]
    }
    ```

- **Get Actions by Action Name**

  - **Endpoint**: `/actions/name/:action_name`
  - **Parameters**:
    - `action_name`: Action name (case-insensitive, e.g., "Transfer").
    - `limit`: Number of actions to return (default: 10).
    - `offset`: Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/actions/name/transfer?limit=5&offset=0"`
  - **Output**:

    ```bash
    {
      "counter": 2,
      "items": [
        {
          "ID": 3,
          "TxHash": "3cMm96hvrC4Pg6ffKgspodRUCiqqDztYzVumaJtGeD3772hmP",
          "ActionType": 0,
          "ActionName": "Transfer",
          "ActionIndex": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 5000000000
          },
          "Output": {
            "receiver_balance": 5000000000,
            "sender_balance": 852999994999876700
          },
          "Timestamp": "2024-12-03T21:18:36Z"
        }
      ]
    }
    ```

- **Get Actions For a specific user**

  - **Endpoint**: `/actions/user/:user`
  - **Description**: Retrieves all actions associated with the specified user. Supports both plain and 0x-prefixed identifiers.
  - **Parameters**:
    - Limit (optional): Number of actions to return (default: 10).
    - offset (optional): Offset for pagination (default: 0).
  - **Example**: `curl http://localhost:8080/actions/user/0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9`
  - **Output**:

    ```bash
    {
      "counter": 2,
      "items": [
        {
          "ID": 3,
          "TxHash": "3cMm96hvrC4Pg6ffKgspodRUCiqqDztYzVumaJtGeD3772hmP",
          "ActionType": 0,
          "ActionName": "Transfer",
          "ActionIndex": 0,
          "Input": {
            "asset_address": "0x00cf77495ce1bdbf11e5e45463fad5a862cb6cc0a20e00e658c4ac3355dcdc64bb",
            "memo": "",
            "to": "0x006835bfa9c67557da9fe6c7ad69089e17c6cad3e18284e037c78aa307e3c0c840",
            "value": 5000000000
          },
          "Output": {
            "receiver_balance": 5000000000,
            "sender_balance": 852999994999876700
          },
          "Timestamp": "2024-12-03T21:18:36Z"
        },
        {
          "ID": 1,
          "TxHash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
          "ActionType": 4,
          "ActionName": "CreateAsset",
          "ActionIndex": 0,
          "Input": {
            "asset_type": 0,
            "decimals": 0,
            "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "max_supply": 0,
            "metadata": "test",
            "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "name": "Kiran",
            "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "symbol": "KP1"
          },
          "Output": {
            "asset_balance": 0,
            "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
            "nft_id": ""
          },
          "Timestamp": "2024-12-03T21:17:28Z"
        }
      ]
    }
    ```

#### Assets Endpoints

- **Get All Assets**

  - **Endpoint**: `/assets`
  - **Description**: Retrieve all assets stored in the database with pagination.
  - **Parameters**:
    - limit (optional): Number of assets to return (default: 10).
    - offset (optional): Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/assets?limit=5&offset=0"`
  - **Output**:

    ```json
    {
      "counter": 1,
      "items": [
        {
          "id": 1,
          "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
          "asset_type_id": 0,
          "asset_type": "fungible",
          "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "tx_hash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
          "name": "Kiran",
          "symbol": "KP1",
          "decimals": 0,
          "metadata": "test",
          "max_supply": 0,
          "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "timestamp": "2024-12-03T21:17:28Z"
        }
      ]
    }
    ```

- **Get Assets by Type**

  - **Endpoint**: `/assets/type/:type`
  - **Description**: Retrieve all assets of a specific type.
  - **Path Parameters**:
    - type: Asset type ID (0 = fungible, 1 = non-fungible, 2 = fractional).
  - **Parameters**:
    - limit (optional): Number of assets to return (default: 10).
    - offset (optional): Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/assets/type/0?limit=5&offset=0"`
  - **Output**:

    ```json
    {
      "counter": 1,
      "items": [
        {
          "id": 1,
          "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
          "asset_type_id": 0,
          "asset_type": "fungible",
          "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "tx_hash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
          "name": "Kiran",
          "symbol": "KP1",
          "decimals": 0,
          "metadata": "test",
          "max_supply": 0,
          "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "timestamp": "2024-12-03T21:17:28Z"
        }
      ]
    }
    ```

- **Get Assets by User**

  - **Endpoint**: `/assets/user/:user`
  - **Description**: Retrieve all assets created by a specific user.
  - **Path Parameters**:
    - user: The user's address (supports both plain and 0x-prefixed addresses).
  - **Parameters**:
    - limit (optional): Number of assets to return (default: 10).
    - offset (optional): Offset for pagination (default: 0).
  - **Example**: `curl "http://localhost:8080/assets/user/00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9?limit=5&offset=0"`
  - **Output**:

    ```json
    {
      "counter": 1,
      "items": [
        {
          "id": 1,
          "asset_id": "00e751a0142a82aae56048bcc35d61bf23a43a364ba2dd76f72fddf75764b6f75f",
          "asset_type_id": 0,
          "asset_type": "fungible",
          "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "tx_hash": "2KGTRrhqvSzBjQDfSA8JHFJvPXmLnxzdKUdNASeYcgRSqFGGrh",
          "name": "Kiran",
          "symbol": "KP1",
          "decimals": 0,
          "metadata": "test",
          "max_supply": 0,
          "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
          "timestamp": "2024-12-03T21:17:28Z"
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
