package repository

import (
	"strings"
	"sync"

	"github.com/Snoksik/task-manager/internal/model"
)

type TaskRepository struct {
	tasks  []model.Task
	nextID int
	mu     sync.RWMutex
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: []model.Task{
			{ID: 1, Name: "Изучить net/http", UserID: 1},
			{ID: 2, Name: "Сделать REST API", UserID: 2},
			{ID: 3, Name: "Сделать REST ", UserID: 3},
		},
		nextID: 4,
	}

}

func (r *TaskRepository) GetByID(id int) (model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, err := r.findIndex(id)
	if err != nil {
		return model.Task{}, err
	}

	return r.tasks[i], nil
}

func search(str string, r []model.Task) []model.Task {
	result := make([]model.Task, 0)

	for _, task := range r {
		if strings.Contains(task.Name, str) {
			result = append(result, task)
		}
	}

	return result
}

func (r *TaskRepository) GetAll(filter model.TaskFilter) ([]model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	searchResult := search(filter.Search, r.tasks)
	start := filter.Offset
	end := filter.Limit + filter.Offset

	if start > len(searchResult) {
		return []model.Task{}, nil
	}
	if end >= len(searchResult) {
		end = len(searchResult)
	}

	result := searchResult[start:end]

	return result, nil
}

func (r *TaskRepository) Create(task model.Task) (model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task.ID = r.nextID

	r.nextID++

	r.tasks = append(r.tasks, task)

	return task, nil

}

func (r *TaskRepository) Update(id int, task model.Task) (model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, err := r.findIndex(id)
	if err != nil {
		return model.Task{}, err
	}

	task.ID = id
	r.tasks[i] = task
	return task, nil
}

// лишний мьютекст - дедлок
func (r *TaskRepository) findIndex(id int) (int, error) {
	for i, task := range r.tasks {
		if id == task.ID {
			return i, nil
		}
	}
	return -1, model.ErrTaskNotFound
}

func (r *TaskRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	i, err := r.findIndex(id)
	if err != nil {
		return err
	}
	r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
	return nil
}
