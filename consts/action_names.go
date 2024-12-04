// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package consts

import "github.com/nuklai/nuklaivm/consts"

var ActionNames = map[uint8]string{
	consts.TransferID:                    "Transfer",
	consts.ContractCallID:                "ContractCall",
	consts.ContractDeployID:              "ContractDeploy",
	consts.ContractPublishID:             "ContractPublish",
	consts.CreateAssetID:                 "CreateAsset",
	consts.UpdateAssetID:                 "UpdateAsset",
	consts.MintAssetFTID:                 "MintAssetFT",
	consts.MintAssetNFTID:                "MintAssetNFT",
	consts.BurnAssetFTID:                 "BurnAssetFT",
	consts.BurnAssetNFTID:                "BurnAssetNFT",
	consts.RegisterValidatorStakeID:      "RegisterValidatorStake",
	consts.WithdrawValidatorStakeID:      "WithdrawValidatorStake",
	consts.ClaimValidatorStakeRewardsID:  "ClaimValidatorStakeRewards",
	consts.DelegateUserStakeID:           "DelegateUserStake",
	consts.UndelegateUserStakeID:         "UndelegateUserStake",
	consts.ClaimDelegationStakeRewardsID: "ClaimDelegationStakeRewards",
	consts.CreateDatasetID:               "CreateDataset",
	consts.UpdateDatasetID:               "UpdateDataset",
	consts.InitiateContributeDatasetID:   "InitiateContributeDataset",
	consts.CompleteContributeDatasetID:   "CompleteContributeDataset",
	consts.PublishDatasetMarketplaceID:   "PublishDatasetMarketplace",
	consts.SubscribeDatasetMarketplaceID: "SubscribeDatasetMarketplace",
	consts.ClaimMarketplacePaymentID:     "ClaimMarketplacePayment",
}
