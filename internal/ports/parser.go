package ports

import "github.com/magomedcoder/monic/internal/domain"

type EventParser interface {
	Parse(msg string) *domain.Event
}
