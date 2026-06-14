package repository

import (
	"context"

	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/model"
	"github.com/google/uuid"
)

// ProductRepository defines the data access contract.
// Depending on this interface (not the concrete type) keeps the service layer testable.
type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	List(ctx context.Context, page, size int) ([]*model.Product, int64, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}
