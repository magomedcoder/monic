package config

import (
	"github.com/magomedcoder/monic/pkg"
	"os"
)

type Config struct {
	WebhookURL       string `json:"webhook_url"`
	SharedSecret     string `json:"shared_secret"`
	JournalMatchUnit string `json:"journal_match_unit"`
	StateDir         string `json:"state_dir"`
}

func Load() Config {
	cfg := Config{
		WebhookURL:       os.Getenv("MONIC_WEBHOOK_URL"),
		SharedSecret:     os.Getenv("MONIC_SHARED_SECRET"),
		JournalMatchUnit: pkg.GetEnv("MONIC_JOURNAL_UNIT", "sshd.service"),
		StateDir:         pkg.GetEnv("MONIC_STATE_DIR", "/var/lib/monic-agent"),
	}
	_ = os.MkdirAll(cfg.StateDir, 0o755)

	return cfg
}
