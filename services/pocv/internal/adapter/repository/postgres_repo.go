package repository

import (
	"context"
	"errors"

	"tmf/services/pocv/internal/core/domain"

	"gorm.io/gorm"
)

type txContextKey struct{}

var transactionContextKey txContextKey

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionContextKey, tx)
}

func DBFromContext(ctx context.Context) (*gorm.DB, bool) {
	if tx, ok := ctx.Value(transactionContextKey).(*gorm.DB); ok {
		return tx, true
	}
	return nil, false
}

type SagaRepository struct {
	db *gorm.DB
}

func NewSagaRepository(db *gorm.DB) *SagaRepository {
	return &SagaRepository{db: db}
}

// Helper to retrieve TX from context or use default DB
func (r *SagaRepository) getDB(ctx context.Context) *gorm.DB {
	tx, ok := DBFromContext(ctx)
	if ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *SagaRepository) GetByCartID(ctx context.Context, cartID string) (*domain.SagaInstance, error) {
	var dao SagaTable
	if err := r.getDB(ctx).First(&dao, "cart_id = ?", cartID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainSaga(&dao), nil
}

func (r *SagaRepository) Create(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
	dao := toDAOSaga(saga)
	outbox := toDAOOutbox(events)

	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(dao).Error; err != nil {
			return err
		}
		if len(outbox) > 0 {
			if err := tx.Create(&outbox).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SagaRepository) Get(ctx context.Context, id string) (*domain.SagaInstance, error) {
	var dao SagaTable
	if err := r.getDB(ctx).First(&dao, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainSaga(&dao), nil
}

func (r *SagaRepository) Update(ctx context.Context, saga *domain.SagaInstance, events []domain.OutboxEvent) error {
	dao := toDAOSaga(saga)
	outbox := toDAOOutbox(events)

	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(dao).Error; err != nil {
			return err
		}
		if len(outbox) > 0 {
			if err := tx.Create(&outbox).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
