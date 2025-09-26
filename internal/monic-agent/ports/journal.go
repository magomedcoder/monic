package ports

import "time"

type JournalEntry struct {
	Message  string
	DateTime time.Time
	Cursor   string
}

type JournalReader interface {
	Init() error

	Next() (*JournalEntry, error)

	Wait() error

	SaveCursor(cur string) error

	Close() error
}
