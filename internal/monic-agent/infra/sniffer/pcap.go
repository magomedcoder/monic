package sniffer

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
)

type PcapSniffer struct {
	ifaces  []string
	promisc bool
	bpf     string
	handles []*pcap.Handle
	outCh   chan *domain.Event
}

func NewPcapSniffer(ifaces []string, promisc bool, bpf string) ports.ProbesSource {
	return &PcapSniffer{
		ifaces:  ifaces,
		promisc: promisc,
		bpf:     bpf,
		outCh:   make(chan *domain.Event, 1024),
	}
}

func (s *PcapSniffer) Start(ctx context.Context) (<-chan *domain.Event, error) {
	if len(s.ifaces) == 0 {
		return s.outCh, nil
	}

	filter := s.bpf
	if filter == "" {
		filter = "(tcp[tcpflags] & (tcp-syn) != 0 and (tcp[tcpflags] & tcp-ack) = 0) or udp"
	}

	for _, ifc := range s.ifaces {
		h, err := pcap.OpenLive(ifc, 65535, s.promisc, pcap.BlockForever)
		if err != nil {
			return nil, fmt.Errorf("pcap open %s: %w", ifc, err)
		}

		if err := h.SetBPFFilter(filter); err != nil {
			h.Close()
			return nil, fmt.Errorf("bpf on %s: %w", ifc, err)
		}

		s.handles = append(s.handles, h)
		go s.loop(ctx, h)
	}

	return s.outCh, nil
}

func (s *PcapSniffer) loop(ctx context.Context, h *pcap.Handle) {
	src := gopacket.NewPacketSource(h, h.LinkType())
	src.NoCopy = true
	for {
		select {
		case <-ctx.Done():
			return
		case pkt, ok := <-src.Packets():
			if !ok {
				return
			}
			if ev := parse(pkt); ev != nil {
				select {
				case s.outCh <- ev:
				default:
				}
			}
		}
	}
}

func parse(pkt gopacket.Packet) *domain.Event {
	nl := pkt.NetworkLayer()
	if nl == nil {
		return nil
	}

	var srcIP net.IP
	switch ip := nl.(type) {
	case *layers.IPv4:
		srcIP = ip.SrcIP
	case *layers.IPv6:
		srcIP = ip.SrcIP
	default:
		return nil
	}

	now := time.Now().UTC()

	switch tl := pkt.TransportLayer().(type) {
	case *layers.TCP:
		if tl != nil && tl.SYN && !tl.ACK {
			return &domain.Event{
				DateTime: now,
				Type:     "net_probe",
				RemoteIP: srcIP.String(),
				Port:     fmt.Sprintf("%d", tl.DstPort),
				Method:   "tcp",
				Message:  "syn_probe",
			}
		}
	case *layers.UDP:
		if tl != nil {
			return &domain.Event{
				DateTime: now,
				Type:     "net_probe",
				RemoteIP: srcIP.String(),
				Port:     fmt.Sprintf("%d", tl.DstPort),
				Method:   "udp",
				Message:  "probe",
			}
		}
	}
	return nil
}

func (s *PcapSniffer) Close() error {
	for _, h := range s.handles {
		h.Close()
	}

	close(s.outCh)

	return nil
}
