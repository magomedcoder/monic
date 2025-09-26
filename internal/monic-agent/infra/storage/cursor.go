package storage

import (
	"os"
	"path/filepath"
	"strings"
)

type CursorStore struct {
	path string
}

func NewCursorStore(stateDir string) *CursorStore {
	_ = os.MkdirAll(stateDir, 0o755)
	return &CursorStore{
		path: filepath.Join(stateDir, "cursor"),
	}
}

func (c *CursorStore) Load() (string, error) {
	b, err := os.ReadFile(c.path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}

func (c *CursorStore) Save(cur string) error {
	return os.WriteFile(c.path, []byte(cur), 0o644)
}
