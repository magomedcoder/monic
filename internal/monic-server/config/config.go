package config

import (
	"github.com/magomedcoder/monic/pkg"
	"os"
)

type Config struct {
	Addr          string
	GRPCAddr      string
	TLSCertFile   string
	TLSKeyFile    string
	Secret        string
	ClickHouseDSN string
	BatchSize     int
	BatchWindowMs int
}

func Load() Config {
	cfg := Config{
		Addr:          pkg.GetEnv("MONIC_SERVER_HTTP_ADDR", ":8000"),
		GRPCAddr:      pkg.GetEnv("MONIC_SERVER_GRPC_ADDR", ""),
		TLSCertFile:   pkg.GetEnv("MONIC_SERVER_TLS_CERT", ""),
		TLSKeyFile:    pkg.GetEnv("MONIC_SERVER_TLS_KEY", ""),
		Secret:        os.Getenv("MONIC_SECRET"),
		ClickHouseDSN: pkg.GetEnv("MONIC_SERVER_CLICKHOUSE_DSN", "tcp://127.0.0.1:9000?database=monic_db"),
		BatchSize:     pkg.MustInt(pkg.GetEnv("MONIC_SERVER_BATCH_SIZE", "500"), 500),
		BatchWindowMs: pkg.MustInt(pkg.GetEnv("MONIC_SERVER_BATCH_WINDOW_MS", "500"), 500),
	}

	return cfg
}
