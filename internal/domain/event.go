package domain

import "time"

type Event struct {
	TS       time.Time `json:"ts"`
	Server   string    `json:"server"`
	Type     string    `json:"type"`
	User     string    `json:"user"`
	RemoteIP string    `json:"rhost"`
	Port     string    `json:"port,omitempty"`
	Method   string    `json:"method,omitempty"`
	Message  string    `json:"message"`
	Raw      string    `json:"raw"`
}
