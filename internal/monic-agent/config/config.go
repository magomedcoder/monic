package config

import (
	"github.com/magomedcoder/monic/pkg"
	"os"
	"strings"
)

type Config struct {
	DebugMode             bool
	WebhookURL            string
	SharedSecret          string
	GRPCAddress           string
	GRPCInsecure          bool
	JournalMatchUnit      string
	EnablePortscan        bool
	PortscanWindowSeconds int
	PortscanDistinctPorts int
	EnableSnifferPromisc  bool
	SnifferIfaces         []string
	SnifferBPF            string
	StateDir              string
}

func Load() Config {
	cfg := Config{
		DebugMode:             pkg.GetEnv("MONIC_DEBUG", "false") == "true",
		WebhookURL:            os.Getenv("MONIC_HTTP_URL"),
		SharedSecret:          os.Getenv("MONIC_SHARED_SECRET"),
		GRPCAddress:           os.Getenv("MONIC_GRPC_ADDR"),
		GRPCInsecure:          pkg.GetEnv("MONIC_GRPC_INSECURE", "false") == "true",
		JournalMatchUnit:      pkg.GetEnv("MONIC_JOURNAL_UNIT", "sshd.service"),
		EnablePortscan:        pkg.GetEnv("MONIC_ENABLE_PORTSCAN", "false") == "true",
		EnableSnifferPromisc:  pkg.GetEnv("MONIC_SNIFFER_PROMISC", "false") == "true",
		SnifferBPF:            pkg.GetEnv("MONIC_SNIFFER_BPF", ""),
		PortscanWindowSeconds: pkg.MustInt(pkg.GetEnv("MONIC_PORTSCAN_WINDOW_SECONDS", "10"), 10),
		PortscanDistinctPorts: pkg.MustInt(pkg.GetEnv("MONIC_PORTSCAN_DISTINCT_PORTS", "12"), 12),
		StateDir:              "/var/lib/monic-agent",
	}

	ifaces := pkg.GetEnv("MONIC_SNIFFER_IFACES", "")
	var list []string
	if ifaces != "" {
		for _, s := range strings.Split(ifaces, ",") {
			if t := strings.TrimSpace(s); t != "" {
				list = append(list, t)
			}
		}
	}

	_ = os.MkdirAll(cfg.StateDir, 0o755)

	return cfg
}
