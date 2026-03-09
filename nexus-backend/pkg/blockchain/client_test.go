package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_EmptyRPCURL(t *testing.T) {
	client, err := NewClient("", "0x0000000000000000000000000000000000000000")
	require.NoError(t, err)
	assert.Nil(t, client, "empty RPC URL should return nil client without error")
}

func TestShipmentTrackerABI_IsValid(t *testing.T) {

	assert.Contains(t, ShipmentTrackerABI, "recordEvent")
	assert.Contains(t, ShipmentTrackerABI, "getEvents")
}
