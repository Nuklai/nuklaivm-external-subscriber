# Transactions APIs

## Get All Transactions

- **Endpoint**: `/transactions`
- **Parameters**:
  - `tx_hash`: Filter by transaction hash.
  - `block_hash`: Filter by block hash.
  - `action_type`: Filter by action type(eg. 0 for transfer, 4 for createasset).
  - `action_name`: Filter by action name(eg. transfer, createasset). This is case insensitive.
  - `user`: Filter by user(case insensitive/with or without 0x prefix).
  - `limit`: Number of transactions to return (default: 10).
  - `offset`: Offset for pagination (default: 0).

### Get the last 2 transactions

- **Example**: `curl "http://localhost:8080/transactions?limit=2&offset=0"`
- **Output**:

```json
{
  "counter": 2,
  "items": [
    {
      "ID": 3,
      "TxHash": "xxnhyCwDAaqQ7oW8WWGuctcxKHWER76EWsBK6xfsj5MaEZHUK",
      "BlockHash": "2mhyYEw9LCkGfAUc8jsPawUMVmmdz83YDdhRyfZnNxfYgNTi36",
      "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "MaxFee": 58500,
      "Success": true,
      "Fee": 75000,
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
            "metadata": "test1",
            "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "name": "Kiran1",
            "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "symbol": "KP1"
          },
          "Output": {
            "asset_address": "00cc1b688e61ca24a3ad49007263f61b983fb953db5dca7fbb57bcbc0984a8f06e",
            "asset_balance": 0,
            "dataset_parent_nft_address": ""
          }
        }
      ],
      "Timestamp": "2024-12-10T15:17:49Z"
    },
    {
      "ID": 1,
      "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
      "BlockHash": "apSs1J24ppuNu2RoXtRZRoiq8jSVdRWD4f3YZqaEiY2zYmmMU",
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
            "sender_balance": 852999994999951500
          }
        }
      ],
      "Timestamp": "2024-12-10T15:15:04Z"
    }
  ]
}
```

### Get transactions and filter by tx_hash

- **Example**: `curl "http://localhost:8080/transactions?tx_hash=2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC"`

### Get transactions and filter by block_hash

- **Example**: `curl "http://localhost:8080/transactions?block_hash=apSs1J24ppuNu2RoXtRZRoiq8jSVdRWD4f3YZqaEiY2zYmmMU"`

### Get transactions and filter by action_type

- **Example**: `curl "http://localhost:8080/transactions?action_type=0"`

### Get transactions and filter by action_name

- **Example**: `curl "http://localhost:8080/transactions?action_name=transfer"`

### Get transactions and filter by user

- **Example**: `curl "http://localhost:8080/transactions?user=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9"`

## Get Transaction by Hash

- **Endpoint**: `/transactions/:tx_hash`
- **Description**: Retrieve transaction by its transaction hash
- **Example**: `curl http://localhost:8080/transactions/2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC`
  - **Output**:

```json
{
  "ID": 1,
  "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
  "BlockHash": "apSs1J24ppuNu2RoXtRZRoiq8jSVdRWD4f3YZqaEiY2zYmmMU",
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
        "sender_balance": 852999994999951500
      }
    }
  ],
  "Timestamp": "2024-12-10T15:15:04Z"
}
```

## Get Transactions by Block

- **Endpoint**: `/transactions/block/:identifier`
- **Description**: Retrieve transactions associated with a block, specified by either block height or hash.
- **Example**: `curl http://localhost:8080/transactions/block/701` or `curl http://localhost:8080/transactions/block/apSs1J24ppuNu2RoXtRZRoiq8jSVdRWD4f3YZqaEiY2zYmmMU`
- **Output**:

```json
[
  {
    "ID": 1,
    "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
    "BlockHash": "apSs1J24ppuNu2RoXtRZRoiq8jSVdRWD4f3YZqaEiY2zYmmMU",
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
          "sender_balance": 852999994999951500
        }
      }
    ],
    "Timestamp": "2024-12-10T15:15:04Z"
  }
]
```

## Get Transactions For a specific user

- **Endpoint**: `/transactions/user/:user`
- **Description**: Retrieves all transactions associated with the specified user(case insensitive/with or without 0x prefix).
- **Parameters**:
  - Limit (optional): Number of transactions to return (default: 10).
  - offset (optional): Offset for pagination (default: 0).
- **Example**: `curl http://localhost:8080/transactions/user/00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9`
- **Output**:

```json
{
  "counter": 2,
  "items": [
    {
      "ID": 3,
      "TxHash": "xxnhyCwDAaqQ7oW8WWGuctcxKHWER76EWsBK6xfsj5MaEZHUK",
      "BlockHash": "2mhyYEw9LCkGfAUc8jsPawUMVmmdz83YDdhRyfZnNxfYgNTi36",
      "Sponsor": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "MaxFee": 58500,
      "Success": true,
      "Fee": 75000,
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
            "metadata": "test1",
            "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "name": "Kiran1",
            "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
            "symbol": "KP1"
          },
          "Output": {
            "asset_address": "00cc1b688e61ca24a3ad49007263f61b983fb953db5dca7fbb57bcbc0984a8f06e",
            "asset_balance": 0,
            "dataset_parent_nft_address": ""
          }
        }
      ],
      "Timestamp": "2024-12-10T15:17:49Z"
    },
    {
      "ID": 1,
      "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
      "BlockHash": "apSs1J24ppuNu2RoXtRZRoiq8jSVdRWD4f3YZqaEiY2zYmmMU",
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
            "sender_balance": 852999994999951500
          }
        }
      ],
      "Timestamp": "2024-12-10T15:15:04Z"
    }
  ]
}
```

## Get Transfer Volumes

- **Endpoint**: `/transactions/volumes`
- **Description**: Retrieves total transfer volume for 12 hours, 24 hours, 7 days, 30 days
- **Example**: `curl http://localhost:8080/transactions/volumes`
- **Output**:

```json
{
  "12_hours": 22500000000,
  "24_hours": 22500000000,
  "7_days": 22500000000,
  "30_days": 22500000000
}
```

## Get Aggregated Estimated Fees for different Transactions

- **Endpoint**: `/transactions/estimated_fee`
- **Description**: Retrieve the average, minimum, and maximum fees for all transaction types and names over a specified time interval.
- **Query Parameters**:
  - `interval` (optional): Time interval for which the estimated fee is calculated (e.g., 1m, 1h, 1d). Default is `1m`.
- **Example**: `curl http://localhost:8080/transactions/estimated_fee?interval=1h`
- **Output**:

```json
{
  "avg_fee": 61750,
  "min_fee": 48500,
  "max_fee": 75000,
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
      "avg_fee": 75000,
      "max_fee": 75000,
      "min_fee": 75000,
      "tx_count": 1
    }
  ]
}
```

## Get Estimated Fee for different Transactions by action type

- **Endpoint**: `/transactions/estimated_fee/action_type/:action_type`
- **Description**: Retrieve the average, minimum, and maximum fees for transactions of a specific action type.
- **Path Parameters**:
  - `action_type`: The ID of the action type (e.g., 0 for "Transfer").
- **Query Parameters**:
  - `interval` (optional): Time interval for which the estimated fee is calculated (e.g., 1m, 1h, 1d). Default is `1m`.
- **Example**: `curl http://localhost:8080/transactions/estimated_fee/action_type/0?interval=1h`
- **Output**:

```json
{
  "action_name": "Transfer",
  "action_type": 0,
  "avg_fee": 48500,
  "max_fee": 48500,
  "min_fee": 48500,
  "tx_count": 1
}
```

## Get Estimated Fee for different Transactions by action name

- **Endpoint**: `/transactions/estimated_fee/action_name/:action_name`
- **Description**: Retrieve the average, minimum, and maximum fees for transactions of a specific action name.
- **Path Parameters**:
  - `action_name`: The name of the action (e.g., "Transfer", "CreateAsset", etc). Case-insensitive.
- **Query Parameters**:
  - `interval` (optional): Time interval for which the estimated fee is calculated (e.g., 1m, 1h, 1d). Default is `1m`.
- **Example**: `curl http://localhost:8080/transactions/estimated_fee/action_name/createasset?interval=1h`
- **Output**:

```json
{
  "action_name": "CreateAsset",
  "action_type": 4,
  "avg_fee": 75000,
  "max_fee": 75000,
  "min_fee": 75000,
  "tx_count": 1
}
```
