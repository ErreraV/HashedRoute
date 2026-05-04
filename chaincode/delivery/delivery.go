package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

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
	ID             string         `json:"id"`
	Origin         string         `json:"origin"`
	Destination    string         `json:"destination"`
	Customer       string         `json:"customer"`
	Carrier        string         `json:"carrier"`
	CreatedAt      string         `json:"createdAt"`
	Status         string         `json:"status"`
	Notes          string         `json:"notes,omitempty"`
	StatusHistory  []StatusChange `json:"statusHistory"`
}

func shipmentKey(id string) string {
	return "SHM:" + id
}

func normalizeID(id string) string {
	return strings.TrimSpace(id)
}

// CreateShipment stores a new shipment with status CREATED.
func (c *DeliveryContract) CreateShipment(ctx contractapi.TransactionContextInterface, id, origin, destination, customer, carrier string) error {
	id = normalizeID(id)
	if id == "" {
		return fmt.Errorf("id is required")
	}
	exists, err := c.shipmentExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("shipment %s already exists", id)
	}
	ts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	created := time.Unix(ts.GetSeconds(), int64(ts.GetNanos())).UTC().Format(time.RFC3339)
	sh := Shipment{
		ID:            id,
		Origin:        strings.TrimSpace(origin),
		Destination:   strings.TrimSpace(destination),
		Customer:      strings.TrimSpace(customer),
		Carrier:       strings.TrimSpace(carrier),
		CreatedAt:     created,
		Status:        StatusCreated,
		StatusHistory: []StatusChange{},
	}
	return c.putShipment(ctx, &sh)
}

// UpdateStatus moves a shipment forward in the state machine and appends history.
func (c *DeliveryContract) UpdateStatus(ctx contractapi.TransactionContextInterface, id, nextStatus, notes string) error {
	id = normalizeID(id)
	if id == "" {
		return fmt.Errorf("id is required")
	}
	nextStatus = strings.TrimSpace(strings.ToUpper(nextStatus))
	if err := validateStatusValue(nextStatus); err != nil {
		return err
	}
	sh, err := c.readShipment(ctx, id)
	if err != nil {
		return err
	}
	if err := assertValidTransition(sh.Status, nextStatus); err != nil {
		return err
	}
	ts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	when := time.Unix(ts.GetSeconds(), int64(ts.GetNanos())).UTC().Format(time.RFC3339)
	entry := StatusChange{Status: nextStatus, Notes: strings.TrimSpace(notes), Timestamp: when}
	sh.StatusHistory = append(sh.StatusHistory, entry)
	sh.Status = nextStatus
	sh.Notes = strings.TrimSpace(notes)
	return c.putShipment(ctx, sh)
}

// GetShipment returns a shipment by id.
func (c *DeliveryContract) GetShipment(ctx contractapi.TransactionContextInterface, id string) (*Shipment, error) {
	id = normalizeID(id)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	return c.readShipment(ctx, id)
}

// ListShipments returns all shipments under the contract key prefix.
func (c *DeliveryContract) ListShipments(ctx contractapi.TransactionContextInterface) ([]*Shipment, error) {
	iter, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var out []*Shipment
	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(kv.Key, "SHM:") {
			continue
		}
		var sh Shipment
		if err := json.Unmarshal(kv.Value, &sh); err != nil {
			return nil, err
		}
		out = append(out, &sh)
	}
	return out, nil
}

func (c *DeliveryContract) shipmentExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	data, err := ctx.GetStub().GetState(shipmentKey(id))
	if err != nil {
		return false, fmt.Errorf("failed to read world state: %w", err)
	}
	return len(data) > 0, nil
}

func (c *DeliveryContract) readShipment(ctx contractapi.TransactionContextInterface, id string) (*Shipment, error) {
	data, err := ctx.GetStub().GetState(shipmentKey(id))
	if err != nil {
		return nil, fmt.Errorf("failed to read world state: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("shipment %s does not exist", id)
	}
	var sh Shipment
	if err := json.Unmarshal(data, &sh); err != nil {
		return nil, err
	}
	return &sh, nil
}

func (c *DeliveryContract) putShipment(ctx contractapi.TransactionContextInterface, sh *Shipment) error {
	payload, err := json.Marshal(sh)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(shipmentKey(sh.ID), payload)
}

func validateStatusValue(s string) error {
	switch s {
	case StatusCreated, StatusPickedUp, StatusInTransit, StatusDelivered:
		return nil
	default:
		return fmt.Errorf("invalid status %q", s)
	}
}

func assertValidTransition(from, to string) error {
	if from == to {
		return fmt.Errorf("status is already %s", from)
	}
	switch from {
	case StatusCreated:
		if to != StatusPickedUp {
			return fmt.Errorf("cannot go from %s to %s", from, to)
		}
	case StatusPickedUp:
		if to != StatusInTransit {
			return fmt.Errorf("cannot go from %s to %s", from, to)
		}
	case StatusInTransit:
		if to != StatusDelivered {
			return fmt.Errorf("cannot go from %s to %s", from, to)
		}
	case StatusDelivered:
		return fmt.Errorf("shipment is already delivered")
	default:
		return fmt.Errorf("unknown current status %s", from)
	}
	return nil
}
