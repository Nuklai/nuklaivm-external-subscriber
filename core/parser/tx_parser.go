// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
)

type TxParser struct{}

func NewTxParser() *TxParser {
	return &TxParser{}
}

// Parse a txn and extracts their messages and signers
func (p *TxParser) ParseTx(txBytes []byte) (models.Messages, []string, string, error) {
	var txJSON map[string]interface{}
	err := json.Unmarshal(txBytes, &txJSON)
	if err != nil {
		// Might be a protobuf encoded txn
		// Empty and log error
		return models.Messages{}, []string{}, "", nil
	}

	// Extract messages
	var messages models.Messages
	var signers []string
	var memo string

	if body, ok := txJSON["body"].(map[string]interface{}); ok {
		// Extract memo
		if memoVal, ok := body["memo"].(string); ok {
			memo = memoVal
		}

		// Extract messages
		if msgs, ok := body["messages"].([]interface{}); ok {
			for _, msg := range msgs {
				msgMap, ok := msg.(map[string]interface{})
				if !ok {
					continue
				}

				// Extract type and content
				typeURL, _ := msgMap["@type"].(string)
				if typeURL == "" {
					continue
				}

				// Remove the "/" prefix if present
				typeName := strings.TrimPrefix(typeURL, "/")

				// Convert the message to JSON
				msgJSON, err := json.Marshal(msgMap)
				if err != nil {
					continue
				}

				// Add the message to the list
				messages = append(messages, models.Message{
					Type:    typeName,
					Content: msgJSON,
				})

				// Try and extract signers from the message
				if from, ok := msgMap["from_address"].(string); ok {
					signers = appendIfUnique(signers, from)
				}
				if from, ok := msgMap["sender"].(string); ok {
					signers = appendIfUnique(signers, from)
				}
				if delegator, ok := msgMap["delegator_address"].(string); ok {
					signers = appendIfUnique(signers, delegator)
				}
				if voter, ok := msgMap["voter"].(string); ok {
					signers = appendIfUnique(signers, voter)
				}
				if proposer, ok := msgMap["proposer"].(string); ok {
					signers = appendIfUnique(signers, proposer)
				}
			}
		}
	}

	// Check if we have auth_info for more signer info
	if authInfo, ok := txJSON["auth_info"].(map[string]interface{}); ok {
		if signerInfos, ok := authInfo["signer_infos"].([]interface{}); ok {
			for _, signerInfo := range signerInfos {
				siMap, ok := signerInfo.(map[string]interface{})
				if !ok {
					continue
				}

				if pubKey, ok := siMap["public_key"].(map[string]interface{}); ok {
					if _, ok := pubKey["key"].(string); ok {
					}
				}
			}
		}
	}

	return messages, signers, memo, nil
}

// ExtractFee extracts the fee from a transaction
func (p *TxParser) ExtractFee(txBytes []byte) (string, error) {
	// Try to parse the transaction as JSON
	var txJSON map[string]interface{}
	err := json.Unmarshal(txBytes, &txJSON)
	if err != nil {
		return "0", fmt.Errorf("could not parse transaction: %w", err)
	}

	// Check if we have auth_info
	if authInfo, ok := txJSON["auth_info"].(map[string]interface{}); ok {
		if fee, ok := authInfo["fee"].(map[string]interface{}); ok {
			if amount, ok := fee["amount"].([]interface{}); ok && len(amount) > 0 {
				// Convert the fee to string
				feeStr := ""
				for i, coin := range amount {
					coinMap, ok := coin.(map[string]interface{})
					if !ok {
						continue
					}

					denom, _ := coinMap["denom"].(string)
					coinAmount, _ := coinMap["amount"].(string)

					if i > 0 {
						feeStr += ","
					}
					feeStr += coinAmount + denom
				}

				if feeStr != "" {
					return feeStr, nil
				}
			}
		}
	}

	return "0", nil
}

// MapEventTypeToModule maps an event type to its module
func MapEventTypeToModule(eventType string) string {
	// Split the event type on dots
	parts := strings.Split(eventType, ".")
	if len(parts) > 0 {
		return parts[0]
	}

	// Special cases
	switch {
	case strings.HasPrefix(eventType, "message"):
		return "sdk"
	case strings.HasPrefix(eventType, "tx"):
		return "sdk"
	case strings.HasPrefix(eventType, "coin_spent"):
		return "bank"
	case strings.HasPrefix(eventType, "coin_received"):
		return "bank"
	case strings.HasPrefix(eventType, "transfer"):
		return "bank"
	case strings.HasPrefix(eventType, "withdraw_rewards"):
		return "distribution"
	case strings.HasPrefix(eventType, "delegate"):
		return "staking"
	case strings.HasPrefix(eventType, "unbond"):
		return "staking"
	case strings.HasPrefix(eventType, "redelegate"):
		return "staking"
	case strings.HasPrefix(eventType, "proposal_vote"):
		return "gov"
	case strings.HasPrefix(eventType, "proposal_deposit"):
		return "gov"
	case strings.HasPrefix(eventType, "instantiate_contract"):
		return "wasm"
	case strings.HasPrefix(eventType, "execute_contract"):
		return "wasm"
	case strings.HasPrefix(eventType, "migrate_contract"):
		return "wasm"
	case strings.HasPrefix(eventType, "store_code"):
		return "wasm"
	case strings.HasPrefix(eventType, "set_contract_admin"):
		return "wasm"
	case strings.HasPrefix(eventType, "clear_contract_admin"):
		return "wasm"
	case strings.HasPrefix(eventType, "ibc_transfer"):
		return "ibc"
	case strings.HasPrefix(eventType, "dataset_"):
		return "dataset"
	}

	return "unknown"
}

// Helper function to append a string to a slice if it's not already present
func appendIfUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

// Helper for parsing JSON transactions
func (p *TxParser) parseJSONTx(txJSON map[string]interface{}) (models.Messages, []string, string, error) {
	var messages models.Messages
	var signers []string
	var memo string

	if body, ok := txJSON["body"].(map[string]interface{}); ok {
		// Extract the memo
		if memoVal, ok := body["memo"].(string); ok {
			memo = memoVal
		}

		// Extract messages
		if msgs, ok := body["messages"].([]interface{}); ok {
			for _, msg := range msgs {
				msgMap, ok := msg.(map[string]interface{})
				if !ok {
					continue
				}

				// Extract type and content
				typeURL, _ := msgMap["@type"].(string)
				if typeURL == "" {
					continue
				}

				// Remove the "/" prefix if present
				typeName := strings.TrimPrefix(typeURL, "/")

				// Convert the message to JSON
				msgJSON, err := json.Marshal(msgMap)
				if err != nil {
					continue
				}

				messages = append(messages, models.Message{
					Type:    typeName,
					Content: msgJSON,
				})

				if from, ok := msgMap["from_address"].(string); ok {
					signers = appendIfUnique(signers, from)
				}
				if from, ok := msgMap["sender"].(string); ok {
					signers = appendIfUnique(signers, from)
				}
				if delegator, ok := msgMap["delegator_address"].(string); ok {
					signers = appendIfUnique(signers, delegator)
				}
				if voter, ok := msgMap["voter"].(string); ok {
					signers = appendIfUnique(signers, voter)
				}
				if proposer, ok := msgMap["proposer"].(string); ok {
					signers = appendIfUnique(signers, proposer)
				}
			}
		}
	}

	// Check if we have auth_info for more signer info
	if authInfo, ok := txJSON["auth_info"].(map[string]interface{}); ok {
		if signerInfos, ok := authInfo["signer_infos"].([]interface{}); ok {
			for _, signerInfo := range signerInfos {
				siMap, ok := signerInfo.(map[string]interface{})
				if !ok {
					continue
				}

				if pubKey, ok := siMap["public_key"].(map[string]interface{}); ok {
					if _, ok := pubKey["key"].(string); ok {
					}
				}
			}
		}
	}

	return messages, signers, memo, nil
}
