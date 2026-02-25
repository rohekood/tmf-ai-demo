package domain

import (
	"testing"
)

func TestSagaInstance_TableName(t *testing.T) {
	s := SagaInstance{}
	if s.TableName() != "saga_instances" {
		t.Errorf("expected saga_instances, got %s", s.TableName())
	}
}

func TestOutboxEvent_TableName(t *testing.T) {
	o := OutboxEvent{}
	if o.TableName() != "outbox_events" {
		t.Errorf("expected outbox_events, got %s", o.TableName())
	}
}
