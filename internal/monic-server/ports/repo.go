package ports

import (
	"context"
	"github.com/magomedcoder/monic/internal/monic-server/domain"
)

type Repo interface {
	Ensure(ctx context.Context) error

	InsertBatch(ctx context.Context, batch []domain.IngestedEvent) error

	Close() error
}
