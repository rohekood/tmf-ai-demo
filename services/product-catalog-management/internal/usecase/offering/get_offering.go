package offering

import (
	"context"

	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

type GetProductOfferingUseCase struct {
	repo         ports.ProductOfferingRepository
	specRepo     ports.ProductSpecificationRepository
	categoryRepo ports.CategoryRepository
}

func NewGetProductOffering(
	repo ports.ProductOfferingRepository,
	specRepo ports.ProductSpecificationRepository,
	categoryRepo ports.CategoryRepository,
) ports.GetProductOfferingUseCase {
	return &GetProductOfferingUseCase{
		repo:         repo,
		specRepo:     specRepo,
		categoryRepo: categoryRepo,
	}
}

func (uc *GetProductOfferingUseCase) Execute(ctx context.Context, input ports.GetProductOfferingInput) (*domain.ProductOffering, error) {
	offering, err := uc.repo.Get(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Enrich {
		if offering.ProductSpecificationID != nil {
			spec, err := uc.specRepo.Get(ctx, *offering.ProductSpecificationID)
			if err == nil {
				offering.ProductSpecification = spec
			}
		}

		for _, catID := range offering.CategoryIDs {
			cat, err := uc.categoryRepo.Get(ctx, catID)
			if err == nil {
				offering.Categories = append(offering.Categories, *cat)
			}
		}
	}

	return offering, nil
}
