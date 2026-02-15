package scheduler

import (
	"context"
	//"github.com/evok02/jcrawler/internal/worker"
	//"github.com/evok02/jcrawler/internal/parser"
	//"github.com/evok02/jcrawler/internal/filter"
	//"github.com/evok02/jcrawler/internal/db"
)

type JobQueue struct {
	queue chan string
}

func NewJobQueue(n int) *JobQueue {
	return &JobQueue{
		queue: make(chan string, n),
	}
}

func (jq *JobQueue)Pop(ctx context.Context, n int) chan string {
	res := make(chan string, n)
	go func() {
		for range n {
			res <- <- jq.queue
		}
		close(res)
	}()
	return res
}

func (jq *JobQueue) Push(n int, in <- chan string) {
	go func() {
		for range n {
			jq.queue <- <- in
		}
	}()
}
