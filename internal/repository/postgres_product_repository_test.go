//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/model"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/repository"
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
			description TEXT,
			price       NUMERIC(10,2) NOT NULL CHECK (price > 0),
			stock       INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
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

func TestPostgresProductRepository_Create(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	p := &model.Product{
		Name:        "Integration Test Widget",
		Description: "Created in integration test",
		Price:       19.99,
		Stock:       50,
	}

	created, err := repo.Create(ctx, p)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Integration Test Widget", created.Name)
	assert.InDelta(t, 19.99, created.Price, 0.001)
	assert.Equal(t, 50, created.Stock)
}

func TestPostgresProductRepository_GetByID_NotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000001")
	assert.ErrorIs(t, err, model.ErrNotFound{ID: "00000000-0000-0000-0000-000000000001"})
}

func TestPostgresProductRepository_List(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	for i := range 3 {
		_, err := repo.Create(ctx, &model.Product{
			Name:  "Product " + string(rune('A'+i)),
			Price: float64(i+1) * 10.0,
			Stock: 10,
		})
		require.NoError(t, err)
	}

	products, total, err := repo.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, products, 3)
}

func TestPostgresProductRepository_Update(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &model.Product{
		Name:  "Original Name",
		Price: 9.99,
		Stock: 5,
	})
	require.NoError(t, err)

	created.Name = "Updated Name"
	created.Price = 19.99
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.InDelta(t, 19.99, updated.Price, 0.001)
}

func TestPostgresProductRepository_Delete(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewPostgresProductRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, &model.Product{Name: "ToDelete", Price: 1.0, Stock: 1})
	require.NoError(t, err)

	err = repo.Delete(ctx, created.ID.String())
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, created.ID.String())
	assert.Error(t, err)
}
