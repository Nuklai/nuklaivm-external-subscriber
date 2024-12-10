# Assets APIs

## Get All Assets

- **Endpoint**: `/assets`
- **Description**: Retrieve all assets stored in the database with pagination.
- **Parameters**:
  - `type`: Filter by asset type(0 for fungible, 1 for non-fungible, 3 for fractional)
  - `user`: Filter by user(case insensitive/with or without 0x prefix).
  - `asset_address`: Filter by asset address(case insensitive/with or without 0x prefix).
  - `name`: Filter by name(case insensitive).
  - `symbol`: Filter by symbol(case insensitive).
  - `limit`: Number of assets to return (default: 10).
  - `offset`: Offset for pagination (default: 0).

### Get the last 2 created assets

- **Example**: `curl "http://localhost:8080/assets?limit=2&offset=0"`
- **Output**:

```json
{
  "counter": 2,
  "items": [
    {
      "id": 3,
      "asset_address": "01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706",
      "asset_type_id": 1,
      "asset_type": "non-fungible",
      "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "tx_hash": "18Mv4d3YrnNhz5PbhLmJCfeciEWsuwwABe1SYDbMnmW3dtgjW",
      "name": "Kiran2",
      "symbol": "KP2",
      "decimals": 0,
      "metadata": "test2",
      "max_supply": 0,
      "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "timestamp": "2024-12-10T15:40:06Z"
    },
    {
      "id": 1,
      "asset_address": "00cc1b688e61ca24a3ad49007263f61b983fb953db5dca7fbb57bcbc0984a8f06e",
      "asset_type_id": 0,
      "asset_type": "fungible",
      "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "tx_hash": "xxnhyCwDAaqQ7oW8WWGuctcxKHWER76EWsBK6xfsj5MaEZHUK",
      "name": "Kiran1",
      "symbol": "KP1",
      "decimals": 0,
      "metadata": "test1",
      "max_supply": 0,
      "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "timestamp": "2024-12-10T15:17:49Z"
    }
  ]
}
```

### Get assets and filter by type

- **Example**: `curl "http://localhost:8080/assets?type=0"`

### Get assets and filter by user

- **Example**: `curl "http://localhost:8080/assets?user=00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9"`

### Get assets and filter by asset_address

- **Example**: `curl "http://localhost:8080/assets?asset_address=01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706"`

### Get assets and filter by name

- **Example**: `curl "http://localhost:8080/assets?name=Kiran"`

### Get assets and filter by symbol

- **Example**: `curl "http://localhost:8080/assets?symbol=KP"`

## Get Assets by Asset Address

- **Endpoint**: `/assets/:asset_address`
- **Description**: Retrieve the asset info whose asset_address matches the value passed.
- **Path Parameters**:
  - asset_address: Asset Address(eg. `01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706`)
- **Example**: `curl http://localhost:8080/assets/01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706`
- **Output**:

```json
{
  "id": 3,
  "asset_address": "01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706",
  "asset_type_id": 1,
  "asset_type": "non-fungible",
  "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
  "tx_hash": "18Mv4d3YrnNhz5PbhLmJCfeciEWsuwwABe1SYDbMnmW3dtgjW",
  "name": "Kiran2",
  "symbol": "KP2",
  "decimals": 0,
  "metadata": "test2",
  "max_supply": 0,
  "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
  "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
  "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
  "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
  "timestamp": "2024-12-10T15:40:06Z"
}
```

## Get Assets by Type

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
      "asset_address": "00cc1b688e61ca24a3ad49007263f61b983fb953db5dca7fbb57bcbc0984a8f06e",
      "asset_type_id": 0,
      "asset_type": "fungible",
      "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "tx_hash": "xxnhyCwDAaqQ7oW8WWGuctcxKHWER76EWsBK6xfsj5MaEZHUK",
      "name": "Kiran1",
      "symbol": "KP1",
      "decimals": 0,
      "metadata": "test1",
      "max_supply": 0,
      "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "timestamp": "2024-12-10T15:17:49Z"
    }
  ]
}
```

## Get Assets by User

- **Endpoint**: `/assets/user/:user`
- **Description**: Retrieve all assets created by a specific user.
- **Path Parameters**:
  - user: The user's address(case insensitive/with or without 0x prefix).
- **Parameters**:
  - limit (optional): Number of assets to return (default: 10).
  - offset (optional): Offset for pagination (default: 0).
- **Example**: `curl "http://localhost:8080/assets/user/00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9?limit=5&offset=0"`
- **Output**:

```json
{
  "counter": 2,
  "items": [
    {
      "id": 3,
      "asset_address": "01460b81b3f0da802affe8d9f4fb4d0d1d63ae4e9876227572c7713bc21b8ab706",
      "asset_type_id": 1,
      "asset_type": "non-fungible",
      "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "tx_hash": "18Mv4d3YrnNhz5PbhLmJCfeciEWsuwwABe1SYDbMnmW3dtgjW",
      "name": "Kiran2",
      "symbol": "KP2",
      "decimals": 0,
      "metadata": "test2",
      "max_supply": 0,
      "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "timestamp": "2024-12-10T15:40:06Z"
    },
    {
      "id": 1,
      "asset_address": "00cc1b688e61ca24a3ad49007263f61b983fb953db5dca7fbb57bcbc0984a8f06e",
      "asset_type_id": 0,
      "asset_type": "fungible",
      "asset_creator": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "tx_hash": "xxnhyCwDAaqQ7oW8WWGuctcxKHWER76EWsBK6xfsj5MaEZHUK",
      "name": "Kiran1",
      "symbol": "KP1",
      "decimals": 0,
      "metadata": "test1",
      "max_supply": 0,
      "mint_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "pause_unpause_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "freeze_unfreeze_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "enable_disable_kyc_account_admin": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "timestamp": "2024-12-10T15:17:49Z"
    }
  ]
}
```
