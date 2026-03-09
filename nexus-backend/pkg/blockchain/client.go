package blockchain

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

const ShipmentTrackerABI = `[
  {
    "inputs": [
      {"internalType": "string",  "name": "shipmentId", "type": "string"},
      {"internalType": "string",  "name": "status",     "type": "string"},
      {"internalType": "string",  "name": "notes",      "type": "string"}
    ],
    "name": "recordEvent",
    "outputs": [],
    "stateMutability": "nonpayable",
    "type": "function"
  },
  {
    "inputs": [
      {"internalType": "string", "name": "shipmentId", "type": "string"}
    ],
    "name": "getEvents",
    "outputs": [
      {
        "components": [
          {"internalType": "string",  "name": "status",    "type": "string"},
          {"internalType": "string",  "name": "notes",     "type": "string"},
          {"internalType": "uint256", "name": "timestamp", "type": "uint256"}
        ],
        "internalType": "struct ShipmentTracker.Event[]",
        "name": "",
        "type": "tuple[]"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]`

type Client struct {
	eth             *ethclient.Client
	contractAddress common.Address
	parsedABI       abi.ABI
}

func NewClient(rpcURL, contractAddress string) (*Client, error) {
	if rpcURL == "" {
		log.Warn().Msg("ETHEREUM_RPC_URL not set - blockchain features disabled")
		return nil, nil
	}

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("ethclient dial: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(ShipmentTrackerABI))
	if err != nil {
		return nil, fmt.Errorf("parse ABI: %w", err)
	}

	addr := common.HexToAddress(contractAddress)

	return &Client{
		eth:             ethClient,
		contractAddress: addr,
		parsedABI:       parsedABI,
	}, nil
}

func (c *Client) RecordEvent(ctx context.Context, shipmentID, status, notes string) (string, error) {
	data, err := c.parsedABI.Pack("recordEvent", shipmentID, status, notes)
	if err != nil {
		return "", fmt.Errorf("ABI pack: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &c.contractAddress,
		Data: data,
	}

	gas, err := c.eth.EstimateGas(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("gas estimation: %w", err)
	}
	log.Debug().Uint64("gas", gas).Msg("blockchain gas estimate")

	return "0x_tx_hash_placeholder", nil
}

func (c *Client) GetEvents(ctx context.Context, shipmentID string) ([]map[string]interface{}, error) {
	data, err := c.parsedABI.Pack("getEvents", shipmentID)
	if err != nil {
		return nil, fmt.Errorf("ABI pack: %w", err)
	}

	result, err := c.eth.CallContract(ctx, ethereum.CallMsg{
		To:   &c.contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("contract call: %w", err)
	}

	unpacked, err := c.parsedABI.Unpack("getEvents", result)
	if err != nil {
		return nil, fmt.Errorf("ABI unpack: %w", err)
	}
	_ = unpacked
	return nil, nil
}
