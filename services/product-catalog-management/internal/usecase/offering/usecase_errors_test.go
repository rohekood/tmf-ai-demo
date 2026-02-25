package offering

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tmf/services/product-catalog-management/internal/adapter/repository"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"
)

func TestOfferingUseCases_Errors(t *testing.T) {
	ctx := context.Background()
	tm := &repository.NoOpTransactionManager{}

	t.Run("CreateOffering_ValidateError", func(t *testing.T) {
		uc := NewCreateProductOffering(nil, nil, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateProductOfferingInput{Name: ""}) // empty fails validation
		assert.Error(t, err)
	})

	t.Run("CreateOffering_SpecNotFound", func(t *testing.T) {
		specRepo := new(MockSpecRepo)
		specId := "1"
		specRepo.On("Get", ctx, "1").Return((*domain.ProductSpecification)(nil), errors.New("spec err"))
		uc := NewCreateProductOffering(nil, specRepo, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateProductOfferingInput{Name: "test", ProductSpecID: &specId})
		assert.EqualError(t, err, "spec err")
	})

	t.Run("CreateOffering_RepoError", func(t *testing.T) {
		repo := new(MockOfferingRepo)
		repo.On("Create", ctx, mock.Anything).Return(errors.New("repo err"))
		uc := NewCreateProductOffering(repo, nil, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateProductOfferingInput{Name: "test"})
		assert.EqualError(t, err, "repo err")
	})

	t.Run("CreateOffering_PublishError", func(t *testing.T) {
		repo := new(MockOfferingRepo)
		repo.On("Create", ctx, mock.Anything).Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishProductOfferingCreated", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewCreateProductOffering(repo, nil, pub, tm)
		_, err := uc.Execute(ctx, ports.CreateProductOfferingInput{Name: "test"})
		assert.EqualError(t, err, "pub err")
	})

	t.Run("UpdateOffering_GetError", func(t *testing.T) {
		repo := new(MockOfferingRepo)
		repo.On("Get", ctx, "1").Return((*domain.ProductOffering)(nil), errors.New("get err"))
		uc := NewUpdateProductOfferingUseCase(repo, nil, nil, tm)
		_, err := uc.Execute(ctx, ports.UpdateProductOfferingInput{ID: "1"})
		assert.EqualError(t, err, "get err")
	})

	t.Run("UpdateOffering_RepoError", func(t *testing.T) {
		repo := new(MockOfferingRepo)
		name := "valid"
		repo.On("Get", ctx, "1").Return(&domain.ProductOffering{ID: "1", Name: name, LifecycleStatus: "Draft"}, nil)
		repo.On("Update", ctx, mock.Anything).Return(errors.New("upd err"))
		uc := NewUpdateProductOfferingUseCase(repo, nil, nil, tm)
		_, err := uc.Execute(ctx, ports.UpdateProductOfferingInput{ID: "1", Name: &name})
		assert.EqualError(t, err, "upd err")
	})

	t.Run("UpdateOffering_PublishError", func(t *testing.T) {
		repo := new(MockOfferingRepo)
		name := "valid"
		repo.On("Get", ctx, "1").Return(&domain.ProductOffering{ID: "1", Name: name, LifecycleStatus: "Draft"}, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishProductOfferingUpdated", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewUpdateProductOfferingUseCase(repo, nil, pub, tm)
		_, err := uc.Execute(ctx, ports.UpdateProductOfferingInput{ID: "1", Name: &name})
		assert.EqualError(t, err, "pub err")
	})

	t.Run("DeleteOffering_GetError", func(t *testing.T) {
		repo := new(MockOfferingRepo)
		repo.On("Get", ctx, "1").Return((*domain.ProductOffering)(nil), errors.New("get err"))
		uc := NewDeleteProductOfferingUseCase(repo, nil, tm)
		err := uc.Execute(ctx, ports.DeleteProductOfferingInput{ID: "1"})
		assert.EqualError(t, err, "get err")
	})

	t.Run("DeleteOffering_RepoError", func(t *testing.T) {
		repo := new(MockOfferingRepo)
		repo.On("Get", ctx, "1").Return(&domain.ProductOffering{ID: "1"}, nil)
		repo.On("Delete", ctx, "1").Return(errors.New("del err"))
		uc := NewDeleteProductOfferingUseCase(repo, nil, tm)
		err := uc.Execute(ctx, ports.DeleteProductOfferingInput{ID: "1"})
		assert.EqualError(t, err, "del err")
	})

	t.Run("DeleteOffering_PublishError", func(t *testing.T) {
		repo := new(MockOfferingRepo)
		repo.On("Get", ctx, "1").Return(&domain.ProductOffering{ID: "1"}, nil)
		repo.On("Delete", ctx, "1").Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishProductOfferingDeleted", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewDeleteProductOfferingUseCase(repo, pub, tm)
		err := uc.Execute(ctx, ports.DeleteProductOfferingInput{ID: "1"})
		assert.EqualError(t, err, "pub err")
	})
}

func TestListOfferings_AllFilters(t *testing.T) {
	repo := new(MockOfferingRepo)
	uc := NewListProductOfferings(repo)
	ctx := context.Background()

	name := "test"
	cat := "cat1"
	min := 10.0
	max := 100.0
	filters := ports.ProductOfferingFilters{
		Name:     &name,
		Category: &cat,
		MinPrice: &min,
		MaxPrice: &max,
	}

	repo.On("List", ctx, mock.Anything).Return([]*domain.ProductOffering{}, nil)
	
	_, err := uc.Execute(ctx, ports.ListProductOfferingsInput{Filters: filters})
	assert.NoError(t, err)
}

func TestUpdateOffering_AllFields(t *testing.T) {
	repo := new(MockOfferingRepo)
	specRepo := new(MockSpecRepo)
	pub := new(MockEventPublisher)
	tm := &repository.NoOpTransactionManager{}
	uc := NewUpdateProductOfferingUseCase(repo, specRepo, pub, tm)
	ctx := context.Background()

	name := "name"
	desc := "desc"
	status := "Active" // Transition Draft -> Active is allowed
	start := time.Now()
	end := time.Now()
	vf := domain.TimePeriod{StartDateTime: &start, EndDateTime: &end}
	isBundle := true
	isSellable := true
	
	specID := "specId"
	cats := []string{"catId"}
	prices := []domain.ProductOfferingPrice{}
	atts := []domain.Attachment{{Name: "test"}}

	repo.On("Get", ctx, "1").Return(&domain.ProductOffering{ID: "1", Name: "old", LifecycleStatus: "Draft", ProductSpecificationID: &specID}, nil)
	specRepo.On("Get", ctx, specID).Return(&domain.ProductSpecification{ID: specID, LifecycleStatus: "Active"}, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	pub.On("PublishProductOfferingUpdated", ctx, mock.Anything).Return(nil)
	
	// mock spec repo for Create offering? Oh wait, CreateOffering error test didn't test all fields.
	
	_, err := uc.Execute(ctx, ports.UpdateProductOfferingInput{
		ID:              "1",
		Name:            &name,
		Description:     &desc,
		LifecycleStatus: &status,
		ValidFor:        &vf,
		IsBundle:        &isBundle,
		IsSellable:      &isSellable,
		CategoryIDs:     cats,
		Prices:          prices,
		Attachments:     atts,
	})
	assert.NoError(t, err)
}

func TestCreateOffering_AllFields(t *testing.T) {
	repo := new(MockOfferingRepo)
	specRepo := new(MockSpecRepo)
	pub := new(MockEventPublisher)
	tm := &repository.NoOpTransactionManager{}
	uc := NewCreateProductOffering(repo, specRepo, pub, tm)
	ctx := context.Background()

	name := "name"
	desc := "desc"
	status := ""
	start := time.Now()
	end := time.Now()
	vf := domain.TimePeriod{StartDateTime: &start, EndDateTime: &end}
	isBundle := true
	isSellable := true
	specID := "specId"
	
	cats := []string{"catId"}
	prices := []domain.ProductOfferingPrice{}
	atts := []domain.Attachment{{Name: "test"}}

	specRepo.On("Get", ctx, specID).Return(&domain.ProductSpecification{ID: specID, Name: "Spec"}, nil)
	repo.On("Create", ctx, mock.Anything).Return(nil)
	pub.On("PublishProductOfferingCreated", ctx, mock.Anything).Return(nil)

	_, err := uc.Execute(ctx, ports.CreateProductOfferingInput{
		Name:            name,
		Description:     desc,
		LifecycleStatus: status,
		ValidFor:        vf,
		IsBundle:        isBundle,
		IsSellable:      isSellable,
		ProductSpecID:   &specID,
		CategoryIDs:     cats,
		Prices:          prices,
		Attachments:     atts,
	})
	assert.NoError(t, err)
}
