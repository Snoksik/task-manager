package workerpool

import (
	"context"
	"testing"
	"time"
)

func TestRunWorkerPool(t *testing.T) {
	jobs := []Job{
		{ID: 1, Value: 2},
		{ID: 2, Value: 3},
		{ID: 3, Value: 4},
		{ID: 4, Value: 5},
		{ID: 5, Value: 6},
	}
	result := RunWorkerPool(context.Background(), 2, jobs)
	if len(result) != 5 {
		t.Fatalf("expected len 5, got %d", len(result))
	}
	want := map[int]int{
		1: 4,
		2: 9,
		3: 16,
		4: 25,
		5: 36,
	}

	for _, r := range result {
		expected, ok := want[r.JobID]
		if !ok {
			t.Fatalf("unexpected job id %d", r.JobID)
		}

		if r.Value != expected {
			t.Fatalf("job %d: expected %d, got %d", r.JobID, expected, r.Value)
		}
	}
}

func TestRunWorkerPoolContext(t *testing.T) {
	jobs := make([]Job, 20)

	for i := range jobs {
		jobs[i] = Job{
			ID:    i + 1,
			Value: i + 2,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan []Result)

	go func() {
		done <- RunWorkerPool(ctx, 2, jobs)
	}()

	time.Sleep(250 * time.Millisecond)

	cancel()

	select {
	case result := <-done:
		if len(result) >= len(jobs) {
			t.Fatalf("expected partial result after cancellation, got %d", len(result))
		}

	case <-time.After(2 * time.Second):
		t.Fatal("worker pool did not stop after cancellation")
	}

}
