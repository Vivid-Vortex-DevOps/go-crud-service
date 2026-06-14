package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateProductRequest is the API input for creating a product.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

// UpdateProductRequest is the API input for updating a product.
type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

// Validate checks CreateProductRequest business rules.
func (r *CreateProductRequest) Validate() error {
	return validateProductFields(r.Name, r.Price, r.Quantity)
}

// Validate checks UpdateProductRequest business rules.
func (r *UpdateProductRequest) Validate() error {
	return validateProductFields(r.Name, r.Price, r.Quantity)
}

func validateProductFields(name string, price float64, quantity int) error {
	if name == "" {
		return ErrValidation("name is required")
	}
	if len(name) > 255 {
		return ErrValidation("name must not exceed 255 characters")
	}
	if price < 0 {
		return ErrValidation("price must be >= 0")
	}
	if quantity < 0 {
		return ErrValidation("quantity must be >= 0")
	}
	return nil
}

// ErrValidation is a sentinel type for validation errors (maps to HTTP 422).
type ErrValidation string

func (e ErrValidation) Error() string { return string(e) }

// ErrNotFound is returned when a product cannot be found (maps to HTTP 404).
type ErrNotFound struct {
	ID uuid.UUID
}

func (e ErrNotFound) Error() string {
	return "product not found: " + e.ID.String()
}
