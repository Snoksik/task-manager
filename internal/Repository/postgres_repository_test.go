package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Snoksik/task-manager/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepository_CreateAndGetByID(t *testing.T) {
	ctx := context.Background()
	repo, pool := setupPostgresRepository(t)

	input := model.Task{
		Name:   "Postgres integration test",
		UserID: 42,
	}

	result, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("expected zero error, got %v", err)
	}
	if result.ID == 0 {
		t.Fatalf("expected id not 0, got 0")
	}

	result_2, err := repo.GetByID(ctx, result.ID)

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
	ctx := context.Background()
	repo, _ := setupPostgresRepository(t)

	_, err := repo.GetByID(ctx, 999999999)
	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("expected error: %v , got: %v", model.ErrTaskNotFound, err)
	}

}

func TestPostgresRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo, pool := setupPostgresRepository(t)

	created, err := repo.Create(ctx, model.Task{Name: "wow", UserID: 5})
	if err != nil {
		t.Fatalf("expected zero errors with create, got %v", err)
	}
	t.Cleanup(func() {
		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM tasks WHERE id = $1",
			created.ID,
		)
		if err != nil {
			t.Errorf("cleanup task: %v", err)
		}
	})
	updated, err := repo.Update(ctx, created.ID, model.Task{Name: "new", UserID: 10})
	if err != nil {
		t.Fatalf("expected zero errors with update, got %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("expected id tasks don't changed, got changed set: %d != got: %d", created.ID, updated.ID)
	}
	if updated.Name != "new" {
		t.Fatalf("expected name %q, got %q", "new", updated.Name)
	}

	if updated.UserID != 10 {
		t.Fatalf("expected user id %d, got %d", 10, updated.UserID)
	}

	stored, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected zero errors with get by id, got %v", err)
	}
	if stored != updated {
		t.Fatalf("expected stored task %+v, got %+v", updated, stored)
	}
}

func TestPostgresRepository_UpdateNotFound(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupPostgresRepository(t)

	_, err := repo.Update(ctx, -1, model.Task{Name: "new", UserID: 10})
	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("expected %v, got %v", model.ErrTaskNotFound, err)
	}
}

func TestPostgresRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo, pool := setupPostgresRepository(t)

	created, err := repo.Create(ctx, model.Task{Name: "new", UserID: 10})
	if err != nil {
		t.Fatalf("expected zero errors with created, got %v", err)
	}

	t.Cleanup(func() {
		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM tasks WHERE id = $1",
			created.ID,
		)
		if err != nil {
			t.Errorf("cleanup task: %v", err)
		}
	})

	err = repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected zero errors with Delete, got %v", err)
	}
	_, err = repo.GetByID(ctx, created.ID)
	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("expected no found id after delete")
	}

}

func TestPostgresRepository_DeleteNotFound(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupPostgresRepository(t)

	err := repo.Delete(ctx, -1)
	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("expected error: %v", model.ErrTaskNotFound)
	}
}

func TestPostgresRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	repo, pool := setupPostgresRepository(t)

	marker := fmt.Sprintf("getall-%d", time.Now().UnixNano())

	first, err := repo.Create(ctx, model.Task{
		Name:   marker + " first",
		UserID: 1,
	})

	if err != nil {
		t.Fatalf("expected zero errors with created, got %v", err)
	}
	t.Cleanup(func() {
		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM tasks WHERE id = $1",
			first.ID,
		)
		if err != nil {
			t.Errorf("cleanup task: %v", err)
		}
	})

	second, err := repo.Create(ctx, model.Task{
		Name:   marker + " second",
		UserID: 1,
	})

	if err != nil {
		t.Fatalf("expected zero errors with created, got %v", err)
	}
	t.Cleanup(func() {
		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM tasks WHERE id = $1",
			second.ID,
		)
		if err != nil {
			t.Errorf("cleanup task: %v", err)
		}
	})
	third, err := repo.Create(ctx, model.Task{
		Name:   marker + " third",
		UserID: 1,
	})

	if err != nil {
		t.Fatalf("expected zero errors with created, got %v", err)
	}
	t.Cleanup(func() {
		_, err := pool.Exec(
			context.Background(),
			"DELETE FROM tasks WHERE id = $1",
			third.ID,
		)
		if err != nil {
			t.Errorf("cleanup task: %v", err)
		}
	})

	tasks, err := repo.GetAll(ctx, model.TaskFilter{Search: marker, Limit: 2, Offset: 1})
	if err != nil {
		t.Fatalf("expected zero error with get all, got %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected len tasks = 2, got %d", len(tasks))
	}
	if tasks[0] != second {
		t.Fatalf("expected task = %+v, got %+v", second, tasks[0])
	}
	if tasks[1] != third {
		t.Fatalf("expected task = %+v, got %+v", third, tasks[0])
	}
}

func setupPostgresRepository(t *testing.T) (*PostgresRepository, *pgxpool.Pool) {
	t.Helper()
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

	repo := NewPostgresRepository(pool)
	return repo, pool
}

func TestPostgresRepository_GetAllEmpty(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupPostgresRepository(t)

	tasks, err := repo.GetAll(ctx, model.TaskFilter{
		Search: "какая-то гарантированно отсутствующая строка",
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("expected zero error, got %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected len tasks = 0, got %d", len(tasks))
	}
}

func TestPostgresRepository_CreateWithCanceledContext(t *testing.T) {
	repo, _ := setupPostgresRepository(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Create(ctx, model.Task{Name: "ao", UserID: 12})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected error: context.Canceled, got %v", err)
	}
}
