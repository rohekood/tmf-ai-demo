package category

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

func TestCategoryUseCases_Errors(t *testing.T) {
	ctx := context.Background()
	tm := &repository.NoOpTransactionManager{}

	t.Run("CreateCategory_ValidateError", func(t *testing.T) {
		uc := NewCreateCategory(nil, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateCategoryInput{Name: ""}) // empty name fails validation
		assert.Error(t, err)
	})

	t.Run("CreateCategory_RepoError", func(t *testing.T) {
		repo := new(MockCategoryRepo)
		repo.On("Create", ctx, mock.Anything).Return(errors.New("repo err"))
		uc := NewCreateCategory(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.CreateCategoryInput{Name: "test"})
		assert.EqualError(t, err, "repo err")
	})

	t.Run("CreateCategory_PublishError", func(t *testing.T) {
		repo := new(MockCategoryRepo)
		repo.On("Create", ctx, mock.Anything).Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishCategoryCreated", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewCreateCategory(repo, pub, tm)
		_, err := uc.Execute(ctx, ports.CreateCategoryInput{Name: "test"})
		assert.EqualError(t, err, "pub err")
	})

	t.Run("UpdateCategory_GetError", func(t *testing.T) {
		repo := new(MockCategoryRepo)
		repo.On("Get", ctx, "1").Return((*domain.Category)(nil), errors.New("get err"))
		uc := NewUpdateCategoryUseCase(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.UpdateCategoryInput{ID: "1"})
		assert.EqualError(t, err, "get err")
	})

	t.Run("UpdateCategory_RepoError", func(t *testing.T) {
		repo := new(MockCategoryRepo)
		name := "valid"
		repo.On("Get", ctx, "1").Return(&domain.Category{ID: "1", Name: name}, nil)
		repo.On("Update", ctx, mock.Anything).Return(errors.New("upd err"))
		uc := NewUpdateCategoryUseCase(repo, nil, tm)
		_, err := uc.Execute(ctx, ports.UpdateCategoryInput{ID: "1", Name: &name})
		assert.EqualError(t, err, "upd err")
	})

	t.Run("UpdateCategory_PublishError", func(t *testing.T) {
		repo := new(MockCategoryRepo)
		name := "valid"
		repo.On("Get", ctx, "1").Return(&domain.Category{ID: "1", Name: name}, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishCategoryUpdated", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewUpdateCategoryUseCase(repo, pub, tm)
		_, err := uc.Execute(ctx, ports.UpdateCategoryInput{ID: "1", Name: &name})
		assert.EqualError(t, err, "pub err")
	})

	t.Run("DeleteCategory_GetError", func(t *testing.T) {
		repo := new(MockCategoryRepo)
		repo.On("Get", ctx, "1").Return((*domain.Category)(nil), errors.New("get err"))
		uc := NewDeleteCategoryUseCase(repo, nil, tm)
		err := uc.Execute(ctx, ports.DeleteCategoryInput{ID: "1"})
		assert.EqualError(t, err, "get err")
	})

	t.Run("DeleteCategory_RepoError", func(t *testing.T) {
		repo := new(MockCategoryRepo)
		repo.On("Get", ctx, "1").Return(&domain.Category{ID: "1"}, nil)
		repo.On("Delete", ctx, "1").Return(errors.New("del err"))
		uc := NewDeleteCategoryUseCase(repo, nil, tm)
		err := uc.Execute(ctx, ports.DeleteCategoryInput{ID: "1"})
		assert.EqualError(t, err, "del err")
	})

	t.Run("DeleteCategory_PublishError", func(t *testing.T) {
		repo := new(MockCategoryRepo)
		repo.On("Get", ctx, "1").Return(&domain.Category{ID: "1"}, nil)
		repo.On("Delete", ctx, "1").Return(nil)
		pub := new(MockEventPublisher)
		pub.On("PublishCategoryDeleted", ctx, mock.Anything).Return(errors.New("pub err"))
		uc := NewDeleteCategoryUseCase(repo, pub, tm)
		err := uc.Execute(ctx, ports.DeleteCategoryInput{ID: "1"})
		assert.EqualError(t, err, "pub err")
	})
}

func TestUpdateCategory_AllFields(t *testing.T) {
	repo := new(MockCategoryRepo)
	pub := new(MockEventPublisher)
	tm := &repository.NoOpTransactionManager{}
	uc := NewUpdateCategoryUseCase(repo, pub, tm)
	ctx := context.Background()

	name := "name"
	desc := "desc"
	status := "Draft"
	start := time.Now()
	end := time.Now()
	vf := domain.TimePeriod{StartDateTime: &start, EndDateTime: &end}
	parent := "parent"
	isRoot := false
	catID := "catId"

	repo.On("Get", ctx, "1").Return(&domain.Category{ID: "1", Name: "old"}, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)
	pub.On("PublishCategoryUpdated", ctx, mock.Anything).Return(nil)

	_, err := uc.Execute(ctx, ports.UpdateCategoryInput{
		ID:              "1",
		Name:            &name,
		Description:     &desc,
		LifecycleStatus: &status,
		ValidFor:        &vf,
		ParentID:        &parent,
		IsRoot:          &isRoot,
		CatalogID:       &catID,
	})
	assert.NoError(t, err)
}
