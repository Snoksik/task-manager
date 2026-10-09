package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Snoksik/task-manager/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool: pool,
	}
}

func (r *PostgresRepository) Create(task model.Task) (model.Task, error) {
	ctx := context.Background()
	query := `INSERT INTO tasks (name, user_id)
		VALUES ($1, $2)
		RETURNING id`
	err := r.pool.QueryRow(ctx, query, task.Name, task.UserID).Scan(&task.ID)
	if err != nil {
		return model.Task{}, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}

func (r *PostgresRepository) GetByID(id int) (model.Task, error) {
	ctx := context.Background()

	query := `SELECT id, name, user_id
				FROM tasks
				WHERE id = $1;`

	result := model.Task{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&result.ID, &result.Name, &result.UserID)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Task{}, model.ErrTaskNotFound
	}

	if err != nil {
		return model.Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return result, nil
}
