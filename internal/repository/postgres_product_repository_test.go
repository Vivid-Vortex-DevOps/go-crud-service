//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/model"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		CREATE EXTENSION IF NOT EXISTS pgcrypto;
		CREATE TABLE IF NOT EXISTS products (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name        VARCHAR(255) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			price       NUMERIC(10,2) NOT NULL CHECK (price >= 0),
			quantity    INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	require.NoError(t, err)

	return pool, func() {
		pool.Close()
		pgContainer.Terminate(ctx) //nolint:errcheck
	}
}

// insert is a helper that sets a UUID and calls repo.Create (which mutates p in-place).
func insert(t *testing.T, repo repository.ProductRepository, ctx context.Context, name string, price float64, quantity int) *model.Product {
	t.Helper()
	p := &model.Product{
		ID:          uuid.New(),
		Name:        name,
		Description: "test",
		Price:       price,
		Quantity:    quantity,
	}
	require.NoError(t, repo.Create(ctx, p))
	return p
}

func TestPostgresProductRepository_Create(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	p := &model.Product{
		ID:          uuid.New(),
		Name:        "Integration Test Widget",
		Description: "Created in integration test",
		Price:       19.99,
		Quantity:    50,
	}

	err := repo.Create(ctx, p)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, p.ID)
	assert.Equal(t, "Integration Test Widget", p.Name)
	assert.InDelta(t, 19.99, p.Price, 0.001)
	assert.Equal(t, 50, p.Quantity)
	assert.False(t, p.CreatedAt.IsZero(), "CreatedAt should be set by RETURNING clause")
}

func TestPostgresProductRepository_GetByID_NotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	var notFound model.ErrNotFound
	assert.ErrorAs(t, err, &notFound)
}

func TestPostgresProductRepository_List(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	for i := range 3 {
		insert(t, repo, ctx, "Product "+string(rune('A'+i)), float64(i+1)*10.0, 10)
	}

	products, total, err := repo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, products, 3)
}

func TestPostgresProductRepository_Update(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	p := insert(t, repo, ctx, "Original Name", 9.99, 5)

	p.Name = "Updated Name"
	p.Price = 19.99
	err := repo.Update(ctx, p)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", p.Name)
	assert.InDelta(t, 19.99, p.Price, 0.001)
	assert.False(t, p.UpdatedAt.IsZero())
}

func TestPostgresProductRepository_Delete(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	p := insert(t, repo, ctx, "ToDelete", 1.0, 1)

	err := repo.Delete(ctx, p.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, p.ID)
	var notFound model.ErrNotFound
	assert.ErrorAs(t, err, &notFound)
}
