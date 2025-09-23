package main

import (
	"context"
	"github.com/magomedcoder/monic/internal/app"
	"github.com/magomedcoder/monic/internal/config"
	"github.com/magomedcoder/monic/internal/infra/journal"
	"github.com/magomedcoder/monic/internal/infra/parser"
	"github.com/magomedcoder/monic/internal/infra/storage"
	"github.com/magomedcoder/monic/internal/infra/webhook"
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

	p := parser.NewSSHDRegexParser()
	wh := webhook.NewHTTPSender(cfg.WebhookURL, cfg.SharedSecret)

	application := app.New(cfg, host, jrnl, p, wh)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := application.Run(ctx); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}
