# Validator Stake APIs

## Get All Validator Stakes

- **Endpoint**: `/validator_stake`
- **Parameters**:
  - `limit`: Number of validator stakes to return (default: 10).
  - `offset`: Offset for pagination (default: 0).
- **Example**: `curl "http://localhost:8080/validator_stake?limit=2&offset=0"`
- **Output**:

```json
{
  "counter": 2,
  "items": [
    {
      "node_id": "NodeID-Nxy5Q8K9YkLasVKkdd4ftaHnVwdSPnKE5",
      "actor": "02299b842c7c90de831f025d9670be2449007c1bb84cafa7b02680d2f953a541ed",
      "stake_start_block": 7600,
      "stake_end_block": 10000000,
      "staked_amount": 100000000000,
      "delegation_fee_rate": 90,
      "reward_address": "02299b842c7c90de831f025d9670be2449007c1bb84cafa7b02680d2f953a541ed",
      "tx_hash": "2oz5TRCtbWUVkYXA4TGqrWJ9Ho2Vhe9eBrt7TfLWxZrEW9s6Xp",
      "timestamp": "2024-12-26T20:07:00Z"
    },
    {
      "node_id": "NodeID-51nr959VoL7doZhm6QKS4CFK4TM77z9ta",
      "actor": "02ffe89807f1915c66d56be575a68a3c4c232eaf0f92e794dcc3e26a5bc78ecd6f",
      "stake_start_block": 7248,
      "stake_end_block": 7548,
      "staked_amount": 100000000000,
      "delegation_fee_rate": 50,
      "reward_address": "02ffe89807f1915c66d56be575a68a3c4c232eaf0f92e794dcc3e26a5bc78ecd6f",
      "tx_hash": "MMpzLJU2fpgTZoRThZ3f5dgZR35mZ3UecZZNieXqxWxuGM6U5",
      "timestamp": "2024-12-26T20:03:38Z"
    }
  ]
}
```

## Get Validator Stake by Node ID

- **Endpoint**: `/validator_stake/:node_id`
- **Description**: Retrieve a specific validator stake by its node ID.
- **Example**: `curl http://localhost:8080/validator_stake/NodeID-Nxy5Q8K9YkLasVKkdd4ftaHnVwdSPnKE5`
- **Output**:

```json
{
      "node_id": "NodeID-Nxy5Q8K9YkLasVKkdd4ftaHnVwdSPnKE5",
      "actor": "02299b842c7c90de831f025d9670be2449007c1bb84cafa7b02680d2f953a541ed",
      "stake_start_block": 7600,
      "stake_end_block": 10000000,
      "staked_amount": 100000000000,
      "delegation_fee_rate": 90,
      "reward_address": "02299b842c7c90de831f025d9670be2449007c1bb84cafa7b02680d2f953a541ed",
      "tx_hash": "2oz5TRCtbWUVkYXA4TGqrWJ9Ho2Vhe9eBrt7TfLWxZrEW9s6Xp",
      "timestamp": "2024-12-26T20:07:00Z"
}
```
