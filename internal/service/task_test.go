package service

import (
	"context"
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
	receivedCtx  context.Context
}

func (s *taskRepositorySpy) Create(ctx context.Context, task model.Task) (model.Task, error) {
	s.createCalls++
	s.createInput = task
	s.receivedCtx = ctx
	return s.createResult, nil
}

func TestCreateInvalidName(t *testing.T) {
	ctx := context.Background()
	repo := &taskRepositorySpy{}
	svc := NewTaskService(repo)

	_, err := svc.Create(ctx, model.Task{Name: "   "})

	if !errors.Is(err, model.ErrInvalidTaskName) {
		t.Errorf("expected ErrInvalidTaskName, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected calls repository 0, got %d", repo.createCalls)
	}

}

func TestCreateValidName(t *testing.T) {
	ctx := context.Background()
	repo := &taskRepositorySpy{
		createResult: model.Task{
			ID:   42,
			Name: "Купить хлеб",
		},
	}
	svc := NewTaskService(repo)

	created, err := svc.Create(ctx, model.Task{
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

func TestTaskServicePassesContextToRepository(t *testing.T) {
	repo := &taskRepositorySpy{}
	svc := NewTaskService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := svc.Create(
		ctx,
		model.Task{Name: "test"},
	)
	if err != nil {
		t.Fatalf("Expected zero error, got %v", err)
	}
	if repo.receivedCtx != ctx {
		t.Fatalf("expected context %v, got %v", ctx, repo.receivedCtx)
	}
}
