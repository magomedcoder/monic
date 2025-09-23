package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/magomedcoder/monic/internal/monic-agent/config"
	ports2 "github.com/magomedcoder/monic/internal/monic-agent/ports"
	"log"
	"time"
)

type App struct {
	cfg     config.Config
	host    string
	jrnl    ports2.JournalReader
	parser  ports2.EventParser
	webhook ports2.WebhookSender
}

func New(cfg config.Config, host string, jrnl ports2.JournalReader, parser ports2.EventParser, webhook ports2.WebhookSender) *App {
	return &App{
		cfg:     cfg,
		host:    host,
		jrnl:    jrnl,
		parser:  parser,
		webhook: webhook,
	}
}

func (a *App) Run(ctx context.Context) error {
	if err := a.jrnl.Init(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Monic] stopping")
			return nil
		default:
		}

		entry, err := a.jrnl.Next()
		if err != nil {
			log.Printf("journal next: %v", err)
			time.Sleep(time.Second)
			continue
		}

		if entry == nil {
			if err := a.jrnl.Wait(); err != nil {
				return err
			}
			continue
		}

		if entry.Message == "" {
			continue
		}

		if ev := a.parser.Parse(entry.Message); ev != nil {
			ev.Server = a.host
			ev.TS = entry.TS.UTC()
			ev.Raw = entry.Message

			out, _ := json.Marshal(ev)
			fmt.Println(string(out))

			if a.cfg.WebhookURL != "" {
				if err := a.webhook.Send(ctx, out); err != nil {
					log.Printf("webhook error: %v", err)
				}
			}
		}

		if cur := entry.Cursor; cur != "" {
			_ = a.jrnl.SaveCursor(cur)
		}
	}
}
