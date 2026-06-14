package repository

import (
	"context"
	"errors"

	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("go-crud-service/repository")

type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *model.Product) error {
	ctx, span := tracer.Start(ctx, "postgres.Product.Create")
	defer span.End()

	span.SetAttributes(attribute.String("product.name", p.Name))

	query := `
		INSERT INTO products (id, name, description, price, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		p.ID, p.Name, p.Description, p.Price, p.Quantity,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	ctx, span := tracer.Start(ctx, "postgres.Product.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("product.id", id.String()))

	p := &model.Product{}
	query := `
		SELECT id, name, description, price, quantity, created_at, updated_at
		FROM products WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Quantity,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound{ID: id}
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, page, size int) ([]*model.Product, int64, error) {
	ctx, span := tracer.Start(ctx, "postgres.Product.List")
	defer span.End()

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, price, quantity, created_at, updated_at
		 FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		size, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p := &model.Product{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Quantity, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *model.Product) error {
	ctx, span := tracer.Start(ctx, "postgres.Product.Update")
	defer span.End()

	span.SetAttributes(attribute.String("product.id", p.ID.String()))

	query := `
		UPDATE products
		SET name = $2, description = $3, price = $4, quantity = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, query,
		p.ID, p.Name, p.Description, p.Price, p.Quantity,
	).Scan(&p.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrNotFound{ID: p.ID}
	}
	return err
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := tracer.Start(ctx, "postgres.Product.Delete")
	defer span.End()

	span.SetAttributes(attribute.String("product.id", id.String()))

	result, err := r.pool.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return model.ErrNotFound{ID: id}
	}
	return nil
}
