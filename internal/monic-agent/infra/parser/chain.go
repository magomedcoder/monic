package parser

import (
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
)

type chainParser struct {
	parsers []ports.EventParser
}

func NewChain(parsers ...ports.EventParser) ports.EventParser {
	return &chainParser{parsers: parsers}
}

func (c *chainParser) Parse(msg string) *domain.Event {
	for _, p := range c.parsers {
		if ev := p.Parse(msg); ev != nil {
			return ev
		}
	}

	return nil
}
