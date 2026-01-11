package postgres

import (
	"context"
	"tmf/services/party-management/internal/domain"

	"gorm.io/gorm"
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Save(ctx context.Context, event *domain.OutboxEvent) error {
	tx := GetTx(ctx, r.db)
	return tx.Create(event).Error
}

func (r *OutboxRepository) FetchPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	var events []domain.OutboxEvent
	// TODO: Add SKIP LOCKED for concurrency
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.StatusPending).
		Order("created_at asc").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *OutboxRepository) MarkAsProcessed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&domain.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       domain.StatusPublished,
			"processed_at": gorm.Expr("NOW()"),
		}).Error
}
