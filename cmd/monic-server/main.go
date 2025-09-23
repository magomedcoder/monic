package main

import (
	"context"
	"github.com/magomedcoder/monic/internal/monic-server/config"
	"github.com/magomedcoder/monic/internal/monic-server/infra/auth"
	httpInfra "github.com/magomedcoder/monic/internal/monic-server/infra/http"
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

	srv := httpInfra.NewServer(cfg, verifier)
	if err := srv.Start(ctx); err != nil {
		panic(err)
	}
}
