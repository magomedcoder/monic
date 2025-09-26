package ports

import (
	"context"
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
)

type EventSender interface {
	Send(ctx context.Context, ev *domain.Event) error

	Close() error
}
