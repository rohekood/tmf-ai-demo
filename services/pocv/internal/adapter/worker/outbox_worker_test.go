package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"tmf/services/pocv/internal/adapter/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	err = db.AutoMigrate(&repository.OutboxEventModel{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

type mockPublisher struct {
	publish func(ctx context.Context, exchange, routingKey string, payload any) error
}

func (m *mockPublisher) Publish(ctx context.Context, exchange, routingKey string, payload any) error {
	if m.publish != nil {
		return m.publish(ctx, exchange, routingKey, payload)
	}
	return nil
}

func (m *mockPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

func TestOutboxWorker_StartAndProcessBatch(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	events := []repository.OutboxEventModel{
		{
			ID:        [16]byte{1},
			Topic:     "test.topic.1",
			Payload:   []byte(`{"k": "v"}`),
			Headers:   []byte(`{"user": "test"}`),
			Status:    repository.StatusPending,
			CreatedAt: now,
		},
		{
			ID:        [16]byte{2},
			Topic:     "test.topic.2",
			Payload:   []byte(`{"k": "v2"}`),
			Headers:   []byte(`invalid`),
			Status:    repository.StatusPending,
			CreatedAt: now,
		},
		{
			ID:        [16]byte{3},
			Topic:     "test.topic.3",
			Payload:   []byte(`{"k": "v3"}`),
			Status:    repository.StatusPending,
			CreatedAt: now,
		},
	}

	for _, e := range events {
		if err := db.Create(&e).Error; err != nil {
			t.Fatalf("failed to insert test event: %v", err)
		}
	}

	publishCount := 0
	pubMock := &mockPublisher{
		publish: func(ctx context.Context, exchange, routingKey string, payload any) error {
			publishCount++
			if routingKey == "test.topic.3" {
				return errors.New("publish failed")
			}
			return nil
		},
	}

	worker := NewOutboxWorker(db, pubMock, "test-exchange")

	worker.processBatch(context.Background())

	if publishCount != 3 {
		t.Errorf("expected 3 publishes, got %d", publishCount)
	}

	var dbEvents []repository.OutboxEventModel
	db.Order("topic").Find(&dbEvents)

	if len(dbEvents) != 3 {
		t.Fatalf("expected 3 events in db, got %d", len(dbEvents))
	}

	for _, e := range dbEvents {
		switch e.Topic {
		case "test.topic.1", "test.topic.2":
			if e.Status != repository.StatusPublished {
				t.Errorf("expected status PUBLISHED for %s, got %s", e.Topic, e.Status)
			}
		case "test.topic.3":
			if e.Status != repository.StatusFailed {
				t.Errorf("expected status FAILED for %s, got %s", e.Topic, e.Status)
			}
		}
	}

	publishCount = 0
	worker.processBatch(context.Background())
	if publishCount != 0 {
		t.Errorf("expected 0 publishes on empty batch, got %d", publishCount)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)
	go func() {
		worker.Start(ctx)
		done <- true
	}()

	time.Sleep(250 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Errorf("Start() did not return after context canceled")
	}
}

func TestOutboxWorker_DBError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	pubMock := &mockPublisher{}
	worker := NewOutboxWorker(db, pubMock, "test-exchange")

	worker.processBatch(context.Background())
}

func (m *mockPublisher) PublishToQueue(ctx context.Context, exchange string, queue string, body any) error {
	return nil
}
