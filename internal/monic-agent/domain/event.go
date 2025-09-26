package domain

import "time"

type Event struct {
	DateTime time.Time `json:"dateTime"`
	Server   string    `json:"server"`
	Type     string    `json:"type"`
	User     string    `json:"user"`
	RemoteIP string    `json:"remoteIp"`
	Port     string    `json:"port,omitempty"`
	Method   string    `json:"method,omitempty"`
	Message  string    `json:"message"`
	Raw      string    `json:"raw"`
}
