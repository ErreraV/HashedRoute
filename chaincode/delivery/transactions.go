package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

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
