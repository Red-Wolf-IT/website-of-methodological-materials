package worker

import (
	"context"
	"log"

	"github.com/google/uuid"
)

type ViewsRepository interface {
	IncrementViews(ctx context.Context, id uuid.UUID) error
}

// ViewsWorker читает события просмотра из канала и обновляет счётчик в БД
type ViewsWorker struct {
	viewsChan chan uuid.UUID
	repo      ViewsRepository
	done      chan struct{}
}

func NewViewsWorker(repo ViewsRepository, bufferSize int) *ViewsWorker {
	return &ViewsWorker{
		viewsChan: make(chan uuid.UUID, bufferSize),
		repo:      repo,
		done:      make(chan struct{}),
	}
}

func (w *ViewsWorker) ViewsChan() chan<- uuid.UUID {
	return w.viewsChan
}

func (w *ViewsWorker) Run(ctx context.Context) {
	defer close(w.done)

	for {
		select {
		case <-ctx.Done():
			w.drain()
			return
		case id := <-w.viewsChan:
			w.increment(ctx, id)
		}
	}
}

// Wait блокируется до завершения worker (после отмены context)
func (w *ViewsWorker) Wait() {
	<-w.done
}

func (w *ViewsWorker) drain() {
	ctx := context.Background()
	for {
		select {
		case id := <-w.viewsChan:
			w.increment(ctx, id)
		default:
			return
		}
	}
}

func (w *ViewsWorker) increment(ctx context.Context, id uuid.UUID) {
	if err := w.repo.IncrementViews(ctx, id); err != nil {
		log.Printf("views worker: increment manual %s: %v", id, err)
	}
}
