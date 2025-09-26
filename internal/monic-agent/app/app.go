package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/magomedcoder/monic/internal/monic-agent/config"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
	"log"
	"time"
)

type App struct {
	cfg    config.Config
	host   string
	jrnl   ports.JournalReader
	parser ports.EventParser
	sender ports.EventSender
}

func New(cfg config.Config, host string, jrnl ports.JournalReader, parser ports.EventParser, sender ports.EventSender) *App {
	return &App{
		cfg:    cfg,
		host:   host,
		jrnl:   jrnl,
		parser: parser,
		sender: sender,
	}
}

func (a *App) Run(ctx context.Context) error {
	defer a.sender.Close()

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
			ev.DateTime = entry.DateTime.UTC()
			ev.Raw = entry.Message

			out, _ := json.Marshal(ev)
			fmt.Println(string(out))

			if err := a.sender.Send(ctx, ev); err != nil {
				log.Printf("send error: %v", err)
			}
		}

		if cur := entry.Cursor; cur != "" {
			_ = a.jrnl.SaveCursor(cur)
		}
	}
}
