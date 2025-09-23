package journal

import (
	"errors"
	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/magomedcoder/monic/internal/ports"
	"time"
)

type cursorStore interface {
	Load() (string, error)

	Save(cur string) error
}

type systemdReader struct {
	unit  string
	store cursorStore
	j     *sdjournal.Journal
}

func NewSystemdReader(unit string, store cursorStore) (ports.JournalReader, error) {
	j, err := sdjournal.NewJournal()
	if err != nil {
		return nil, err
	}

	return &systemdReader{
		unit:  unit,
		store: store,
		j:     j,
	}, nil
}

func (s *systemdReader) Init() error {
	if s.unit != "" {
		if err := s.j.AddMatch("_SYSTEMD_UNIT=" + s.unit); err != nil {
			return err
		}
	} else {
		if err := s.j.AddMatch("SYSLOG_IDENTIFIER=sshd"); err != nil {
			return err
		}
	}

	if cur, err := s.store.Load(); err == nil && cur != "" {
		if err := s.j.SeekCursor(cur); err == nil {
			_, _ = s.j.Next()
			return nil
		}
	}

	if err := s.j.SeekTail(); err == nil {
		_, _ = s.j.Next()
	}

	return nil
}

func (s *systemdReader) Next() (*ports.JournalEntry, error) {
	n, err := s.j.Next()
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, nil
	}

	e, err := s.j.GetEntry()
	if err != nil {
		return nil, err
	}

	msg := e.Fields["MESSAGE"]
	if msg == "" {
		return &ports.JournalEntry{
			Message: "",
			TS:      time.Time{},
			Cursor:  "",
		}, nil
	}

	cursor, _ := s.j.GetCursor()
	return &ports.JournalEntry{
		Message: msg,
		TS:      time.Unix(int64(e.RealtimeTimestamp/1e6), int64((e.RealtimeTimestamp%1e6)*1e3)),
		Cursor:  cursor,
	}, nil
}

func (s *systemdReader) Wait() error {
	s.j.Wait(sdjournal.IndefiniteWait)
	return nil
}

func (s *systemdReader) SaveCursor(cur string) error {
	if s.store == nil {
		return errors.New("no cursor store")
	}

	return s.store.Save(cur)
}

func (s *systemdReader) Close() error {
	return s.j.Close()
}
