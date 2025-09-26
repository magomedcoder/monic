package parser

import (
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
	"regexp"
	"strings"
)

type sshdRegexParser struct{}

func NewSSHDRegexParser() ports.EventParser {
	return &sshdRegexParser{}
}

var (
	reAccepted     = regexp.MustCompile(`^Accepted\s+(\S+)\s+for\s+(\S+)\s+from\s+(\S+)\s+port\s+(\d+)`)
	reFailed       = regexp.MustCompile(`^Failed\s+(\S+)\s+for\s+(\S+)\s+from\s+(\S+)\s+port\s+(\d+)`)
	reInvalidUser  = regexp.MustCompile(`^Invalid\s+user\s+(\S+)\s+from\s+(\S+)\s+port\s+(\d+)`)
	reDisconnected = regexp.MustCompile(`^Disconnected\s+from\s+(\S+)\s+port\s+(\d+)`)
	reConnClosed   = regexp.MustCompile(`^Connection\s+closed\s+by\s+(\S+)\s+port\s+(\d+)`)
)

func (s *sshdRegexParser) Parse(msg string) *domain.Event {
	msg = strings.TrimSpace(msg)
	if m := reAccepted.FindStringSubmatch(msg); m != nil {
		return &domain.Event{
			Type:     "ssh_accepted",
			User:     m[2],
			RemoteIP: m[3],
			Port:     m[4],
			Method:   m[1],
			Message:  "accepted",
		}
	}

	if m := reFailed.FindStringSubmatch(msg); m != nil {
		return &domain.Event{
			Type:     "ssh_failed",
			User:     m[2],
			RemoteIP: m[3],
			Port:     m[4],
			Method:   m[1],
			Message:  "failed",
		}
	}

	if m := reInvalidUser.FindStringSubmatch(msg); m != nil {
		return &domain.Event{
			Type:     "ssh_invalid_user",
			User:     m[1],
			RemoteIP: m[2],
			Port:     m[3],
			Message:  "invalid_user",
		}
	}

	if m := reDisconnected.FindStringSubmatch(msg); m != nil {
		return &domain.Event{
			Type:     "ssh_disconnect",
			RemoteIP: m[1],
			Port:     m[2],
			Message:  "disconnected",
		}
	}

	if m := reConnClosed.FindStringSubmatch(msg); m != nil {
		return &domain.Event{
			Type:     "ssh_disconnect",
			RemoteIP: m[1],
			Port:     m[2],
			Message:  "connection_closed",
		}
	}

	return nil
}
