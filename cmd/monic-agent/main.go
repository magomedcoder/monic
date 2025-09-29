package main

import (
	"context"
	"github.com/magomedcoder/monic/internal/monic-agent/app"
	"github.com/magomedcoder/monic/internal/monic-agent/config"
	grpcSender "github.com/magomedcoder/monic/internal/monic-agent/infra/grpc"
	httpInfra "github.com/magomedcoder/monic/internal/monic-agent/infra/http"
	"github.com/magomedcoder/monic/internal/monic-agent/infra/journal"
	"github.com/magomedcoder/monic/internal/monic-agent/infra/parser"
	"github.com/magomedcoder/monic/internal/monic-agent/infra/storage"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(0)
	cfg := config.Load()

	host, _ := os.Hostname()
	log.Printf("[Monic] start on %s, unit=%s", host, cfg.JournalMatchUnit)

	cursorStore := storage.NewCursorStore(cfg.StateDir)
	jrnl, err := journal.NewSystemdReader(cfg.JournalMatchUnit, cursorStore)
	if err != nil {
		log.Fatalf("open journal: %v", err)
	}
	defer jrnl.Close()

	var senderIface ports.EventSender
	if cfg.GRPCAddress != "" {
		senderIface, err = grpcSender.NewGRPCSender(cfg.GRPCAddress, cfg.Secret, cfg.GRPCInsecure)
		if err != nil {
			log.Fatalf("grpc sender: %v", err)
		}
		log.Printf("[Monic] gRPC -> %s (insecure=%v)", cfg.GRPCAddress, cfg.GRPCInsecure)
	} else {
		senderIface = httpInfra.NewHTTPSender(cfg.WebhookURL, cfg.Secret)
		log.Printf("[Monic] HTTP webhook -> %s", cfg.WebhookURL)
	}

	p := parser.NewChain(
		parser.NewSSHDRegexParser(),
		parser.NewNetfilterRegexParser(),
	)

	application := app.New(cfg, host, jrnl, p, senderIface)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}
