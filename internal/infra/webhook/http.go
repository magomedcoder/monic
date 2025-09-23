package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/magomedcoder/monic/internal/ports"
	"io"
	"net/http"
	"time"
)

type httpSender struct {
	url    string
	secret string
	client *http.Client
}

func NewHTTPSender(url, secret string) ports.WebhookSender {
	return &httpSender{
		url:    url,
		secret: secret,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}
func (h *httpSender) Send(ctx context.Context, payload []byte) error {
	if h.url == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytesReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if h.secret != "" {
		h := hmac.New(sha256.New, []byte(h.secret))
		_, _ = h.Write(payload)
		sig := hex.EncodeToString(h.Sum(nil))
		req.Header.Set("X-Signature", "sha256="+sig)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}

	return nil
}

func bytesReader(b []byte) *bytesReaderWrapper {
	return &bytesReaderWrapper{b: b}
}

type bytesReaderWrapper struct {
	b []byte
	i int
}

func (b *bytesReaderWrapper) Read(p []byte) (int, error) {
	n := copy(p, b.b[b.i:])
	b.i += n
	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}
