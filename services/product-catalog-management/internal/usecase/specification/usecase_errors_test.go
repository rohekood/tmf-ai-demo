package specification

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

func TestSpecificationUseCases_Errors(t *testing.T) {
	ctx := context.Background()
	tm := &repository.NoOpTransactionManager{}

	t.Run("CreateSpecification_ValidateError", func(t *testing.T) {
		uc := NewCreateProductSpecification(nil, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateProductSpecificationInput{Name: ""}) // empty fails validation
		assert.Error(t, err)
	})

	t.Run("CreateSpecification_RepoError", func(t *testing.T) {
		repo := new(MockSpecificationRepo)
		repo.On("Create", ctx, mock.Anything).Return(errors.New("repo err"))
		uc := NewCreateProductSpecification(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateProductSpecificationInput{Name: "test", ProductNumber: "1"})
		assert.EqualError(t, err, "repo err")
	})

	t.Run("CreateSpecification_PublishError", func(t *testing.T) {
		repo := new(MockSpecificationRepo)
		repo.On("Create", ctx, mock.Anything).Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishProductSpecificationCreated", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewCreateProductSpecification(repo, pub, tm)
		_, err := uc.Execute(ctx, ports.CreateProductSpecificationInput{Name: "test", ProductNumber: "1"})
		assert.EqualError(t, err, "pub err")
	})

	t.Run("UpdateSpecification_GetError", func(t *testing.T) {
		repo := new(MockSpecificationRepo)
		repo.On("Get", ctx, "1").Return((*domain.ProductSpecification)(nil), errors.New("get err"))
		uc := NewUpdateProductSpecificationUseCase(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.UpdateProductSpecificationInput{ID: "1"})
		assert.EqualError(t, err, "get err")
	})

	t.Run("UpdateSpecification_RepoError", func(t *testing.T) {
		repo := new(MockSpecificationRepo)
		name := "valid"
		repo.On("Get", ctx, "1").Return(&domain.ProductSpecification{ID: "1", Name: name, ProductNumber: "1"}, nil)
		repo.On("Update", ctx, mock.Anything).Return(errors.New("upd err"))
		uc := NewUpdateProductSpecificationUseCase(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.UpdateProductSpecificationInput{ID: "1", Name: &name})
		assert.EqualError(t, err, "upd err")
	})

	t.Run("UpdateSpecification_PublishError", func(t *testing.T) {
		repo := new(MockSpecificationRepo)
		name := "valid"
		repo.On("Get", ctx, "1").Return(&domain.ProductSpecification{ID: "1", Name: name, ProductNumber: "1"}, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishProductSpecificationUpdated", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewUpdateProductSpecificationUseCase(repo, pub, tm)
		_, err := uc.Execute(ctx, ports.UpdateProductSpecificationInput{ID: "1", Name: &name})
		assert.EqualError(t, err, "pub err")
	})

	t.Run("DeleteSpecification_GetError", func(t *testing.T) {
		repo := new(MockSpecificationRepo)
		repo.On("Get", ctx, "1").Return((*domain.ProductSpecification)(nil), errors.New("get err"))
		uc := NewDeleteProductSpecificationUseCase(repo, nil, tm)
		err := uc.Execute(ctx, ports.DeleteProductSpecificationInput{ID: "1"})
		assert.EqualError(t, err, "get err")
	})

	t.Run("DeleteSpecification_RepoError", func(t *testing.T) {
		repo := new(MockSpecificationRepo)
		repo.On("Get", ctx, "1").Return(&domain.ProductSpecification{ID: "1"}, nil)
		repo.On("Delete", ctx, "1").Return(errors.New("del err"))
		uc := NewDeleteProductSpecificationUseCase(repo, nil, tm)
		err := uc.Execute(ctx, ports.DeleteProductSpecificationInput{ID: "1"})
		assert.EqualError(t, err, "del err")
	})

	t.Run("DeleteSpecification_PublishError", func(t *testing.T) {
		repo := new(MockSpecificationRepo)
		repo.On("Get", ctx, "1").Return(&domain.ProductSpecification{ID: "1"}, nil)
		repo.On("Delete", ctx, "1").Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishProductSpecificationDeleted", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewDeleteProductSpecificationUseCase(repo, pub, tm)
		err := uc.Execute(ctx, ports.DeleteProductSpecificationInput{ID: "1"})
		assert.EqualError(t, err, "pub err")
	})
}

func TestUpdateSpec_AllFields(t *testing.T) {
	repo := new(MockSpecificationRepo)
	pub := new(MockEventPublisher)
	tm := &repository.NoOpTransactionManager{}
	uc := NewUpdateProductSpecificationUseCase(repo, pub, tm)
	ctx := context.Background()

	name := "name"
	desc := "desc"
	status := "Draft"
	start := time.Now()
	end := time.Now()
	vf := domain.TimePeriod{StartDateTime: &start, EndDateTime: &end}
	isBundle := true
	pnum := "123"
	chars := map[string]domain.ProductSpecCharacteristic{"a": {}}

	repo.On("Get", ctx, "1").Return(&domain.ProductSpecification{ID: "1", Name: "old"}, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	pub.On("PublishProductSpecificationUpdated", ctx, mock.Anything).Return(nil)

	_, err := uc.Execute(ctx, ports.UpdateProductSpecificationInput{
		ID:              "1",
		Name:            &name,
		Description:     &desc,
		LifecycleStatus: &status,
		ValidFor:        &vf,
		IsBundle:        &isBundle,
		ProductNumber:   &pnum,
		Characteristics: chars,
	})
	assert.NoError(t, err)
}
