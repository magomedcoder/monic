package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/magomedcoder/monic/internal/monic-agent/infra/sniffer"
	"log"
	"time"

	"github.com/magomedcoder/monic/internal/monic-agent/config"
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
)

type App struct {
	cfg    config.Config
	host   string
	jrnl   ports.JournalReader
	parser ports.EventParser
	sender ports.EventSender
	probes ports.ProbesSource
	psd    *PortscanDetector
}

func New(
	cfg config.Config,
	host string,
	jrnl ports.JournalReader,
	parser ports.EventParser,
	sender ports.EventSender,
) *App {
	var det *PortscanDetector
	if cfg.EnablePortscan {
		det = NewPortscanDetector(time.Duration(cfg.PortscanWindowSeconds)*time.Second, cfg.PortscanDistinctPorts)
	}

	var probeSrc ports.ProbesSource
	if cfg.EnablePortscan && len(cfg.SnifferIfaces) > 0 {
		probeSrc = sniffer.NewPcapSniffer(cfg.SnifferIfaces, cfg.EnableSnifferPromisc, cfg.SnifferBPF)
	}

	return &App{
		cfg:    cfg,
		host:   host,
		jrnl:   jrnl,
		parser: parser,
		sender: sender,
		probes: probeSrc,
		psd:    det,
	}
}

func (a *App) Run(ctx context.Context) error {
	defer a.sender.Close()

	var probeCh <-chan *domain.Event
	if a.probes != nil {
		ch, err := a.probes.Start(ctx)
		if err != nil {
			log.Printf("sniffer start: %v", err)
		} else {
			probeCh = ch
			log.Printf("[Monic] sniffer started")
		}

		defer a.probes.Close()
	}

	if err := a.jrnl.Init(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Monic] stopping")
			return nil

		case pev := <-probeCh:
			if pev == nil {
				continue
			}

			if !a.cfg.EnablePortscan && pev.Type == "net_probe" {
				continue
			}

			pev.Server = a.host

			if a.cfg.DebugMode {
				if out, err := json.Marshal(pev); err == nil {
					fmt.Println(string(out))
				}
			}

			if err := a.sender.Send(ctx, pev); err != nil {
				log.Printf("send error: %v", err)
			}

			if a.psd != nil && pev.Type == "net_probe" {
				if agg := a.psd.Feed(time.Now().UTC(), pev); agg != nil {
					agg.Server = a.host
					if a.cfg.DebugMode {
						if out, err := json.Marshal(agg); err == nil {
							fmt.Println(string(out))
						}
					}
					if err := a.sender.Send(ctx, agg); err != nil {
						log.Printf("send error (portscan): %v", err)
					}
				}
			}

		default:
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
				if !a.cfg.EnablePortscan && ev.Type == "net_probe" {
					continue
				}

				ev.Server = a.host
				ev.DateTime = entry.DateTime.UTC()
				ev.Raw = entry.Message

				if a.cfg.DebugMode {
					if out, err := json.Marshal(ev); err == nil {
						fmt.Println(string(out))
					}
				}

				if err := a.sender.Send(ctx, ev); err != nil {
					log.Printf("send error: %v", err)
				}

				if a.psd != nil && ev.Type == "net_probe" {
					if agg := a.psd.Feed(entry.DateTime, ev); agg != nil {
						agg.Server = a.host
						if a.cfg.DebugMode {
							if out, err := json.Marshal(agg); err == nil {
								fmt.Println(string(out))
							}
						}
						if err := a.sender.Send(ctx, agg); err != nil {
							log.Printf("send error (portscan): %v", err)
						}
					}
				}
			}

			if cur := entry.Cursor; cur != "" {
				_ = a.jrnl.SaveCursor(cur)
			}
		}
	}
}
