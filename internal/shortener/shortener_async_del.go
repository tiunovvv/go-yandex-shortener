package shortener

import (
	"context"

	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
	"go.uber.org/zap"
)

type Job struct {
	userID   string
	shortURL string
}

type Worker struct {
	jobQueue chan Job
	log      *zap.SugaredLogger
	store    storage.Store
	id       int
}

type Dispatcher struct {
	jobQueue   chan Job
	workerPool []*Worker
}

func NewWorker(id int, jobQueue chan Job, store storage.Store, log *zap.SugaredLogger) *Worker {
	return &Worker{
		id:       id,
		jobQueue: jobQueue,
		store:    store,
		log:      log,
	}
}

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

func (w *Worker) Start(ctx context.Context) {
	for job := range w.jobQueue {
		if err := w.store.SetDeletedFlag(ctx, job.userID, job.shortURL); err != nil {
			w.log.Errorf("failed to set deleted flag: %w", err)
		}
	}
}
