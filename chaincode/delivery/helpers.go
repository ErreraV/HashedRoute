package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

func shipmentKey(id string) string {
	return "SHM:" + id
}

func normalizeID(id string) string {
	return strings.TrimSpace(id)
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
