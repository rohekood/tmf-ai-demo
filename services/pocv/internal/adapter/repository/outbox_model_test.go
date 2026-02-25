package repository

import (
	"encoding/json"
	"testing"
)

func TestOutboxEventModel_TableName(t *testing.T) {
	m := OutboxEventModel{}
	if m.TableName() != "outbox_events" {
		t.Errorf("expected outbox_events, got %s", m.TableName())
	}
}

func TestNewOutboxEvent(t *testing.T) {
	topic := "test.topic"
	payload := map[string]string{"key": "value"}
	headers := map[string]string{"user": "test-user"}

	evt, err := NewOutboxEvent(topic, payload, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if evt.Topic != topic {
		t.Errorf("expected topic %s, got %s", topic, evt.Topic)
	}

	if evt.Status != StatusPending {
		t.Errorf("expected status %s, got %s", StatusPending, evt.Status)
	}

	var payloadMap map[string]string
	if err := json.Unmarshal(evt.Payload, &payloadMap); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payloadMap["key"] != "value" {
		t.Errorf("expected payload key value, got %s", payloadMap["key"])
	}

	var headerMap map[string]string
	if err := json.Unmarshal(evt.Headers, &headerMap); err != nil {
		t.Fatalf("failed to unmarshal headers: %v", err)
	}
	if headerMap["user"] != "test-user" {
		t.Errorf("expected header user test-user, got %s", headerMap["user"])
	}

	// Test invalid payload
	_, err = NewOutboxEvent(topic, make(chan int), headers)
	if err == nil {
		t.Error("expected error for invalid payload")
	}

	// Test invalid headers - not possible with map[string]string unless it's nil which still marshals to "null"
}
