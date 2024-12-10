# Genesis API

## Get Genesis Data

- **Endpoint**: `/genesis`
- **Description**: Retrieve the genesis data.
- **Example**: `curl http://localhost:8080/genesis`
- **Output**:

```json
{
  "customAllocation": [
    {
      "address": "0x00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
      "balance": 853000000000000000
    }
  ],
  "emissionBalancer": {
    "emissionAddress": "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9",
    "maxSupply": 1e19
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
    "sponsorStateKeysMaxChunks": [1],
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
