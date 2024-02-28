package shortener

import (
	"context"

	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
	"go.uber.org/zap"
)

// Job need userID and shortURL to set deleted flag.
type Job struct {
	userID   string
	shortURL string
}

// Worker processes jobs from jobQueue using store.
type Worker struct {
	jobQueue chan Job
	log      *zap.SugaredLogger
	store    storage.Store
	id       int
}

// Dispatcher distributes jobs using workerPool.
type Dispatcher struct {
	jobQueue   chan Job
	workerPool []*Worker
}

// NewWorker creates and returns new Worker.
func NewWorker(id int, jobQueue chan Job, store storage.Store, log *zap.SugaredLogger) *Worker {
	return &Worker{
		id:       id,
		jobQueue: jobQueue,
		store:    store,
		log:      log,
	}
}

// NewDispatcher creates and returns new Dispatcher that distributes jobs to workers.
func NewDispatcher(
	ctx context.Context,
	workerCount int,
	jobQueue chan Job,
	store storage.Store,
	log *zap.SugaredLogger) *Dispatcher {
	workerPool := make([]*Worker, workerCount)
	for i := 0; i < workerCount; i++ {
		worker := NewWorker(i, jobQueue, store, log)
		workerPool[i] = worker
		go worker.Start(ctx)
	}

	return &Dispatcher{
		workerPool: workerPool,
		jobQueue:   jobQueue,
	}
}

// Start gets jobs from jobQueue and sends them to workers.
func (w *Worker) Start(ctx context.Context) {
	for job := range w.jobQueue {
		if err := w.store.SetDeletedFlag(ctx, job.userID, job.shortURL); err != nil {
			w.log.Errorf("failed to set deleted flag: %w", err)
		}
	}
}
