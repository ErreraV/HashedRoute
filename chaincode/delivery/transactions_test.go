package main

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-protos-go-apiv2/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// ---------------------------------------------------------------------------
// mock stub + iterator — implements only the methods called by DeliveryContract
// ---------------------------------------------------------------------------

type testStub struct {
	mu          sync.RWMutex
	state       map[string][]byte
	txTimestamp *timestamppb.Timestamp
	txID        string
	channelID   string
}

func newTestStub() *testStub {
	return &testStub{
		state:       make(map[string][]byte),
		txTimestamp: timestamppb.Now(),
		txID:        "test-txid",
		channelID:   "test-channel",
	}
}

func (s *testStub) setState(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[key] = value
}

// ---- methods called by DeliveryContract ----

func (s *testStub) GetState(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state[key], nil
}

func (s *testStub) PutState(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[key] = value
	return nil
}

func (s *testStub) GetTxTimestamp() (*timestamppb.Timestamp, error) {
	return s.txTimestamp, nil
}

func (s *testStub) GetStateByRange(startKey, endKey string) (shim.StateQueryIteratorInterface, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var entries []kvEntry
	for k, v := range s.state {
		if k >= startKey && (endKey == "" || k < endKey) {
			vc := make([]byte, len(v))
			copy(vc, v)
			entries = append(entries, kvEntry{k, vc})
		}
	}
	return &testIterator{entries: entries, index: -1}, nil
}

func (s *testStub) GetTxID() string                       { return s.txID }
func (s *testStub) GetChannelID() string                  { return s.channelID }
func (s *testStub) GetArgs() [][]byte                     { return nil }
func (s *testStub) GetStringArgs() []string               { return nil }
func (s *testStub) GetFunctionAndParameters() (string, []string) { return "", nil }
func (s *testStub) GetArgsSlice() ([]byte, error)         { return nil, nil }

// ---- remaining methods (unused by our chaincode) ----

func (s *testStub) InvokeChaincode(string, [][]byte, string) *peer.Response {
	panic("testStub.InvokeChaincode not implemented")
}
func (s *testStub) DelState(string) error                                          { panic("not implemented") }
func (s *testStub) SetStateValidationParameter(string, []byte) error               { panic("not implemented") }
func (s *testStub) GetStateValidationParameter(string) ([]byte, error)              { panic("not implemented") }
func (s *testStub) GetStateByRangeWithPagination(string, string, int32, string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	panic("not implemented")
}
func (s *testStub) GetStateByPartialCompositeKey(string, []string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented")
}
func (s *testStub) GetStateByPartialCompositeKeyWithPagination(string, []string, int32, string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	panic("not implemented")
}
func (s *testStub) CreateCompositeKey(string, []string) (string, error)            { panic("not implemented") }
func (s *testStub) SplitCompositeKey(string) (string, []string, error)              { panic("not implemented") }
func (s *testStub) GetQueryResult(string) (shim.StateQueryIteratorInterface, error) { panic("not implemented") }
func (s *testStub) GetQueryResultWithPagination(string, int32, string) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	panic("not implemented")
}
func (s *testStub) GetHistoryForKey(string) (shim.HistoryQueryIteratorInterface, error) { panic("not implemented") }
func (s *testStub) GetPrivateData(string, string) ([]byte, error)                       { panic("not implemented") }
func (s *testStub) GetPrivateDataHash(string, string) ([]byte, error)                    { panic("not implemented") }
func (s *testStub) PutPrivateData(string, string, []byte) error                          { panic("not implemented") }
func (s *testStub) DelPrivateData(string, string) error                                  { panic("not implemented") }
func (s *testStub) PurgePrivateData(string, string) error                                { panic("not implemented") }
func (s *testStub) SetPrivateDataValidationParameter(string, string, []byte) error      { panic("not implemented") }
func (s *testStub) GetPrivateDataValidationParameter(string, string) ([]byte, error)     { panic("not implemented") }
func (s *testStub) GetPrivateDataByRange(string, string, string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented")
}
func (s *testStub) GetPrivateDataByPartialCompositeKey(string, string, []string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented")
}
func (s *testStub) GetPrivateDataQueryResult(string, string) (shim.StateQueryIteratorInterface, error) {
	panic("not implemented")
}
func (s *testStub) GetCreator() ([]byte, error)                   { panic("not implemented") }
func (s *testStub) GetTransient() (map[string][]byte, error)      { panic("not implemented") }
func (s *testStub) GetBinding() ([]byte, error)                    { panic("not implemented") }
func (s *testStub) GetDecorations() map[string][]byte              { panic("not implemented") }
func (s *testStub) GetSignedProposal() (*peer.SignedProposal, error) { panic("not implemented") }
func (s *testStub) SetEvent(string, []byte) error                  { panic("not implemented") }

// ---- iterator ----

type kvEntry struct {
	key   string
	value []byte
}

type testIterator struct {
	entries []kvEntry
	index   int
	closed  bool
}

func (it *testIterator) HasNext() bool {
	if it.closed {
		panic("iterator closed")
	}
	return it.index+1 < len(it.entries)
}

func (it *testIterator) Next() (*queryresult.KV, error) {
	if it.closed {
		panic("iterator closed")
	}
	it.index++
	e := it.entries[it.index]
	return &queryresult.KV{Key: e.key, Value: e.value, Namespace: ""}, nil
}

func (it *testIterator) Close() error {
	it.closed = true
	return nil
}

// ---- helper: create a transaction context with our mock stub ----

func newTxContext(stub *testStub) contractapi.TransactionContextInterface {
	ctx := &contractapi.TransactionContext{}
	ctx.SetStub(stub)
	return ctx
}

// ---- helper: store a serialized shipment in the stub's state ----

func storeShipment(stub *testStub, sh *Shipment) {
	payload, _ := json.Marshal(sh)
	stub.setState(shipmentKey(sh.ID), payload)
}

// ---- helper: extract shipment from stub state for verification ----

func loadShipment(stub *testStub, id string) *Shipment {
	s, _ := stub.GetState(shipmentKey(id))
	if len(s) == 0 {
		return nil
	}
	var sh Shipment
	json.Unmarshal(s, &sh)
	return &sh
}

// ===================================================================
// Tests
// ===================================================================

func TestCreateShipment_Success(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.CreateShipment(ctx, "SHP-001", "NYC", "LAX", "Acme", "FastShip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sh := loadShipment(stub, "SHP-001")
	if sh == nil {
		t.Fatal("shipment not stored")
	}
	if sh.ID != "SHP-001" {
		t.Errorf("id = %q, want SHP-001", sh.ID)
	}
	if sh.Origin != "NYC" {
		t.Errorf("origin = %q, want NYC", sh.Origin)
	}
	if sh.Destination != "LAX" {
		t.Errorf("destination = %q, want LAX", sh.Destination)
	}
	if sh.Customer != "Acme" {
		t.Errorf("customer = %q, want Acme", sh.Customer)
	}
	if sh.Carrier != "FastShip" {
		t.Errorf("carrier = %q, want FastShip", sh.Carrier)
	}
	if sh.Status != StatusCreated {
		t.Errorf("status = %q, want CREATED", sh.Status)
	}
	if sh.CreatedAt == "" {
		t.Error("CreatedAt is empty")
	}
	if sh.StatusHistory == nil || len(sh.StatusHistory) != 0 {
		t.Error("StatusHistory should be empty slice, not nil")
	}
	if sh.Notes != "" {
		t.Errorf("notes = %q, want empty", sh.Notes)
	}
}

func TestCreateShipment_EmptyID(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.CreateShipment(ctx, "", "NYC", "LAX", "Acme", "FastShip")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
	if !strings.Contains(err.Error(), "id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateShipment_WhitespaceOnlyID(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.CreateShipment(ctx, "   \t  ", "NYC", "LAX", "Acme", "FastShip")
	if err == nil {
		t.Fatal("expected error for whitespace-only id")
	}
	if !strings.Contains(err.Error(), "id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateShipment_TrimsWhitespaceFromID(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.CreateShipment(ctx, "  SHP-002  ", "NYC", "LAX", "Acme", "FastShip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Key should use trimmed ID
	_, err = stub.GetState(shipmentKey("  SHP-002  "))
	if err != nil {
		t.Fatal(err)
	}

	sh := loadShipment(stub, "SHP-002")
	if sh == nil {
		t.Fatal("shipment not found under trimmed key")
	}
	if sh.ID != "SHP-002" {
		t.Errorf("stored id = %q, want SHP-002", sh.ID)
	}
}

func TestCreateShipment_TrimsFieldWhitespace(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.CreateShipment(ctx, "SHP-003", "  NYC  ", "  LAX  ", "  Acme  ", "  FastShip  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sh := loadShipment(stub, "SHP-003")
	if sh.Origin != "NYC" {
		t.Errorf("origin = %q, want NYC", sh.Origin)
	}
	if sh.Destination != "LAX" {
		t.Errorf("destination = %q, want LAX", sh.Destination)
	}
}

func TestCreateShipment_DuplicateID(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.CreateShipment(ctx, "SHP-001", "NYC", "LAX", "Acme", "FastShip")
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	err = c.CreateShipment(ctx, "SHP-001", "NYC", "LAX", "Acme", "FastShip")
	if err == nil {
		t.Fatal("expected error for duplicate id")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateStatus tests
// ---------------------------------------------------------------------------

func newCreatedShipment(stub *testStub, id string) {
	storeShipment(stub, &Shipment{
		ID:            id,
		Origin:        "NYC",
		Destination:   "LAX",
		Customer:      "Acme",
		Carrier:       "FastShip",
		CreatedAt:     "2025-01-15T10:30:00Z",
		Status:        StatusCreated,
		StatusHistory: []StatusChange{},
	})
}

func TestUpdateStatus_ValidTransitions(t *testing.T) {
	stub := newTestStub()
	var c DeliveryContract

	// Create shipment
	newCreatedShipment(stub, "SHP-T1")
	ctx := newTxContext(stub)

	// CREATED → PICKED_UP
	err := c.UpdateStatus(ctx, "SHP-T1", StatusPickedUp, "at dock")
	if err != nil {
		t.Fatalf("PICKED_UP: %v", err)
	}
	sh := loadShipment(stub, "SHP-T1")
	if sh.Status != StatusPickedUp {
		t.Errorf("status = %q, want PICKED_UP", sh.Status)
	}
	if len(sh.StatusHistory) != 1 {
		t.Errorf("history len = %d, want 1", len(sh.StatusHistory))
	}
	if sh.StatusHistory[0].Status != StatusPickedUp {
		t.Errorf("history[0].status = %q", sh.StatusHistory[0].Status)
	}

	// PICKED_UP → IN_TRANSIT
	err = c.UpdateStatus(ctx, "SHP-T1", StatusInTransit, "on the road")
	if err != nil {
		t.Fatalf("IN_TRANSIT: %v", err)
	}
	sh = loadShipment(stub, "SHP-T1")
	if sh.Status != StatusInTransit {
		t.Errorf("status = %q, want IN_TRANSIT", sh.Status)
	}
	if len(sh.StatusHistory) != 2 {
		t.Errorf("history len = %d, want 2", len(sh.StatusHistory))
	}

	// IN_TRANSIT → DELIVERED
	err = c.UpdateStatus(ctx, "SHP-T1", StatusDelivered, "done")
	if err != nil {
		t.Fatalf("DELIVERED: %v", err)
	}
	sh = loadShipment(stub, "SHP-T1")
	if sh.Status != StatusDelivered {
		t.Errorf("status = %q, want DELIVERED", sh.Status)
	}
	if len(sh.StatusHistory) != 3 {
		t.Errorf("history len = %d, want 3", len(sh.StatusHistory))
	}
}

func TestUpdateStatus_NonExistentID(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.UpdateStatus(ctx, "NO-SUCH", StatusPickedUp, "")
	if err == nil {
		t.Fatal("expected error for non-existent id")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_InvalidStatusValue(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-001")
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.UpdateStatus(ctx, "SHP-001", "BOGUS", "")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_SameStatus(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-001")
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.UpdateStatus(ctx, "SHP-001", StatusCreated, "")
	if err == nil {
		t.Fatal("expected error for same status")
	}
	if !strings.Contains(err.Error(), "already") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_FromDelivered(t *testing.T) {
	stub := newTestStub()
	storeShipment(stub, &Shipment{
		ID:            "SHP-DONE",
		Origin:        "NYC",
		Destination:   "LAX",
		Customer:      "Acme",
		Carrier:       "FastShip",
		CreatedAt:     "2025-01-15T10:30:00Z",
		Status:        StatusDelivered,
		StatusHistory: []StatusChange{{Status: StatusDelivered, Timestamp: "2025-01-17T14:00:00Z"}},
	})
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.UpdateStatus(ctx, "SHP-DONE", StatusPickedUp, "")
	if err == nil {
		t.Fatal("expected error from delivered")
	}
	if !strings.Contains(err.Error(), "already delivered") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_LowercaseNormalized(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-LC")
	ctx := newTxContext(stub)

	var c DeliveryContract
	// passing lowercase should still work because UpdateStatus uppercases
	err := c.UpdateStatus(ctx, "SHP-LC", "picked_up", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sh := loadShipment(stub, "SHP-LC")
	if sh.Status != StatusPickedUp {
		t.Errorf("status = %q, want PICKED_UP", sh.Status)
	}
	if sh.StatusHistory[0].Status != StatusPickedUp {
		t.Errorf("history status = %q", sh.StatusHistory[0].Status)
	}
}

func TestUpdateStatus_HistoryTimestamped(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-HST")
	ctx := newTxContext(stub)

	var c DeliveryContract
	_ = c.UpdateStatus(ctx, "SHP-HST", StatusPickedUp, "first")
	_ = c.UpdateStatus(ctx, "SHP-HST", StatusInTransit, "second")

	sh := loadShipment(stub, "SHP-HST")
	if len(sh.StatusHistory) != 2 {
		t.Fatalf("history len = %d, want 2", len(sh.StatusHistory))
	}
	for i, entry := range sh.StatusHistory {
		if entry.Timestamp == "" {
			t.Errorf("history[%d] timestamp is empty", i)
		}
	}
}

func TestUpdateStatus_UpdatesNotes(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-NTS")
	ctx := newTxContext(stub)

	var c DeliveryContract
	_ = c.UpdateStatus(ctx, "SHP-NTS", StatusPickedUp, "  dock A  ")

	sh := loadShipment(stub, "SHP-NTS")
	if sh.Notes != "dock A" {
		t.Errorf("notes = %q, want 'dock A'", sh.Notes)
	}
	if sh.StatusHistory[0].Notes != "dock A" {
		t.Errorf("history notes = %q", sh.StatusHistory[0].Notes)
	}
}

func TestUpdateStatus_EmptyNotesOK(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-NE")
	ctx := newTxContext(stub)

	var c DeliveryContract
	err := c.UpdateStatus(ctx, "SHP-NE", StatusPickedUp, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sh := loadShipment(stub, "SHP-NE")
	if sh.Notes != "" {
		t.Errorf("notes = %q, want empty", sh.Notes)
	}
}

// ---------------------------------------------------------------------------
// GetShipment tests
// ---------------------------------------------------------------------------

func TestGetShipment_Found(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-GET")
	ctx := newTxContext(stub)

	var c DeliveryContract
	sh, err := c.GetShipment(ctx, "SHP-GET")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sh.ID != "SHP-GET" {
		t.Errorf("id = %q", sh.ID)
	}
	if sh.Status != StatusCreated {
		t.Errorf("status = %q", sh.Status)
	}
}

func TestGetShipment_NotFound(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	_, err := c.GetShipment(ctx, "NO-SUCH")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetShipment_EmptyID(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	_, err := c.GetShipment(ctx, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetShipment_TrimsID(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-TRIM")
	ctx := newTxContext(stub)

	var c DeliveryContract
	sh, err := c.GetShipment(ctx, "  SHP-TRIM  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sh.ID != "SHP-TRIM" {
		t.Errorf("id = %q, want SHP-TRIM", sh.ID)
	}
}

// ---------------------------------------------------------------------------
// ListShipments tests
// ---------------------------------------------------------------------------

func TestListShipments_Empty(t *testing.T) {
	stub := newTestStub()
	ctx := newTxContext(stub)

	var c DeliveryContract
	shipments, err := c.ListShipments(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shipments) != 0 {
		t.Errorf("expected 0 shipments, got %d", len(shipments))
	}
}

func TestListShipments_WithShipments(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-A")
	newCreatedShipment(stub, "SHP-B")
	newCreatedShipment(stub, "SHP-C")
	ctx := newTxContext(stub)

	var c DeliveryContract
	shipments, err := c.ListShipments(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shipments) != 3 {
		t.Errorf("expected 3 shipments, got %d", len(shipments))
	}

	ids := make(map[string]bool)
	for _, s := range shipments {
		ids[s.ID] = true
	}
	for _, want := range []string{"SHP-A", "SHP-B", "SHP-C"} {
		if !ids[want] {
			t.Errorf("missing shipment %q", want)
		}
	}
}

func TestListShipments_FiltersNonSHMKeys(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-F1")
	// Put a non-SHM key directly in state
	stub.setState("OTHER:key", []byte(`{"not":"a shipment"}`))
	ctx := newTxContext(stub)

	var c DeliveryContract
	shipments, err := c.ListShipments(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shipments) != 1 {
		t.Errorf("expected 1 shipment, got %d", len(shipments))
	}
	if shipments[0].ID != "SHP-F1" {
		t.Errorf("unexpected id: %q", shipments[0].ID)
	}
}

func TestListShipments_UnmarshalErrorSkipsAll(t *testing.T) {
	stub := newTestStub()
	newCreatedShipment(stub, "SHP-GOOD")
	// corrupt value under an SHM: prefix
	stub.setState("SHM:SHP-BAD", []byte(`{not-json`))
	ctx := newTxContext(stub)

	var c DeliveryContract
	_, err := c.ListShipments(ctx)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

// ---------------------------------------------------------------------------
// Full lifecycle
// ---------------------------------------------------------------------------

func TestFullLifecycle(t *testing.T) {
	stub := newTestStub()
	var c DeliveryContract

	// Create
	ctx := newTxContext(stub)
	err := c.CreateShipment(ctx, "LIFE-CYCLE", "NYC", "SFO", "Acme", "FastShip")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Advance through all stages
	stages := []struct {
		status string
		notes  string
	}{
		{StatusPickedUp, "picked up at warehouse"},
		{StatusInTransit, "crossing state line"},
		{StatusDelivered, "signed by recipient"},
	}
	for _, stage := range stages {
		err = c.UpdateStatus(ctx, "LIFE-CYCLE", stage.status, stage.notes)
		if err != nil {
			t.Fatalf("update to %s: %v", stage.status, err)
		}
	}

	// Verify final state
	sh := loadShipment(stub, "LIFE-CYCLE")
	if sh.Status != StatusDelivered {
		t.Errorf("final status = %q, want DELIVERED", sh.Status)
	}
	if len(sh.StatusHistory) != 3 {
		t.Errorf("history len = %d, want 3", len(sh.StatusHistory))
	}

	// Verify history order
	expected := []string{StatusPickedUp, StatusInTransit, StatusDelivered}
	for i, exp := range expected {
		if sh.StatusHistory[i].Status != exp {
			t.Errorf("history[%d].status = %q, want %q", i, sh.StatusHistory[i].Status, exp)
		}
	}

	// Cannot advance from DELIVERED
	err = c.UpdateStatus(ctx, "LIFE-CYCLE", StatusCreated, "")
	if err == nil {
		t.Fatal("expected error advancing from DELIVERED")
	}
	if !strings.Contains(err.Error(), "already delivered") {
		t.Errorf("unexpected error: %v", err)
	}

	// GetShipment returns full history
	fetched, err := c.GetShipment(ctx, "LIFE-CYCLE")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(fetched.StatusHistory) != 3 {
		t.Errorf("fetched history len = %d", len(fetched.StatusHistory))
	}
}
