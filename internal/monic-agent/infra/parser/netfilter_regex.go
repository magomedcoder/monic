package parser

import (
	"regexp"
	"strings"

	"github.com/magomedcoder/monic/internal/monic-agent/domain"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
)

type netfilterRegexParser struct{}

func NewNetfilterRegexParser() ports.EventParser {
	return &netfilterRegexParser{}
}

var (
	reSrc   = regexp.MustCompile(`\bSRC=([0-9a-fA-F:\.]+)\b`)
	reDpt   = regexp.MustCompile(`\bDPT=(\d+)\b`)
	reProto = regexp.MustCompile(`\bPROTO=(\w+)\b`)
	reSyn   = regexp.MustCompile(`\bSYN\b`)
)

func (p *netfilterRegexParser) Parse(msg string) *domain.Event {
	m := strings.TrimSpace(msg)
	if !strings.Contains(m, "SRC=") || !strings.Contains(m, "DPT=") {
		return nil
	}

	src := find1(reSrc, m)
	dpt := find1(reDpt, m)
	proto := strings.ToLower(find1(reProto, m))

	if src == "" || dpt == "" {
		return nil
	}

	ev := &domain.Event{
		Type:     "net_probe",
		User:     "",
		RemoteIP: src,
		Port:     dpt,
		Method:   proto,
		Message:  "probe",
	}
	if reSyn.MatchString(m) {
		ev.Message = "syn_probe"
	}

	return ev
}

func find1(re *regexp.Regexp, s string) string {
	if mm := re.FindStringSubmatch(s); mm != nil {
		return mm[1]
	}

	return ""
}
