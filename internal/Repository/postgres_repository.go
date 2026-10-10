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

func (r *PostgresRepository) Create(ctx context.Context, task model.Task) (model.Task, error) {
	query := `INSERT INTO tasks (name, user_id)
		VALUES ($1, $2)
		RETURNING id`
	err := r.pool.QueryRow(ctx, query, task.Name, task.UserID).Scan(&task.ID)
	if err != nil {
		return model.Task{}, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int) (model.Task, error) {

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

func (r *PostgresRepository) Update(ctx context.Context, id int, task model.Task) (model.Task, error) {

	query := `UPDATE tasks
				SET name = $1, user_id = $2
				WHERE id = $3
				RETURNING id, name, user_id;`
	result := model.Task{}
	err := r.pool.QueryRow(ctx, query, task.Name, task.UserID, id).Scan(&result.ID, &result.Name, &result.UserID)

	if errors.Is(err, pgx.ErrNoRows) {
		return model.Task{}, model.ErrTaskNotFound
	}

	if err != nil {
		return model.Task{}, fmt.Errorf("update task by id: %w", err)
	}

	return result, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int) error {

	query := `DELETE FROM tasks
				WHERE id = $1;`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task by id: %w", err)

	}
	if result.RowsAffected() == 0 {
		return model.ErrTaskNotFound
	}
	return nil
}

func (r *PostgresRepository) GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	query := `SELECT id, name, user_id FROM tasks
				WHERE name LIKE '%' || $1 || '%'
				ORDER BY id ASC
				limit $2
				offset $3;
				 `

	rows, err := r.pool.Query(ctx, query, filter.Search, filter.Limit, filter.Offset)
	if err != nil {
		return []model.Task{}, fmt.Errorf("get all task: %w", err)
	}
	tasks := []model.Task{}
	defer rows.Close()
	for rows.Next() {
		var task model.Task

		err := rows.Scan(&task.ID, &task.Name, &task.UserID)
		if err != nil {
			return []model.Task{}, fmt.Errorf("get all task: %w", err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}
