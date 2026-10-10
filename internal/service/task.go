package service

import (
	"context"
	"strings"

	"github.com/Snoksik/task-manager/internal/model"
)

type TaskRepository interface {
	Create(ctx context.Context, task model.Task) (model.Task, error)
	Update(ctx context.Context, id int, task model.Task) (model.Task, error)
	GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error)
	GetByID(ctx context.Context, id int) (model.Task, error)
	Delete(ctx context.Context, id int) error
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{
		repo: repo,
	}
}

func (s *TaskService) Create(ctx context.Context, task model.Task) (model.Task, error) {
	name, err := normalizeTaskName(task.Name)
	if err != nil {
		return model.Task{}, err
	}

	task.Name = name
	result, err := s.repo.Create(ctx, task)
	if err != nil {
		return model.Task{}, err
	}
	return result, nil
}

func (s *TaskService) Update(ctx context.Context, id int, task model.Task) (model.Task, error) {
	name, err := normalizeTaskName(task.Name)
	if err != nil {
		return model.Task{}, err
	}
	task.Name = name
	return s.repo.Update(ctx, id, task)
}

func (s *TaskService) GetAll(ctx context.Context, filter model.TaskFilter) ([]model.Task, error) {
	if filter.Limit < 1 || filter.Limit > 100 || filter.Offset < 0 {
		return []model.Task{}, model.ErrInvalidTaskFilter
	}
	return s.repo.GetAll(ctx, filter)
}

func (s *TaskService) GetByID(ctx context.Context, id int) (model.Task, error) {
	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func normalizeTaskName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", model.ErrInvalidTaskName
	}
	return name, nil
}
