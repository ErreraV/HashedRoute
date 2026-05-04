package main

import "github.com/hyperledger/fabric-contract-api-go/v2/contractapi"

const (
	StatusCreated   = "CREATED"
	StatusPickedUp  = "PICKED_UP"
	StatusInTransit = "IN_TRANSIT"
	StatusDelivered = "DELIVERED"
)

// DeliveryContract tracks shipments on the ledger.
type DeliveryContract struct {
	contractapi.Contract
}

// StatusChange records one transition in the shipment lifecycle.
type StatusChange struct {
	Status    string `json:"status"`
	Notes     string `json:"notes,omitempty"`
	Timestamp string `json:"timestamp"`
}

// Shipment is the on-chain record for a delivery.
type Shipment struct {
	ID            string         `json:"id"`
	Origin        string         `json:"origin"`
	Destination   string         `json:"destination"`
	Customer      string         `json:"customer"`
	Carrier       string         `json:"carrier"`
	CreatedAt     string         `json:"createdAt"`
	Status        string         `json:"status"`
	Notes         string         `json:"notes,omitempty"`
	StatusHistory []StatusChange `json:"statusHistory"`
}
