# Blocks APIs

## Get All Blocks

- **Endpoint**: `/blocks`
- **Parameters**:
  - `block_height`: (optional) Get specific block by it's height.
  - `block_hash`: (optional) Get specific block by it's hash.
  - `limit`: Number of blocks to return (default: 10).
  - `offset`: Offset for pagination (default: 0).
- **Example**: `curl "http://localhost:8080/blocks?limit=2&offset=0"`
- **Output**:

```json
{
  "counter": 796,
  "items": [
    {
      "BlockHeight": 796,
      "BlockHash": "8RvoHNH41WY2fEXxSmDNMBudtBSB8UUhezeHyF3WW7LTzeQ4B",
      "ParentBlockHash": "Kbo5Vq3R9P3nM9XRgtwKMwvqDCdqRc2BQoigQkirm6oQf67uL",
      "StateRoot": "ESeo5TFTP58DmskhpkiSp6CHJRdDCU9PhBDdW8N66esbqbcDg",
      "BlockSize": 84,
      "TxCount": 0,
      "TotalFee": 0,
      "AvgTxSize": 0,
      "UniqueParticipants": 0,
      "Timestamp": "2024-12-10T15:16:16Z"
    },
    {
      "BlockHeight": 795,
      "BlockHash": "Kbo5Vq3R9P3nM9XRgtwKMwvqDCdqRc2BQoigQkirm6oQf67uL",
      "ParentBlockHash": "MAYdPGwNxCZKZ8HDUWDq8xDA2C8ieZWxH9AZ5cCjjgnYyzeCV",
      "StateRoot": "2JRAjJ9Vw4TTw7eK7WfeWAq92UnSbJZHKSQNw729whSqdEhTKL",
      "BlockSize": 84,
      "TxCount": 0,
      "TotalFee": 0,
      "AvgTxSize": 0,
      "UniqueParticipants": 0,
      "Timestamp": "2024-12-10T15:16:15Z"
    }
  ]
}
```

## Get Block by Height or Hash

- **Description**: Retrieve a block by its height or hash.
- **Example**: `curl "http://localhost:8080/blocks?block_height=701"` or `curl "http://localhost:8080/blocks?block_hash=apSs1J24ppuNu2RoXtRZRoiq8jSVdRWD4f3YZqaEiY2zYmmMU"`
- **Output**:

```json
{
  "BlockHeight": 701,
  "BlockHash": "apSs1J24ppuNu2RoXtRZRoiq8jSVdRWD4f3YZqaEiY2zYmmMU",
  "ParentBlockHash": "2uCE1tkQPymEoksdnDWkXTA23c1WJiGN8bxY1tJCdfo34d6PCt",
  "StateRoot": "2cUqJpLg1HhEZthr5sCybPA8v7ViHRqSHqzLDJv9aZCzCqwkSx",
  "BlockSize": 307,
  "TxCount": 1,
  "TotalFee": 48500,
  "AvgTxSize": 307,
  "UniqueParticipants": 1,
  "Timestamp": "2024-12-10T15:15:04Z"
}
```
