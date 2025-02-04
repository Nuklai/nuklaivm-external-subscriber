# Account APIs

## Get Account Statistics

- **Endpoint**: `/accounts/stats`
- **Description**: Retrieves specific account stats.
- **Example**: `curl http://localhost:8080/accounts/stats`
- **Output**:

```json
{
  "total_accounts": 4720372,
  "total_nai_held": 83545852999978599951500,
  "active_accounts": 28358
}
```

## Get All Accounts

- **Endpoint**: `/accounts`
- **Description**: Retrieves All account on the NuklaiVM, their balances, and the number of transactions they have made.
- **Parameters**:
  - `limit`: Number of accounts to return (default: 20).
  - `offset`: Offset for pagination (default: 0).
- **Example**: `curl "http://localhost:8080/accounts?limit=2&offset=0"`
- **Output**:

```json
{
  "counter": 2,
  "items": [
    {
      "address": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "balance": 649999978599951500,
      "transaction_count": 1
    },
    {
      "address": "0145cc7292db91409269dc567c8be003224c1e29e0b7c30f83c7e93992a29ef700",
      "balance": 42150000000,
      "transaction_count": 1
    }
  ]
}
```

## Get Account Details

- **Endpoint**: `/accounts/:address`
- **Description**: Retrieves account details for a specific account.
- **Example**: `curl http://localhost:8080/accounts/00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9`
- **Output**:

```json
{
  "address":"00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
  "balance": 649999978599951500,
  "transaction_count": 1
}
```
