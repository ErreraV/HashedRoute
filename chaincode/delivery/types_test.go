package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShipmentJSON_RoundTrip(t *testing.T) {
	original := Shipment{
		ID:          "SHP-001",
		Origin:      "New York",
		Destination: "Los Angeles",
		Customer:    "Acme Corp",
		Carrier:     "FastShip",
		CreatedAt:   "2025-01-15T10:30:00Z",
		Status:      StatusCreated,
		StatusHistory: []StatusChange{
			{Status: StatusPickedUp, Notes: "picked up", Timestamp: "2025-01-15T11:00:00Z"},
			{Status: StatusInTransit, Notes: "en route", Timestamp: "2025-01-16T08:00:00Z"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored Shipment
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Origin, restored.Origin)
	assert.Equal(t, original.Destination, restored.Destination)
	assert.Equal(t, original.Customer, restored.Customer)
	assert.Equal(t, original.Carrier, restored.Carrier)
	assert.Equal(t, original.CreatedAt, restored.CreatedAt)
	assert.Equal(t, original.Status, restored.Status)
	assert.Equal(t, len(original.StatusHistory), len(restored.StatusHistory))
	assert.Equal(t, original.StatusHistory[0].Status, restored.StatusHistory[0].Status)
	assert.Equal(t, original.StatusHistory[0].Notes, restored.StatusHistory[0].Notes)
	assert.Equal(t, original.StatusHistory[0].Timestamp, restored.StatusHistory[0].Timestamp)
}

func TestShipmentJSON_EmptyHistory(t *testing.T) {
	sh := Shipment{
		ID:            "SHP-002",
		Origin:        "Boston",
		Destination:   "Chicago",
		Customer:      "Beta Inc",
		Carrier:       "QuickDel",
		CreatedAt:     "2025-01-15T10:30:00Z",
		Status:        StatusCreated,
		StatusHistory: []StatusChange{},
	}

	data, err := json.Marshal(sh)
	require.NoError(t, err)

	// Should marshal as empty array, not null
	assert.Contains(t, string(data), `"statusHistory":[]`)
}

func TestShipmentJSON_ZeroValueHistory(t *testing.T) {
	// StatusHistory is declared as []StatusChange but never assigned
	sh := Shipment{
		ID:          "SHP-003",
		Origin:      "Miami",
		Destination: "Seattle",
		CreatedAt:   "2025-01-15T10:30:00Z",
		Status:      StatusCreated,
	}

	data, err := json.Marshal(sh)
	require.NoError(t, err)

	// nil slice marshals as null, not []
	assert.Contains(t, string(data), `"statusHistory":null`)
}

func TestStatusChangeJSON_RoundTrip(t *testing.T) {
	original := StatusChange{
		Status:    StatusDelivered,
		Notes:     "signed by John",
		Timestamp: "2025-01-17T14:00:00Z",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var restored StatusChange
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, original.Status, restored.Status)
	assert.Equal(t, original.Notes, restored.Notes)
	assert.Equal(t, original.Timestamp, restored.Timestamp)
}

func TestStatusChangeJSON_EmptyNotes(t *testing.T) {
	sc := StatusChange{
		Status:    StatusCreated,
		Timestamp: "2025-01-15T10:30:00Z",
	}

	data, err := json.Marshal(sc)
	require.NoError(t, err)

	var restored StatusChange
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, "", restored.Notes)
}
