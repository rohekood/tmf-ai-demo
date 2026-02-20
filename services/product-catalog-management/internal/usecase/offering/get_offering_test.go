package offering

import (
	"context"
	"testing"
	"tmf/services/product-catalog-management/internal/core/domain"
	"tmf/services/product-catalog-management/internal/core/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCategoryRepoLocal struct {
	mock.Mock
}

func (m *MockCategoryRepoLocal) Create(ctx context.Context, category *domain.Category) error {
	return nil
}
func (m *MockCategoryRepoLocal) Get(ctx context.Context, id string) (*domain.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Category), args.Error(1)
}
func (m *MockCategoryRepoLocal) List(ctx context.Context, filters map[string]interface{}) ([]*domain.Category, error) {
	return nil, nil
}
func (m *MockCategoryRepoLocal) Update(ctx context.Context, category *domain.Category) error {
	return nil
}
func (m *MockCategoryRepoLocal) Delete(ctx context.Context, id string) error {
	return nil
}

func TestGetProductOffering_Execute(t *testing.T) {
	mockRepo := new(MockOfferingRepo)
	mockSpecRepo := new(MockSpecRepo)
	// Need MockCategoryRepo too, likely defined in another package or need to define/import it.
	// Since we are in `offering` package, we can define a stub or reuse if exported.
	// `mock_repo_test.go` has MockOfferingRepo and MockSpecRepo.
	// I'll add MockCategoryRepo to `mock_repo_test.go` if not present (it's not).
	// For now, I'll define it here locally or check `services/product-catalog-management/internal/usecase/category/mock_repo_test.go`.
	// Go doesn't allow importing `.../category` easily due to cyclic deps if not careful, but here it's fine.
	// Actually, better to add MockCategoryRepo to `offering` package test helpers.
	mockCatRepo := new(MockCategoryRepoLocal)

	useCase := NewGetProductOffering(mockRepo, mockSpecRepo, mockCatRepo)

	ctx := context.Background()
	id := "off-1"
	expected := &domain.ProductOffering{ID: id, Name: "Found"}

	mockRepo.On("Get", ctx, id).Return(expected, nil)

	res, err := useCase.Execute(ctx, ports.GetProductOfferingInput{ID: id})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)

	mockRepo.AssertExpectations(t)
}
