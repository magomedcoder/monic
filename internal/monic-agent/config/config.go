package config

import (
	"github.com/magomedcoder/monic/pkg"
	"os"
)

type Config struct {
	WebhookURL       string `json:"webhook_url"`
	SharedSecret     string `json:"shared_secret"`
	GRPCAddress      string `json:"grpc_address"`
	GRPCInsecure     bool   `json:"grpc_insecure"`
	JournalMatchUnit string `json:"journal_match_unit"`
	StateDir         string `json:"state_dir"`
}

func Load() Config {
	cfg := Config{
		WebhookURL:       os.Getenv("MONIC_HTTP_URL"),
		SharedSecret:     os.Getenv("MONIC_SHARED_SECRET"),
		GRPCAddress:      os.Getenv("MONIC_GRPC_ADDR"),
		GRPCInsecure:     pkg.GetEnv("MONIC_GRPC_INSECURE", "false") == "true",
		JournalMatchUnit: pkg.GetEnv("MONIC_JOURNAL_UNIT", "sshd.service"),
		StateDir:         "/var/lib/monic-agent",
	}
	_ = os.MkdirAll(cfg.StateDir, 0o755)

	return cfg
}
