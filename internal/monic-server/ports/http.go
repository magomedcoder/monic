package ports

import "github.com/magomedcoder/monic/internal/monic-server/domain"

type Enqueuer interface {
	Enqueue(e domain.Event) error
}
