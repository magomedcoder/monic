package ports

import (
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
)

type EventParser interface {
	Parse(msg string) *domain.Event
}
