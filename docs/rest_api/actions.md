# Actions APIs

## Get All Actions

- **Endpoint**: `/actions`
- **Parameters**:
  - `limit`: Number of actions to return (default: 10).
  - `offset`: Offset for pagination (default: 0).
- **Example**: `curl "http://localhost:8080/actions?limit=2&offset=0"`
- **Output**:

```json
{
  "counter": 2,
  "items": [
    {
      "ID": 3,
      "TxHash": "xxnhyCwDAaqQ7oW8WWGuctcxKHWER76EWsBK6xfsj5MaEZHUK",
      "ActionType": 4,
      "ActionName": "CreateAsset",
      "ActionIndex": 0,
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
      },
      "Timestamp": "2024-12-10T15:17:49Z"
    },
    {
      "ID": 1,
      "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
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
        "sender_balance": 852999994999951500
      },
      "Timestamp": "2024-12-10T15:15:04Z"
    }
  ]
}
```

## Get Actions by Transaction Hash

- **Endpoint**: `/actions/:tx_hash`
- **Description**: Retrieve actions associated with a specific transaction hash.
- **Example**: `curl http://localhost:8080/actions/tx_hash_here`
- **Output**:

```json
[
  {
    "ID": 1,
    "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
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
      "sender_balance": 852999994999951500
    },
    "Timestamp": "2024-12-10T15:15:04Z"
  }
]
```

## Get Actions by Block

- **Endpoint**: `/actions/block/:identifier`
- **Description**: Retrieve actions within a block, specified by either block height or hash.
- **Example**: `curl http://localhost:8080/actions/block/701`
- **Output**:

```json
[
  {
    "ID": 1,
    "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
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
      "sender_balance": 852999994999951500
    },
    "Timestamp": "2024-12-10T15:15:04Z"
  }
]
```

## Get Actions by Action Type

- **Endpoint**: `/actions/type/:action_type`
- **Parameters**:
  - `action_type`: Action type ID (e.g., 0 for "Transfer").
  - `limit`: Number of actions to return (default: 10).
  - `offset`: Offset for pagination (default: 0).
- **Example**: `curl "http://localhost:8080/actions/type/0?limit=5&offset=0"`
- **Output**:

```json
{
  "counter": 1,
  "items": [
    {
      "ID": 1,
      "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
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
        "sender_balance": 852999994999951500
      },
      "Timestamp": "2024-12-10T15:15:04Z"
    }
  ]
}
```

## Get Actions by Action Name

- **Endpoint**: `/actions/name/:action_name`
- **Parameters**:
  - `action_name`: Action name (case-insensitive, e.g., "Transfer").
  - `limit`: Number of actions to return (default: 10).
  - `offset`: Offset for pagination (default: 0).
- **Example**: `curl "http://localhost:8080/actions/name/transfer?limit=5&offset=0"`
- **Output**:

```json
{
  "counter": 1,
  "items": [
    {
      "ID": 1,
      "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
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
        "sender_balance": 852999994999951500
      },
      "Timestamp": "2024-12-10T15:15:04Z"
    }
  ]
}
```

## Get Actions For a specific user

- **Endpoint**: `/actions/user/:user`
- **Description**: Retrieves all actions associated with the specified user(case insensitive/with or without 0x prefix).
- **Parameters**:
  - Limit (optional): Number of actions to return (default: 10).
  - offset (optional): Offset for pagination (default: 0).
- **Example**: `curl http://localhost:8080/actions/user/0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9`
- **Output**:

```json
{
  "counter": 2,
  "items": [
    {
      "ID": 3,
      "TxHash": "xxnhyCwDAaqQ7oW8WWGuctcxKHWER76EWsBK6xfsj5MaEZHUK",
      "ActionType": 4,
      "ActionName": "CreateAsset",
      "ActionIndex": 0,
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
      },
      "Timestamp": "2024-12-10T15:17:49Z"
    },
    {
      "ID": 1,
      "TxHash": "2Sib9Hch2ECYZCXq1Xy1YxMbJvSRzd5mm4jLFtkYBVYN17jAAC",
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
        "sender_balance": 852999994999951500
      },
      "Timestamp": "2024-12-10T15:15:04Z"
    }
  ]
}
```
