// Copyright (C) 2025, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package parser

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/nuklai/nuklaivm-external-subscriber/db/models"
	"github.com/nuklai/nuklaivm-external-subscriber/utils/logger"
	"go.uber.org/zap"
)

// Struct for TxRaw, TxBody, and AuthInfo
type txRaw struct {
	BodyBytes     []byte   `protobuf:"bytes,1,opt,name=body_bytes,json=bodyBytes,proto3" json:"body_bytes,omitempty"`
	AuthInfoBytes []byte   `protobuf:"bytes,2,opt,name=auth_info_bytes,json=authInfoBytes,proto3" json:"auth_info_bytes,omitempty"`
	Signatures    [][]byte `protobuf:"bytes,3,rep,name=signatures,proto3" json:"signatures,omitempty"`
}

type txBody struct {
	Messages []anyMsg `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
	Memo     string   `protobuf:"bytes,2,opt,name=memo,proto3" json:"memo,omitempty"`
}

type anyMsg struct {
	TypeUrl string `protobuf:"bytes,1,opt,name=type_url,json=typeUrl,proto3" json:"type_url,omitempty"`
	Value   []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

type authInfo struct {
	Fee *fee `protobuf:"bytes,2,opt,name=fee,proto3" json:"fee,omitempty"`
}

type fee struct {
	Amount []coin `protobuf:"bytes,1,rep,name=amount,proto3" json:"amount,omitempty"`
}

type coin struct {
	Denom  string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Amount string `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

// proto.Message interface for these types (simplified)
func (m *txRaw) Reset()         { *m = txRaw{} }
func (m *txRaw) String() string { return proto.CompactTextString(m) }
func (m *txRaw) ProtoMessage()  {}

func (m *txBody) Reset()         { *m = txBody{} }
func (m *txBody) String() string { return proto.CompactTextString(m) }
func (m *txBody) ProtoMessage()  {}

func (m *authInfo) Reset()         { *m = authInfo{} }
func (m *authInfo) String() string { return proto.CompactTextString(m) }
func (m *authInfo) ProtoMessage()  {}

type TxParser struct{}

func NewTxParser() *TxParser {
	return &TxParser{}
}

// Parse a txn and extracts their messages and signers
func (p *TxParser) ParseTx(txBytes []byte) (models.Messages, []string, string, error) {
	logger.Info("Parsing transaction", zap.Int("tx_bytes_length", len(txBytes)))

	var txJSON map[string]interface{}
	err := json.Unmarshal(txBytes, &txJSON)
	if err == nil {
		logger.Info("Transaction parsed as JSON")
		return p.parseJSONTx(txJSON)
	}

	logger.Info("Transaction is not JSON format, trying protobuf", zap.Error(err))

	messages, signers, memo, err := p.parseProtobufTx(txBytes)
	if err != nil {
		logger.Warn("Failed to parse protobuf transaction", zap.Error(err))
		return models.Messages{}, []string{}, "", nil
	}

	logger.Info("Successfully parsed protobuf transaction",
		zap.Int("messages_count", len(messages)),
		zap.Strings("signers", signers),
		zap.String("memo", memo))

	return messages, signers, memo, nil
}

// ExtractFee extracts the fee from a transaction
func (p *TxParser) ExtractFee(txBytes []byte) (string, error) {
	var txJSON map[string]interface{}
	err := json.Unmarshal(txBytes, &txJSON)
	if err == nil {
		if authInfo, ok := txJSON["auth_info"].(map[string]interface{}); ok {
			if fee, ok := authInfo["fee"].(map[string]interface{}); ok {
				if amount, ok := fee["amount"].([]interface{}); ok && len(amount) > 0 {
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

	var txRawData txRaw
	err = proto.Unmarshal(txBytes, &txRawData)
	if err != nil {
		return "0", nil
	}

	var authInfoData authInfo
	err = proto.Unmarshal(txRawData.AuthInfoBytes, &authInfoData)
	if err != nil {
		return "0", nil
	}

	if authInfoData.Fee != nil && len(authInfoData.Fee.Amount) > 0 {
		feeStr := ""
		for i, coin := range authInfoData.Fee.Amount {
			if i > 0 {
				feeStr += ","
			}
			feeStr += coin.Amount + coin.Denom
		}
		return feeStr, nil
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

// Parse a protobuf-encoded transaction
func (p *TxParser) parseProtobufTx(txBytes []byte) (models.Messages, []string, string, error) {
	var messages models.Messages
	var signers []string
	var memo string

	logger.Info("Attempting to parse protobuf transaction", zap.String("bytes_hex", hex.EncodeToString(txBytes[:min(100, len(txBytes))])))

	// unmarshal as TxRaw
	var txRawData txRaw
	err := proto.Unmarshal(txBytes, &txRawData)
	if err != nil {
		logger.Error("Failed to unmarshal TxRaw", zap.Error(err))
		return messages, signers, memo, fmt.Errorf("failed to unmarshal TxRaw: %w", err)
	}

	logger.Info("Successfully unmarshaled TxRaw",
		zap.Int("body_bytes_length", len(txRawData.BodyBytes)),
		zap.Int("auth_info_bytes_length", len(txRawData.AuthInfoBytes)),
		zap.Int("signatures_count", len(txRawData.Signatures)))

	var txBodyData txBody
	err = proto.Unmarshal(txRawData.BodyBytes, &txBodyData)
	if err != nil {
		logger.Error("Failed to unmarshal TxBody", zap.Error(err))
		return messages, signers, memo, fmt.Errorf("failed to unmarshal TxBody: %w", err)
	}

	// Get memo
	memo = txBodyData.Memo
	logger.Info("Successfully unmarshaled TxBody",
		zap.String("memo", memo),
		zap.Int("messages_count", len(txBodyData.Messages)))

	// Process messages
	for i, anyMsg := range txBodyData.Messages {
		typeURL := anyMsg.TypeUrl
		typeName := strings.TrimPrefix(typeURL, "/")

		logger.Info("Processing message",
			zap.Int("index", i),
			zap.String("type_url", typeURL),
			zap.Int("value_length", len(anyMsg.Value)),
			zap.String("value_hex", hex.EncodeToString(anyMsg.Value[:min(50, len(anyMsg.Value))])))

		var msgContent []byte

		msgObj := map[string]interface{}{
			"@type": typeURL,
			"value": hex.EncodeToString(anyMsg.Value),
		}

		msgContent, err = json.Marshal(msgObj)
		if err != nil {
			logger.Warn("Failed to marshal message content to JSON",
				zap.String("type", typeName),
				zap.Error(err))
			msgContent = []byte(fmt.Sprintf(`{"@type":"%s"}`, typeURL))
		}

		messages = append(messages, models.Message{
			Type:    typeName,
			Content: msgContent,
		})

		signer := p.extractSignerFromNuklaiMsg(typeURL, anyMsg.Value)
		if signer != "" {
			logger.Info("Found signer", zap.String("signer", signer))
			signers = appendIfUnique(signers, signer)
		} else {
			logger.Warn("Could not extract signer from message", zap.String("type_url", typeURL))
		}
	}

	return messages, signers, memo, nil
}

// Helper to find minimum of two values
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func dumpMessageBytes(msgBytes []byte) string {
	if len(msgBytes) < 100 {
		return hex.EncodeToString(msgBytes)
	}

	return fmt.Sprintf("%s...%s",
		hex.EncodeToString(msgBytes[:50]),
		hex.EncodeToString(msgBytes[len(msgBytes)-50:]))
}

func (p *TxParser) extractSignerFromBankMsg(msgBytes []byte) string {
	if len(msgBytes) < 5 {
		return ""
	}

	if msgBytes[0] != 0x0a {
		return ""
	}

	fieldLength := int(msgBytes[1])

	if fieldLength >= 128 {
		return ""
	}

	if len(msgBytes) < 2+fieldLength {
		return ""
	}

	addressBytes := msgBytes[2 : 2+fieldLength]

	return string(addressBytes)
}

func (p *TxParser) extractSignerFromBankMsgSend(msgBytes []byte) string {
	if len(msgBytes) < 3 {
		return ""
	}

	if msgBytes[0] != 0x0A {
		logger.Debug("First byte is not 0x0A as expected for MsgSend",
			zap.String("first_byte", fmt.Sprintf("%x", msgBytes[0])))
		return ""
	}

	addrLen := int(msgBytes[1])

	if len(msgBytes) < 2+addrLen {
		logger.Debug("Message too short for address length",
			zap.Int("needed_length", 2+addrLen),
			zap.Int("actual_length", len(msgBytes)))
		return ""
	}

	address := string(msgBytes[2 : 2+addrLen])
	logger.Info("Successfully extracted address from MsgSend", zap.String("address", address))

	return address
}

func (p *TxParser) extractSignerFromNuklaiMsg(typeURL string, msgBytes []byte) string {
	logger.Info("Extracting signer from message",
		zap.String("type_url", typeURL),
		zap.String("bytes_preview", dumpMessageBytes(msgBytes)))

	if strings.Contains(typeURL, "bank.v1beta1.MsgSend") {
		sender := p.extractSignerFromBankMsgSend(msgBytes)
		if sender != "" {
			return sender
		}
	}

	msgStr := string(msgBytes)

	nuklaiAddrs := p.findNuklaiAddresses(msgStr)
	if len(nuklaiAddrs) > 0 {
		logger.Info("Found potential Nuklai addresses", zap.Strings("addresses", nuklaiAddrs))
		return nuklaiAddrs[0]
	}

	if strings.Contains(typeURL, "dataset") {
		owner := p.extractFieldFromBytes(msgBytes, "owner")
		if owner != "" {
			logger.Info("Found owner field", zap.String("owner", owner))
			return owner
		}
		if strings.Contains(typeURL, "bank.MsgSend") {
			sender := p.extractSignerFromBankMsg(msgBytes)
			if sender != "" {
				logger.Info("Extracted sender from MsgSend", zap.String("sender", sender))
				return sender
			}

			fromAddr := p.extractFieldFromBytes(msgBytes, "from_address")
			if fromAddr != "" {
				logger.Info("Found from_address field", zap.String("from_address", fromAddr))
				return fromAddr
			}
		}
	} else if strings.Contains(typeURL, "staking.MsgDelegate") {
		delegator := p.extractFieldFromBytes(msgBytes, "delegator_address")
		if delegator != "" {
			logger.Info("Found delegator_address field", zap.String("delegator_address", delegator))
			return delegator
		}
	}

	logger.Warn("Could not extract signer from message bytes",
		zap.String("type_url", typeURL),
		zap.String("msg_hex", hex.EncodeToString(msgBytes)))

	return ""
}

func (p *TxParser) findNuklaiAddresses(input string) []string {
	var addresses []string

	offset := 0
	for {
		idx := strings.Index(input[offset:], "nuki")
		if idx == -1 {
			break
		}

		idx += offset

		addrStart := idx
		addrEnd := addrStart

		for i := addrStart; i < len(input) && i < addrStart+50; i++ {
			if isValidAddrChar(input[i]) {
				addrEnd = i
			} else {
				break
			}
		}

		potentialAddr := input[addrStart : addrEnd+1]
		if len(potentialAddr) >= 38 {
			addresses = append(addresses, potentialAddr)
		}

		offset = addrEnd + 1

		if offset >= len(input) {
			break
		}
	}

	return addresses
}

// Helper to check if a character is valid for a bech32 address
func isValidAddrChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
}

func (p *TxParser) extractFieldFromBytes(data []byte, fieldName string) string {
	dataStr := string(data)

	logger.Info("Searching for field in message",
		zap.String("field_name", fieldName),
		zap.Int("data_length", len(dataStr)))

	patterns := []string{
		fieldName + "\"",
		fieldName + ":",
		"\"" + fieldName + "\"",
		fieldName + "=",
	}

	for _, pattern := range patterns {
		idx := strings.Index(dataStr, pattern)
		if idx >= 0 {
			logger.Info("Found field pattern",
				zap.String("pattern", pattern),
				zap.Int("position", idx))

			startPos := idx + len(pattern)

			for startPos < len(dataStr) {
				c := dataStr[startPos]
				if c == ' ' || c == '"' || c == ':' || c == '=' || c == '>' {
					startPos++
				} else {
					break
				}
			}

			endPos := startPos
			for endPos < len(dataStr) {
				c := dataStr[endPos]
				if c == '"' || c == ',' || c == '}' || c == '<' || c == ' ' {
					break
				}
				endPos++
			}

			if startPos < endPos && startPos < len(dataStr) && endPos <= len(dataStr) {
				value := dataStr[startPos:endPos]
				logger.Info("Extracted field value",
					zap.String("field", fieldName),
					zap.String("value", value))
				return value
			}
		}
	}

	logger.Warn("Field not found in message", zap.String("field_name", fieldName))
	return ""
}

// Helper for parsing JSON transactions
func (p *TxParser) parseJSONTx(txJSON map[string]interface{}) (models.Messages, []string, string, error) {
	var messages models.Messages
	var signers []string
	var memo string

	logger.Info("Parsing JSON transaction", zap.Any("tx_json_keys", getMapKeys(txJSON)))

	if body, ok := txJSON["body"].(map[string]interface{}); ok {
		logger.Info("Found body in JSON transaction", zap.Any("body_keys", getMapKeys(body)))

		// Extract the memo
		if memoVal, ok := body["memo"].(string); ok {
			memo = memoVal
			logger.Info("Found memo", zap.String("memo", memo))
		}

		// Extract messages
		if msgs, ok := body["messages"].([]interface{}); ok {
			logger.Info("Found messages in body", zap.Int("messages_count", len(msgs)))

			for i, msg := range msgs {
				msgMap, ok := msg.(map[string]interface{})
				if !ok {
					logger.Warn("Message is not a map", zap.Int("message_index", i))
					continue
				}

				logger.Info("Processing JSON message",
					zap.Int("index", i),
					zap.Any("message_keys", getMapKeys(msgMap)))

				// Extract type and content
				typeURL, _ := msgMap["@type"].(string)
				if typeURL == "" {
					logger.Warn("Message has no @type field", zap.Int("message_index", i))
					continue
				}

				// Remove the "/" prefix if present
				typeName := strings.TrimPrefix(typeURL, "/")
				logger.Info("Message type", zap.String("type", typeName))

				// Convert the message to JSON
				msgJSON, err := json.Marshal(msgMap)
				if err != nil {
					logger.Warn("Failed to marshal message to JSON",
						zap.Int("message_index", i),
						zap.Error(err))
					continue
				}

				messages = append(messages, models.Message{
					Type:    typeName,
					Content: msgJSON,
				})

				if owner, ok := msgMap["owner"].(string); ok {
					logger.Info("Found owner field", zap.String("owner", owner))
					signers = appendIfUnique(signers, owner)
				}
				if from, ok := msgMap["from_address"].(string); ok {
					logger.Info("Found from_address field", zap.String("from_address", from))
					signers = appendIfUnique(signers, from)
				}
				if from, ok := msgMap["sender"].(string); ok {
					logger.Info("Found sender field", zap.String("sender", from))
					signers = appendIfUnique(signers, from)
				}
				if delegator, ok := msgMap["delegator_address"].(string); ok {
					logger.Info("Found delegator_address field", zap.String("delegator_address", delegator))
					signers = appendIfUnique(signers, delegator)
				}
				if voter, ok := msgMap["voter"].(string); ok {
					logger.Info("Found voter field", zap.String("voter", voter))
					signers = appendIfUnique(signers, voter)
				}
				if proposer, ok := msgMap["proposer"].(string); ok {
					logger.Info("Found proposer field", zap.String("proposer", proposer))
					signers = appendIfUnique(signers, proposer)
				}
			}
		}
	}

	// Check if we have auth_info for more signer info
	if authInfo, ok := txJSON["auth_info"].(map[string]interface{}); ok {
		logger.Info("Found auth_info in JSON transaction", zap.Any("auth_info_keys", getMapKeys(authInfo)))

		if signerInfos, ok := authInfo["signer_infos"].([]interface{}); ok {
			logger.Info("Found signer_infos in auth_info", zap.Int("signer_infos_count", len(signerInfos)))

			for i, signerInfo := range signerInfos {
				siMap, ok := signerInfo.(map[string]interface{})
				if !ok {
					logger.Warn("Signer info is not a map", zap.Int("signer_info_index", i))
					continue
				}

				logger.Info("Processing signer info",
					zap.Int("index", i),
					zap.Any("signer_info_keys", getMapKeys(siMap)))

				if pubKey, ok := siMap["public_key"].(map[string]interface{}); ok {
					logger.Info("Found public_key in signer info", zap.Any("public_key_keys", getMapKeys(pubKey)))

					if key, ok := pubKey["key"].(string); ok {
						logger.Info("Found key in public_key", zap.String("key_preview", key[:min(20, len(key))]))
					}
				}
			}
		}
	}

	logger.Info("Parsed JSON transaction results",
		zap.Int("messages_count", len(messages)),
		zap.Strings("signers", signers),
		zap.String("memo", memo))

	return messages, signers, memo, nil
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
