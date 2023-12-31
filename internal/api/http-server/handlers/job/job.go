package job

import (
	"time"
)

type Job interface {
	Execute()
}

type WorkerPool struct {
	workers []Worker
}

type Worker struct {
	jobQueue JobQueue
}

type JobQueue chan Job

var Queue JobQueue

func Dispatch(job Job, delay time.Duration) {
	go func() {
		<-time.After(delay)
		Queue <- job
	}()
}

func NewWorkerPool(size int, queue JobQueue) *WorkerPool {
	workers := make([]Worker, size)
	for i := 0; i < size; i++ {
		workers[i] = NewWorker(queue)
	}
	return &WorkerPool{workers}
}

func (p *WorkerPool) Start() {
	for _, worker := range p.workers {
		worker.Start()
	}
}

func NewWorker(jobQueue JobQueue) Worker {
	return Worker{jobQueue}
}

func (w *Worker) Start() {
	go func() {
		for job := range w.jobQueue {
			job.Execute()
		}
	}()
}
