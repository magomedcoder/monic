package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/magomedcoder/monic/internal/monic-server/config"
	"github.com/magomedcoder/monic/internal/monic-server/domain"
	"github.com/magomedcoder/monic/internal/monic-server/ports"
	"io"
	"log"
	"net/http"
)

type Server struct {
	cfg      config.Config
	verifier ports.Verifier
	app      ports.Enqueuer
}

func NewServer(cfg config.Config, v ports.Verifier, app ports.Enqueuer) *Server {
	return &Server{
		cfg:      cfg,
		verifier: v,
		app:      app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleWebhook)

	srv := &http.Server{
		Addr:    s.cfg.Addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	log.Printf("[Monic] listening on %s", s.cfg.Addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := s.verifier.Verify(r.Header.Get("X-Signature"), payload); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var ev domain.Event
	if err := json.Unmarshal(payload, &ev); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := s.app.Enqueue(ev); err != nil {
		http.Error(w, "queue full", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	_, _ = w.Write([]byte("OK"))
}
