package scheduler

import (
	"context"
)

type JobQueue struct {
	queue chan string
}

func NewJobQueue(n int) *JobQueue {
	return &JobQueue{
		queue: make(chan string, n),
	}
}

func (jq *JobQueue) Pop(ctx context.Context, n int) chan string {
	res := make(chan string, n)
	go func() {
	outer:
		for range n {
			select {
			case res <- <-jq.queue:
			case <-ctx.Done():
				break outer
			}
		}
		close(res)
	}()
	return res
}

func (jq *JobQueue) Push(ctx context.Context, n int, in <-chan string) {
	go func() {
	outer:
		for range n {
			select {
			case jq.queue <- <-in:
			case <-ctx.Done():
				break outer
			}
		}
	}()
}
