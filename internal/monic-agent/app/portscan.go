package app

import (
	"sync"
	"time"

	"github.com/magomedcoder/monic/internal/monic-agent/domain"
)

type PortscanDetector struct {
	window time.Duration
	limit  int
	mu     sync.Mutex
	seen   map[string]map[string]time.Time
}

func NewPortscanDetector(window time.Duration, limit int) *PortscanDetector {
	return &PortscanDetector{
		window: window,
		limit:  limit,
		seen:   make(map[string]map[string]time.Time),
	}
}

func (d *PortscanDetector) Feed(now time.Time, ev *domain.Event) *domain.Event {
	if ev == nil || ev.Type != "net_probe" || ev.RemoteIP == "" || ev.Port == "" {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	m, ok := d.seen[ev.RemoteIP]
	if !ok {
		m = make(map[string]time.Time)
		d.seen[ev.RemoteIP] = m
	}

	m[ev.Port] = now

	cutoff := now.Add(-d.window)
	for p, ts := range m {
		if ts.Before(cutoff) {
			delete(m, p)
		}
	}

	if len(m) >= d.limit {
		delete(d.seen, ev.RemoteIP)

		return &domain.Event{
			DateTime: now.UTC(),
			Server:   "",
			Type:     "port_scan",
			User:     "",
			RemoteIP: ev.RemoteIP,
			Port:     "",
			Method:   ev.Method,
			Message:  "portscan_hit",
			Raw:      "",
		}
	}

	return nil
}
