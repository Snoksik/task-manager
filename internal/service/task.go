package service

import (
	"strings"

	"github.com/Snoksik/task-manager/internal/model"
)

type TaskRepository interface {
	Create(task model.Task) model.Task
	Update(id int, task model.Task) (model.Task, error)
	GetAll(filter model.TaskFilter) ([]model.Task, error)
	GetByID(id int) (model.Task, error)
	Delete(id int) error
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{
		repo: repo,
	}
}

func (s *TaskService) Create(task model.Task) (model.Task, error) {
	name, err := normalizeTaskName(task.Name)
	if err != nil {
		return model.Task{}, err
	}

	task.Name = name
	return s.repo.Create(task), nil
}

func (s *TaskService) Update(id int, task model.Task) (model.Task, error) {
	name, err := normalizeTaskName(task.Name)
	if err != nil {
		return model.Task{}, err
	}
	task.Name = name
	return s.repo.Update(id, task)
}

func (s *TaskService) GetAll(filter model.TaskFilter) ([]model.Task, error) {
	if filter.Limit < 1 || filter.Limit > 100 || filter.Offset < 0 {
		return []model.Task{}, model.ErrInvalidTaskFilter
	}
	return s.repo.GetAll(filter)
}

func (s *TaskService) GetByID(id int) (model.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (s *TaskService) Delete(id int) error {
	err := s.repo.Delete(id)
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
