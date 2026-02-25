package worker_test

import (
	"context"
	"os"
	"testing"
	"time"

	"tmf/services/shopping-cart/internal/adapter/repository"
	"tmf/services/shopping-cart/internal/adapter/worker"

	"github.com/google/uuid"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockPublisher mocks the Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, exchange, routingKey string, payload interface{}) error {
	args := m.Called(ctx, exchange, routingKey, payload)
	return args.Error(0)
}

func (m *MockPublisher) PublishToQueue(ctx context.Context, queueName string, correlationID string, body interface{}) error {
	args := m.Called(ctx, queueName, correlationID, body)
	return args.Error(0)
}

func (m *MockPublisher) DeclareTopicExchange(name string, durable, autoDelete, internal, noWait bool) error {
	args := m.Called(name, durable, autoDelete, internal, noWait)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func setupDB(t *testing.T) *gorm.DB {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable"
	}

	m, err := migrate.New("file://../../../internal/infrastructure/postgres/migrations", dbURL)
	require.NoError(t, err)
	_ = m.Up()

	db, err := gorm.Open(gormPostgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func TestOutboxWorker(t *testing.T) {
	db := setupDB(t)
	mockPub := new(MockPublisher)
	exchangeName := "ex.test"

	w := worker.NewOutboxWorker(db, mockPub, exchangeName)

	t.Run("Should process pending events", func(t *testing.T) {
		// removed
		evtID := uuid.New().String()
		err := db.Create(&repository.OutboxTable{
			ID:        evtID,
			Topic:     "test-topic",
			Payload:   []byte(`{"foo":"bar"}`),
			Status:    "PENDING",
			CreatedAt: time.Now().UTC(),
		}).Error
		require.NoError(t, err)

		mockPub.On("Publish", mock.Anything, exchangeName, mock.Anything, mock.Anything).Return(nil).Maybe()

		ctx, cancel := context.WithCancel(context.Background())
		go w.Start(ctx)
		time.Sleep(2 * time.Second)
		cancel()

		var status string
		err = db.Model(&repository.OutboxTable{}).Where("id = ?", evtID).Select("status").Scan(&status).Error
		require.NoError(t, err)
		assert.Equal(t, "PUBLISHED", status)

		mockPub.AssertExpectations(t)
	})

	t.Run("Should handle no pending events", func(t *testing.T) {
		// removed
		ctx, cancel := context.WithCancel(context.Background())
		go w.Start(ctx)
		time.Sleep(600 * time.Millisecond)
		cancel()
	})

	t.Run("Should handle publish error", func(t *testing.T) {
		// removed
		evtID := uuid.New().String()
		db.Create(&repository.OutboxTable{
			ID:        evtID,
			Topic:     "test-topic",
			Payload:   []byte(`{"foo":"bar"}`),
			Status:    "PENDING",
			CreatedAt: time.Now().UTC(),
		})

		mockPub2 := new(MockPublisher)
		mockPub2.On("Publish", mock.Anything, exchangeName, mock.Anything, mock.Anything).Return(os.ErrPermission)
		w2 := worker.NewOutboxWorker(db, mockPub2, exchangeName)

		ctx, cancel := context.WithCancel(context.Background())
		go w2.Start(ctx)
		time.Sleep(2 * time.Second)
		cancel()

		var status string
		db.Model(&repository.OutboxTable{}).Where("id = ?", evtID).Select("status").Scan(&status)
		assert.Equal(t, "PENDING", status)
	})

	t.Run("Should handle DB error", func(t *testing.T) {
		dbBad, _ := gorm.Open(gormPostgres.Open("postgres://backstage:backstage@localhost:5432/backstage?sslmode=disable"), &gorm.Config{})
		dbSQL, _ := dbBad.DB()
		dbSQL.Close()

		wBad := worker.NewOutboxWorker(dbBad, new(MockPublisher), exchangeName)
		ctx, cancel := context.WithCancel(context.Background())
		go wBad.Start(ctx)
		time.Sleep(600 * time.Millisecond)
		cancel()
	})
}
