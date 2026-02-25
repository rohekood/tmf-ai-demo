package repository

import (
	"bytes"
	"testing"
	"time"

	"tmf/services/pocv/internal/core/domain"
)

func TestSagaTable_TableName(t *testing.T) {
	st := SagaTable{}
	if st.TableName() != "saga_instances" {
		t.Errorf("expected saga_instances, got %s", st.TableName())
	}
}

func TestOutboxTable_TableName(t *testing.T) {
	ot := OutboxTable{}
	if ot.TableName() != "outbox_events" {
		t.Errorf("expected outbox_events, got %s", ot.TableName())
	}
}

func TestToDomainSaga(t *testing.T) {
	now := time.Now()
	customerID := "cust-1"
	dao := &SagaTable{
		ID:          "saga-1",
		CartID:      "cart-1",
		CustomerID:  &customerID,
		CurrentStep: "INVENTORY",
		Status:      "IN_PROGRESS",
		Payload:     []byte(`{"key":"value"}`),
		History:     []byte(`[]`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	domainSaga := toDomainSaga(dao)

	if domainSaga.ID != dao.ID {
		t.Errorf("expected ID %s, got %s", dao.ID, domainSaga.ID)
	}
	if domainSaga.CartID != dao.CartID {
		t.Errorf("expected CartID %s, got %s", dao.CartID, domainSaga.CartID)
	}
	if *domainSaga.CustomerID != *dao.CustomerID {
		t.Errorf("expected CustomerID %s, got %s", *dao.CustomerID, *domainSaga.CustomerID)
	}
	if domainSaga.CurrentStep != domain.StepInventory {
		t.Errorf("expected step INVENTORY, got %s", domainSaga.CurrentStep)
	}
	if domainSaga.Status != domain.SagaStatusInProgress {
		t.Errorf("expected status IN_PROGRESS, got %s", domainSaga.Status)
	}
	if !bytes.Equal(domainSaga.Payload, dao.Payload) {
		t.Errorf("payload mismatch")
	}
	if !bytes.Equal(domainSaga.History, dao.History) {
		t.Errorf("history mismatch")
	}
	if domainSaga.CreatedAt != dao.CreatedAt {
		t.Errorf("created at mismatch")
	}
	if domainSaga.UpdatedAt != dao.UpdatedAt {
		t.Errorf("updated at mismatch")
	}
}

func TestToDAOSaga(t *testing.T) {
	now := time.Now()
	customerID := "cust-1"
	domainSaga := &domain.SagaInstance{
		ID:          "saga-1",
		CartID:      "cart-1",
		CustomerID:  &customerID,
		CurrentStep: domain.StepInventory,
		Status:      domain.SagaStatusInProgress,
		Payload:     []byte(`{"key":"value"}`),
		History:     []byte(`[]`),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	dao := toDAOSaga(domainSaga)

	if dao.ID != domainSaga.ID {
		t.Errorf("expected ID %s, got %s", domainSaga.ID, dao.ID)
	}
	if dao.CartID != domainSaga.CartID {
		t.Errorf("expected CartID %s, got %s", domainSaga.CartID, dao.CartID)
	}
	if *dao.CustomerID != *domainSaga.CustomerID {
		t.Errorf("expected CustomerID %s, got %s", *domainSaga.CustomerID, *dao.CustomerID)
	}
	if dao.CurrentStep != string(domain.StepInventory) {
		t.Errorf("expected step INVENTORY, got %s", dao.CurrentStep)
	}
	if dao.Status != string(domain.SagaStatusInProgress) {
		t.Errorf("expected status IN_PROGRESS, got %s", dao.Status)
	}
	if !bytes.Equal(dao.Payload, domainSaga.Payload) {
		t.Errorf("payload mismatch")
	}
	if !bytes.Equal(dao.History, domainSaga.History) {
		t.Errorf("history mismatch")
	}
	if dao.CreatedAt != domainSaga.CreatedAt {
		t.Errorf("created at mismatch")
	}
	if dao.UpdatedAt != domainSaga.UpdatedAt {
		t.Errorf("updated at mismatch")
	}
}

func TestToDAOOutbox(t *testing.T) {
	now := time.Now()
	events := []domain.OutboxEvent{
		{
			ID:        "evt-1",
			Topic:     "test.topic",
			Payload:   []byte(`{}`),
			Status:    "PENDING",
			CreatedAt: now,
		},
		{
			ID:        "evt-2",
			Topic:     "test.topic.2",
			Payload:   []byte(`{"k":"v"}`),
			Status:    "PUBLISHED",
			CreatedAt: now,
		},
	}

	daos := toDAOOutbox(events)

	if len(daos) != 2 {
		t.Fatalf("expected 2 daos, got %d", len(daos))
	}

	if daos[0].ID != events[0].ID || daos[0].Topic != events[0].Topic || daos[0].Status != events[0].Status {
		t.Errorf("event 0 mismatch")
	}
	if daos[1].ID != events[1].ID || daos[1].Topic != events[1].Topic || daos[1].Status != events[1].Status {
		t.Errorf("event 1 mismatch")
	}
}
