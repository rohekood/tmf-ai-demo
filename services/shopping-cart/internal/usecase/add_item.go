package usecase

import (
	"context"
	"tmf/services/shopping-cart/internal/adapter/repository"
	"tmf/services/shopping-cart/internal/core/domain"
)

type AddItemUseCase struct {
	repo *repository.CartRepository
}

func NewAddItemUseCase(repo *repository.CartRepository) *AddItemUseCase {
	return &AddItemUseCase{repo: repo}
}

func (uc *AddItemUseCase) Execute(ctx context.Context, cartID string, offeringID string) error {
	item := domain.CartItem{
		ID:         "item-" + offeringID, // Simple ID gen
		OfferingID: offeringID,
		Quantity:   1,
	}
	return uc.repo.AddItem(ctx, cartID, item)
}
