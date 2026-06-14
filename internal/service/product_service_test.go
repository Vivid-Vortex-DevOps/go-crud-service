package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/model"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ─── Mock Repository ──────────────────────────────────────────────────────────

type mockRepo struct{ mock.Mock }

func (m *mockRepo) Create(ctx context.Context, p *model.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *mockRepo) List(ctx context.Context, page, size int) ([]*model.Product, int64, error) {
	args := m.Called(ctx, page, size)
	return args.Get(0).([]*model.Product), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepo) Update(ctx context.Context, p *model.Product) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewProductService(repo)

	req := &model.CreateProductRequest{Name: "Widget", Price: 9.99, Quantity: 5}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*model.Product")).Return(nil)

	product, err := svc.Create(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "Widget", product.Name)
	assert.Equal(t, 9.99, product.Price)
	assert.NotEqual(t, uuid.Nil, product.ID)
	repo.AssertExpectations(t)
}

func TestCreate_ValidationError_EmptyName(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewProductService(repo)

	_, err := svc.Create(context.Background(), &model.CreateProductRequest{Name: "", Price: 1.0})

	require.Error(t, err)
	var validation model.ErrValidation
	assert.True(t, errors.As(err, &validation))
	repo.AssertNotCalled(t, "Create")
}

func TestCreate_ValidationError_NegativePrice(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewProductService(repo)

	_, err := svc.Create(context.Background(), &model.CreateProductRequest{Name: "X", Price: -1.0})

	require.Error(t, err)
	var validation model.ErrValidation
	assert.True(t, errors.As(err, &validation))
}

func TestGetByID_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewProductService(repo)

	id := uuid.New()
	expected := &model.Product{ID: id, Name: "Widget"}
	repo.On("GetByID", mock.Anything, id).Return(expected, nil)

	product, err := svc.GetByID(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, product.ID)
	repo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewProductService(repo)

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, model.ErrNotFound{ID: id})

	_, err := svc.GetByID(context.Background(), id)

	require.Error(t, err)
	var notFound model.ErrNotFound
	assert.True(t, errors.As(err, &notFound))
}

func TestUpdate_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewProductService(repo)

	id := uuid.New()
	existing := &model.Product{ID: id, Name: "Old", Price: 1.0, Quantity: 0}
	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*model.Product")).Return(nil)

	updated, err := svc.Update(context.Background(), id, &model.UpdateProductRequest{
		Name: "New", Price: 2.0, Quantity: 3,
	})

	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, 2.0, updated.Price)
}

func TestDelete_Success(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewProductService(repo)

	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockRepo{}
	svc := service.NewProductService(repo)

	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(model.ErrNotFound{ID: id})

	err := svc.Delete(context.Background(), id)

	var notFound model.ErrNotFound
	assert.True(t, errors.As(err, &notFound))
}
