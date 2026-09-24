package repository

import (
	"errors"
	"sync"
	"testing"

	"github.com/Snoksik/task-manager/internal/model"
)

func TestGetByID(t *testing.T) {
	repo := NewTaskRepository()

	task, err := repo.GetByID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}

	if task.Name != "Изучить net/http" {
		t.Errorf("unexpected task name: %q", task.Name)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	repo := NewTaskRepository()

	_, err := repo.GetByID(999)

	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestCreateAfterDelete(t *testing.T) {
	repo := NewTaskRepository()

	if err := repo.Delete(1); err != nil {
		t.Fatalf("failed to prepare test: %v", err)
	}

	created := repo.Create(model.Task{Name: "Новая задача"})

	if created.ID != 3 {
		t.Errorf("expected new task ID 3, got %d", created.ID)
	}
	task, err := repo.GetByID(2)
	if err != nil {
		t.Fatalf("expected no error for task 2, got %v", err)
	}
	if task.ID != 2 {
		t.Errorf("Dont correct id")
	}
	if task.Name != "Сделать REST API" {
		t.Errorf("Dont correct name")
	}
}

func TestUpdate(t *testing.T) {
	repo := NewTaskRepository()

	updated, err := repo.Update(1, model.Task{
		ID:   999,
		Name: "Обновлённая задача",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.ID != 1 {
		t.Errorf("dont correct id, need 1, return %d", updated.ID)
	}

	if updated.Name != "Обновлённая задача" {
		t.Errorf("dont correct name, need 'Обновлённая задача', return %s", updated.Name)
	}

	saved, err := repo.GetByID(1)
	if err != nil {
		t.Fatalf("failed to read updated task: %v", err)
	}

	if saved.ID != 1 {
		t.Errorf("dont correct id, need 1, return %d", saved.ID)
	}

	if saved.Name != "Обновлённая задача" {
		t.Errorf("dont correct name, need 'Обновлённая задача', return %s", saved.Name)
	}
}

func TestDelete(t *testing.T) {
	repo := NewTaskRepository()

	err := repo.Delete(1)

	if err != nil {
		t.Fatalf("failed delet response %v", err)
	}

	_, err = repo.GetByID(1)

	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound after deletion, got %v", err)
	}

	task, err := repo.GetByID(2)

	if err != nil {
		t.Fatalf("expected not errors, got %v", err)
	}

	if task.ID != 2 {
		t.Errorf("dont correct id, need 2, return %d", task.ID)
	}

	if task.Name != "Сделать REST API" {
		t.Errorf("dont correct name, need 'Сделать REST API', return %s", task.Name)
	}

	err = repo.Delete(1)

	if !errors.Is(err, model.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound after deletion, got %v", err)
	}
}

func TestConcurrentCreate(t *testing.T) {
	repo := NewTaskRepository()

	wg := sync.WaitGroup{}

	for range 100 {
		wg.Go(func() {
			repo.Create(model.Task{})
		})
	}

	wg.Wait()
	result := repo.GetAll()
	m := make(map[int]struct{})

	for i := range result {
		m[result[i].ID] = struct{}{}
	}

	if len(m) != 102 {
		t.Fatalf("expected 102 unicum id, got %d", len(m))
	}
	if len(result) != 102 {
		t.Fatalf("expected 102, got %d", len(result))
	}

}

func TestGetAllReturnsCopy(t *testing.T) {
	repo := NewTaskRepository()

	result := repo.GetAll()

	if len(result) == 0 {
		t.Fatalf("expected tasks > 0, got %d", len(result))
	}
	id, name := result[0].ID, result[0].Name

	result[0].Name = "another"

	r, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("expected zero error, got %v", err)
	}
	if r.ID != id {
		t.Errorf("expected %d id, got %d", id, r.ID)
	}
	if r.Name != name {
		t.Errorf("expected  %s name, got %s", name, r.Name)
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	repo := NewTaskRepository()

	wg := sync.WaitGroup{}

	for range 100 {
		wg.Go(func() {
			repo.Create(model.Task{})
		})
		wg.Go(func() {
			result := repo.GetAll()
			for i := range result {
				if result[i].ID <= 0 {
					t.Errorf("expected id > 0, got %d", result[i].ID)
				}
			}
		})
	}

	wg.Wait()

	result := repo.GetAll()
	if len(result) != 102 {
		t.Fatalf("expected 102, got %d", len(result))
	}
}
