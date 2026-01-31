package repository

import (
	"context"
	"errors"
	"fmt"

	"tmf/services/shopping-cart/internal/core/domain"
	"tmf/services/shopping-cart/internal/core/ports"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type postgresCartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) ports.CartRepository {
	return &postgresCartRepo{db: db}
}

// GetTx extracts the transaction from context or returns the DB
func GetTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return tx
	}
	return db.WithContext(ctx)
}

// WithUser executes the function within a transaction context that has app.current_user set.
func (r *postgresCartRepo) WithUser(ctx context.Context, fn func(tx *gorm.DB) error) error {
	userID, _ := ctx.Value("X-User-ID").(string)
	if userID == "" {
		userID = "system"
	}

	runWithUser := func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT set_config('app.current_user', ?, true)", userID).Error; err != nil {
			return err
		}
		return fn(tx)
	}

	if existingTx, ok := ctx.Value("tx").(*gorm.DB); ok {
		return runWithUser(existingTx)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return runWithUser(tx)
	})
}

func (r *postgresCartRepo) Get(ctx context.Context, id string) (*domain.Cart, error) {
	var dao CartTable
	err := GetTx(ctx, r.db).Preload("Items").First(&dao, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("db get failed: %w", err)
	}
	return toDomainCart(&dao), nil
}

func (r *postgresCartRepo) Save(ctx context.Context, cart *domain.Cart, events []domain.OutboxEvent) error {
	dao := toDAOCart(cart)
	outboxDAOs := toDAOOutbox(events)

	return r.WithUser(ctx, func(tx *gorm.DB) error {
		// UPSERT Strategy using Clauses to ensure Parent exists before Children
		// 1. Save Parent (Cart)
		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Omit("Items").Create(dao).Error; err != nil {
			return err
		}

		// 2. Save Items
		// Strategy: Delete all existing items and re-insert (easiest for full replace)
		// Or Upsert items. Given we append in UseCase, full replace is safer to sync state.
		if err := tx.Delete(&CartItemTable{}, "cart_id = ?", dao.ID).Error; err != nil {
			return err
		}

		if len(dao.Items) > 0 {
			if err := tx.Create(&dao.Items).Error; err != nil {
				return err
			}
		}

		// 3. Save Events
		if len(outboxDAOs) > 0 {
			if err := tx.Create(&outboxDAOs).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *postgresCartRepo) UpsertPrice(ctx context.Context, price *domain.ProductPrice) error {
	dao := ProductPriceTable{
		ID:         price.ID,
		UnitAmount: price.UnitAmount,
		Currency:   price.Currency,
		UpdatedAt:  price.UpdatedAt,
	}
	return r.WithUser(ctx, func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&dao).Error
	})
}

func (r *postgresCartRepo) GetPrice(ctx context.Context, offeringID string) (*domain.ProductPrice, error) {
	var dao ProductPriceTable
	err := GetTx(ctx, r.db).First(&dao, "id = ?", offeringID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.ProductPrice{
		ID:         dao.ID,
		UnitAmount: dao.UnitAmount,
		Currency:   dao.Currency,
		UpdatedAt:  dao.UpdatedAt,
	}, nil
}
