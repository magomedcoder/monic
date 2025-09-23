package ports

import "context"

type WebhookSender interface {
	Send(ctx context.Context, payload []byte) error
}
