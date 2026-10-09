package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/Snoksik/task-manager/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepository_CreateAndGetByID(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	repo := NewPostgresRepository(pool)

	input := model.Task{
		Name:   "Postgres integration test",
		UserID: 42,
	}

	result, err := repo.Create(input)
	if err != nil {
		t.Fatalf("expected zero error, got %v", err)
	}
	if result.ID == 0 {
		t.Fatalf("expected id not 0, got 0")
	}

	result_2, err := repo.GetByID(result.ID)

	t.Cleanup(func() {
		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM tasks WHERE id = $1",
			result.ID,
		)
		if err != nil {
			t.Errorf("cleanup task: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("expected zero error, got %v", err)
	}
	if result_2 != result {
		t.Fatalf("Expected %+v, got %+v", result, result_2)
	}

}

func TestPostgresRepository_GetByIDNotFound(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	repo := NewPostgresRepository(pool)

	_, err = repo.GetByID(999999999)
	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("expected error: %v , got: %v", model.ErrTaskNotFound, err)
	}

}
