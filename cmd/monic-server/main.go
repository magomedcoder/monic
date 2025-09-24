package main

import (
	"context"
	"github.com/magomedcoder/monic/internal/monic-server/app"
	"github.com/magomedcoder/monic/internal/monic-server/config"
	"github.com/magomedcoder/monic/internal/monic-server/infra/auth"
	httpInfra "github.com/magomedcoder/monic/internal/monic-server/infra/http"
	"github.com/magomedcoder/monic/internal/monic-server/infra/repo"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(0)
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	verifier := auth.NewHMACVerifier(cfg.SharedSecret)
	store, err := repo.NewClickHouse(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer store.Close()

	application := app.New(cfg, store)
	go application.RunInserter(ctx)

	srv := httpInfra.NewServer(cfg, verifier, application)
	if err := srv.Start(ctx); err != nil {
		panic(err)
	}
}
