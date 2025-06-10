package workers

import (
	"context"
	"sync"
)

type WorkerPool struct {
	ctx          context.Context
	jobs         <-chan string
	workers      map[int]chan<- struct{}
	workersMutex sync.Mutex
}

func NewWorkerPool(ctx context.Context, jobs <-chan string) *WorkerPool {
	return &WorkerPool{
		ctx:     ctx,
		jobs:    jobs,
		workers: make(map[int]chan<- struct{}),
	}
}

func (w *WorkerPool) AddWorker(id int, worker func(string)) bool {
	w.workersMutex.Lock()

	_, ok := w.workers[id]
	if ok {
		w.workersMutex.Unlock()
		return false
	}

	term := make(chan struct{})
	w.workers[id] = term

	w.workersMutex.Unlock()

	go work(w.ctx, w.jobs, term, worker)
	return true
}

func (w *WorkerPool) RemoveWorker(id int) bool {
	w.workersMutex.Lock()

	term, ok := w.workers[id]
	if !ok {
		w.workersMutex.Unlock()
		return false
	}
	delete(w.workers, id)

	w.workersMutex.Unlock()

	close(term)
	return true
}

func (w *WorkerPool) Shutdown() {
	w.workersMutex.Lock()
	for id := range w.workers {
		w.workers[id] <- struct{}{}
		delete(w.workers, id)
	}
	w.workersMutex.Unlock()
}

func work(ctx context.Context, jobs <-chan string, term <-chan struct{}, w func(string)) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-term:
			return
		case str, ok := <-jobs:
			if !ok {
				return
			}
			w(str)
		}
	}
}
