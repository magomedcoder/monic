package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/magomedcoder/monic/internal/monic-agent/domain"
	"github.com/magomedcoder/monic/internal/monic-agent/ports"
	"io"
	"net/http"
	"time"
)

type httpSender struct {
	url    string
	secret string
	client *http.Client
}

func NewHTTPSender(url, secret string) ports.EventSender {
	return &httpSender{
		url:    url,
		secret: secret,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *httpSender) Send(ctx context.Context, ev *domain.Event) error {
	if h.url == "" {
		return nil
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url+"/webhook", bytesReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if h.secret != "" {
		m := hmac.New(sha256.New, []byte(h.secret))
		_, _ = m.Write(payload)
		sig := hex.EncodeToString(m.Sum(nil))
		req.Header.Set("X-Signature", "sha256="+sig)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return errors.New(resp.Status + ": " + string(b))
	}

	return nil
}

func (h *httpSender) Close() error {
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
