// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package models

type MsgSend struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      []Coin `json:"amount"`
}

type MsgMultiSend struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

type Input struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}

type Output struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}

type MsgDelegate struct {
	DelegatorAddress string `json:"delegator_address"`
	ValidatorAddress string `json:"validator_address"`
	Amount           Coin   `json:"amount"`
}

// - MsgUndelegate
// - MsgBeginRedelegate
// - MsgCreateValidator
// - MsgSubmitProposal
// Others
