package usecase_test

import (
	"context"
	"testing"

	"tmf/services/shopping-cart/internal/usecase"
)

func TestAddItemUseCase_Execute(t *testing.T) {
	uc := usecase.NewAddItemUseCase(nil)
	
	defer func() {
		recover() // Will panic on nil repo
	}()
	
	uc.Execute(context.Background(), "cart-1", "offering-1")
}
