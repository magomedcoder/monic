package domain

import "time"

type Event struct {
	TS       time.Time `json:"ts"`
	Server   string    `json:"server"`
	Type     string    `json:"type"`
	User     string    `json:"user"`
	RemoteIP string    `json:"remoteIp"`
	Port     string    `json:"port,omitempty"`
	Method   string    `json:"method,omitempty"`
	Message  string    `json:"message"`
	Raw      string    `json:"raw"`
}

type IngestedEvent struct {
	Event
	ReceivedAt time.Time
}
