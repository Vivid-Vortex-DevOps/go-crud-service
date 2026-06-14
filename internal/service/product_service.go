package service

import (
	"context"

	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/model"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/repository"
	"github.com/google/uuid"
)

// ProductService defines the business logic contract.
type ProductService interface {
	Create(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	List(ctx context.Context, page, size int) ([]*model.Product, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateProductRequest) (*model.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) Create(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	product := &model.Product{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Quantity:    req.Quantity,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *productService) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) List(ctx context.Context, page, size int) ([]*model.Product, int64, error) {
	return s.repo.List(ctx, page, size)
}

func (s *productService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateProductRequest) (*model.Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Price = req.Price
	existing.Quantity = req.Quantity

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *productService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
