package scheduler

import (
	"context"
)

type JobQueue struct {
	Queue chan string
}

func NewJobQueue(n int) *JobQueue {
	return &JobQueue{
		Queue: make(chan string, n),
	}
}

func (jq *JobQueue) Pop(ctx context.Context, n int) chan string {
	res := make(chan string, n)
	go func() {
	outer:
		for range n {
			select {
			case res <- <-jq.Queue:
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
			case jq.Queue <- <-in:
			case <-ctx.Done():
				break outer
			}
		}
	}()
}
