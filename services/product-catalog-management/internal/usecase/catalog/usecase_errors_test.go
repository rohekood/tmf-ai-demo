package catalog

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

func TestCatalogUseCases_Errors(t *testing.T) {
	ctx := context.Background()
	tm := &repository.NoOpTransactionManager{}

	t.Run("CreateCatalog_ValidateError", func(t *testing.T) {
		uc := NewCreateCatalog(nil, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateCatalogInput{Name: ""}) // empty name fails validation
		assert.Error(t, err)
	})

	t.Run("CreateCatalog_RepoError", func(t *testing.T) {
		repo := new(MockCatalogRepo)
		repo.On("Create", ctx, mock.Anything).Return(errors.New("repo err"))
		uc := NewCreateCatalog(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateCatalogInput{Name: "test"})
		assert.EqualError(t, err, "repo err")
	})

	t.Run("CreateCatalog_PublishError", func(t *testing.T) {
		repo := new(MockCatalogRepo)
		repo.On("Create", ctx, mock.Anything).Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishCatalogCreated", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewCreateCatalog(repo, pub, tm)
		_, err := uc.Execute(ctx, ports.CreateCatalogInput{Name: "test"})
		assert.EqualError(t, err, "pub err")
	})

	t.Run("UpdateCatalog_GetError", func(t *testing.T) {
		repo := new(MockCatalogRepo)
		repo.On("Get", ctx, "1").Return((*domain.Catalog)(nil), errors.New("get err"))
		uc := NewUpdateCatalogUseCase(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.UpdateCatalogInput{ID: "1"})
		assert.EqualError(t, err, "get err")
	})

	t.Run("UpdateCatalog_RepoError", func(t *testing.T) {
		repo := new(MockCatalogRepo)
		name := "valid"
		repo.On("Get", ctx, "1").Return(&domain.Catalog{ID: "1", Name: name}, nil)
		repo.On("Update", ctx, mock.Anything).Return(errors.New("upd err"))
		uc := NewUpdateCatalogUseCase(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.UpdateCatalogInput{ID: "1", Name: &name})
		assert.EqualError(t, err, "upd err")
	})

	t.Run("UpdateCatalog_PublishError", func(t *testing.T) {
		repo := new(MockCatalogRepo)
		name := "valid"
		repo.On("Get", ctx, "1").Return(&domain.Catalog{ID: "1", Name: name}, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishCatalogUpdated", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewUpdateCatalogUseCase(repo, pub, tm)
		_, err := uc.Execute(ctx, ports.UpdateCatalogInput{ID: "1", Name: &name})
		assert.EqualError(t, err, "pub err")
	})

	t.Run("DeleteCatalog_GetError", func(t *testing.T) {
		repo := new(MockCatalogRepo)
		repo.On("Get", ctx, "1").Return((*domain.Catalog)(nil), errors.New("get err"))
		uc := NewDeleteCatalogUseCase(repo, nil, tm)
		err := uc.Execute(ctx, ports.DeleteCatalogInput{ID: "1"})
		assert.EqualError(t, err, "get err")
	})

	t.Run("DeleteCatalog_RepoError", func(t *testing.T) {
		repo := new(MockCatalogRepo)
		repo.On("Get", ctx, "1").Return(&domain.Catalog{ID: "1"}, nil)
		repo.On("Delete", ctx, "1").Return(errors.New("del err"))
		uc := NewDeleteCatalogUseCase(repo, nil, tm)
		err := uc.Execute(ctx, ports.DeleteCatalogInput{ID: "1"})
		assert.EqualError(t, err, "del err")
	})

	t.Run("DeleteCatalog_PublishError", func(t *testing.T) {
		repo := new(MockCatalogRepo)
		repo.On("Get", ctx, "1").Return(&domain.Catalog{ID: "1"}, nil)
		repo.On("Delete", ctx, "1").Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishCatalogDeleted", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewDeleteCatalogUseCase(repo, pub, tm)
		err := uc.Execute(ctx, ports.DeleteCatalogInput{ID: "1"})
		assert.EqualError(t, err, "pub err")
	})
}

func TestUpdateCatalog_AllFields(t *testing.T) {
	repo := new(MockCatalogRepo)
	pub := new(MockEventPublisher)
	tm := &repository.NoOpTransactionManager{}
	uc := NewUpdateCatalogUseCase(repo, pub, tm)
	ctx := context.Background()

	name := "name"
	desc := "desc"
	status := "Draft"
	start := time.Now()
	end := time.Now()
	vf := domain.TimePeriod{StartDateTime: &start, EndDateTime: &end}

	repo.On("Get", ctx, "1").Return(&domain.Catalog{ID: "1", Name: "old"}, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	pub.On("PublishCatalogUpdated", ctx, mock.Anything).Return(nil)

	_, err := uc.Execute(ctx, ports.UpdateCatalogInput{
		ID:              "1",
		Name:            &name,
		Description:     &desc,
		LifecycleStatus: &status,
		ValidFor:        &vf,
	})
	assert.NoError(t, err)
}
