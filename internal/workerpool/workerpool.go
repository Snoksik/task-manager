package workerpool

import (
	"context"
	"sync"
	"time"
)

type Job struct {
	ID    int
	Value int
}

type Result struct {
	JobID int
	Value int
}

func RunWorkerPool(ctx context.Context, workers int, jobs []Job) []Result {
	jobsChan := make(chan Job)
	resultChan := make(chan Result)
	wg := sync.WaitGroup{}

	wg.Add(workers)

	go func() {
		defer close(jobsChan)

		for _, x := range jobs {
			select {
			case jobsChan <- x:
			case <-ctx.Done():
				return
			}
		}
	}()

	for range workers {
		go worker(ctx, jobsChan, resultChan, &wg)

	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()
	res := make([]Result, 0)

	for x := range resultChan {
		res = append(res, x)
	}

	return res
}

func worker(ctx context.Context, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case job, ok := <-jobs:
			if !ok {
				return
			}
			time.Sleep(100 * time.Millisecond)

			result := Result{
				JobID: job.ID,
				Value: job.Value * job.Value,
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
