package ports

import (
	"context"
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
)

type ProbesSource interface {
	Start(ctx context.Context) (<-chan *domain.Event, error)

	Close() error
}
