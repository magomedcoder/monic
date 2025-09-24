package app

import (
	"context"
	"errors"
	"github.com/magomedcoder/monic/internal/monic-server/config"
	"github.com/magomedcoder/monic/internal/monic-server/domain"
	"github.com/magomedcoder/monic/internal/monic-server/ports"
	"log"
	"strconv"
	"time"
)

type App struct {
	cfg   config.Config
	repo  ports.Repo
	queue chan domain.IngestedEvent
}

func New(cfg config.Config, repo ports.Repo) *App {
	q := make(chan domain.IngestedEvent, cfg.BatchSize*4)

	return &App{
		cfg:   cfg,
		repo:  repo,
		queue: q,
	}
}

func (a *App) Enqueue(e domain.Event) error {
	select {
	case a.queue <- domain.IngestedEvent{
		Event:      e,
		ReceivedAt: time.Now().UTC(),
	}:
		return nil
	default:
		return errors.New("queue full")
	}
}

func (a *App) RunInserter(ctx context.Context) {
	if err := a.repo.Ensure(ctx); err != nil {
		log.Fatalf("ensure table: %v", err)
	}

	batch := make([]domain.IngestedEvent, 0, a.cfg.BatchSize)
	ticker := time.NewTicker(time.Duration(a.cfg.BatchWindowMs) * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := a.repo.InsertBatch(ctx, batch); err != nil {
			log.Printf("insert error len=%d: %v", len(batch), err)
			time.Sleep(time.Second)
			_ = a.repo.InsertBatch(ctx, batch)
		}

		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case ev := <-a.queue:
			batch = append(batch, ev)
			if len(batch) >= a.cfg.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func ParsePort(s string) uint16 {
	p, err := strconv.Atoi(s)
	if err != nil || p < 0 || p > 65535 {
		return 0
	}

	return uint16(p)
}
