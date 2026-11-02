package pipeline

import (
	"errors"
	"sync"
)

var (
	ErrQueueFull = errors.New("queue is full, please try again later")
)

type Worker struct {
	queue        chan Task
	numOfWorkers int
	wg           sync.WaitGroup
}

func NewWorker(queueSize, numOfWorkers int) *Worker {
	return &Worker{
		queue:        make(chan Task, queueSize),
		numOfWorkers: numOfWorkers,
	}
}

func (w *Worker) Start() {

	for range w.numOfWorkers {

		w.wg.Go(func() {
			for t := range w.queue {
				t.Process()
			}
		})

	}

}

func (w *Worker) Stop() {
	close(w.queue)
	w.wg.Wait()
}

func (w *Worker) Enqueue(t Task) error {

	select {
	case w.queue <- t:
		return nil
	default:
		return ErrQueueFull
	}

}
