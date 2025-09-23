package config

import (
	"github.com/magomedcoder/monic/pkg"
	"os"
)

type Config struct {
	Addr          string
	SharedSecret  string
	BatchSize     int
	BatchWindowMs int
}

func Load() Config {
	cfg := Config{
		Addr:          pkg.GetEnv("MONIC_SERVER_ADDR", ":8000"),
		SharedSecret:  os.Getenv("MONIC_SERVER_SHARED_SECRET"),
		BatchSize:     pkg.MustInt(pkg.GetEnv("MONIC_SERVER_BATCH_SIZE", "500"), 500),
		BatchWindowMs: pkg.MustInt(pkg.GetEnv("MONIC_SERVER_BATCH_WINDOW_MS", "500"), 500),
	}

	return cfg
}
