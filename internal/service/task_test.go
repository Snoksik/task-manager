package service

import (
	"errors"
	"testing"

	"github.com/Snoksik/task-manager/internal/model"
)

func TestNormalizeTaskName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "without surrounding spaces",
			input: "Купить хлеб",
			want:  "Купить хлеб",
		},
		{
			name:  "trim surrounding spaces",
			input: "  Купить хлеб  ",
			want:  "Купить хлеб",
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: model.ErrInvalidTaskName,
		},
		{
			name:    "empty name with 3 space",
			input:   "   ",
			wantErr: model.ErrInvalidTaskName,
		},
		{
			name:    "empty name with '\\t\\n' ",
			input:   "\t\n",
			wantErr: model.ErrInvalidTaskName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTaskName(tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if got != tt.want {
				t.Errorf("expected name %q, got %q", tt.want, got)
			}
		})
	}
}

type taskRepositorySpy struct {
	TaskRepository
	createCalls  int
	createInput  model.Task
	createResult model.Task
}

func (s *taskRepositorySpy) Create(task model.Task) model.Task {
	s.createCalls++
	s.createInput = task
	return s.createResult
}

func TestCreateInvalidName(t *testing.T) {
	repo := &taskRepositorySpy{}
	svc := NewTaskService(repo)

	_, err := svc.Create(model.Task{Name: "   "})

	if !errors.Is(err, model.ErrInvalidTaskName) {
		t.Errorf("expected ErrInvalidTaskName, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected calls repository 0, got %d", repo.createCalls)
	}

}

func TestCreateValidName(t *testing.T) {
	repo := &taskRepositorySpy{
		createResult: model.Task{
			ID:   42,
			Name: "Купить хлеб",
		},
	}
	svc := NewTaskService(repo)

	created, err := svc.Create(model.Task{
		Name: "  Купить хлеб  ",
	})
	if err != nil {
		t.Fatalf("expected zero error, got %v", err)
	}
	if repo.createCalls != 1 {
		t.Errorf("expected calls repository 1, got %d", repo.createCalls)
	}
	if repo.createInput.Name != "Купить хлеб" {
		t.Errorf("expected  'Купить хлеб', got %s", repo.createInput.Name)
	}
	if created.ID != 42 {
		t.Errorf("expected id = 42, got %d", created.ID)
	}
	if created.Name != "Купить хлеб" {
		t.Errorf("expected  'Купить хлеб', got %s", created.Name)
	}
}
