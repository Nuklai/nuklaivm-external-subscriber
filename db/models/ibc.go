package models

import (
	"time"
)

type IBCChannel struct {
	ChannelID             string `db:"channel_id" json:"channel_id"`
	PortID                string `db:"port_id" json:"port_id"`
	ConnectionID          string `db:"connection_id" json:"connection_id"`
	CounterpartyChannelID string `db:"counterparty_channel_id" json:"counterparty_channel_id"`
	CounterpartyPortID    string `db:"counterparty_port_id" json:"counterparty_port_id"`
	State                 string `db:"state" json:"state"`
	Ordering              string `db:"ordering" json:"ordering"`
	Version               string `db:"version" json:"version"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type IBCConnection struct {
	ConnectionID             string `db:"connection_id" json:"connection_id"`
	ClientID                 string `db:"client_id" json:"client_id"`
	CounterpartyConnectionID string `db:"counterparty_connection_id" json:"counterparty_connection_id"`
	CounterpartyClientID     string `db:"counterparty_client_id" json:"counterparty_client_id"`
	State                    string `db:"state" json:"state"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type IBCClient struct {
	ClientID   string `db:"client_id" json:"client_id"`
	ClientType string `db:"client_type" json:"client_type"`
	ChainID    string `db:"chain_id" json:"chain_id"`
	LastHeight int64  `db:"last_height" json:"last_height"`
	State      string `db:"state" json:"state"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type IBCPacket struct {
	ID                  string    `db:"id" json:"id"`
	SourceChannelID     string    `db:"source_channel_id" json:"source_channel_id"`
	SourcePortID        string    `db:"source_port_id" json:"source_port_id"`
	DestChannelID       string    `db:"dest_channel_id" json:"dest_channel_id"`
	DestPortID          string    `db:"dest_port_id" json:"dest_port_id"`
	SequenceNumber      int64     `db:"sequence_number" json:"sequence_number"`
	PacketType          string    `db:"packet_type" json:"packet_type"`
	PacketStatus        string    `db:"packet_status" json:"packet_status"`
	SentBlockHeight     int64     `db:"sent_block_height" json:"sent_block_height"`
	SentBlockTime       time.Time `db:"sent_block_time" json:"sent_block_time"`
	ReceivedBlockHeight int64     `db:"received_block_height" json:"received_block_height,omitempty"`
	ReceivedBlockTime   time.Time `db:"received_block_time" json:"received_block_time,omitempty"`
	AckBlockHeight      int64     `db:"ack_block_height" json:"ack_block_height,omitempty"`
	AckBlockTime        time.Time `db:"ack_block_time" json:"ack_block_time,omitempty"`
	TimeoutBlockHeight  int64     `db:"timeout_block_height" json:"timeout_block_height,omitempty"`
	TimeoutBlockTime    time.Time `db:"timeout_block_time" json:"timeout_block_time,omitempty"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
