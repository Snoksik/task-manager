package repository

import (
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
		},
		nextID: 3,
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
func (r *TaskRepository) GetAll() []model.Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Task, len(r.tasks))
	copy(result, r.tasks)
	return result
}

func (r *TaskRepository) Create(task model.Task) model.Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	task.ID = r.nextID

	r.nextID++

	r.tasks = append(r.tasks, task)

	return task

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
