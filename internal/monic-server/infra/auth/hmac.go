package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type HMACVerifier struct {
	secret string
}

func NewHMACVerifier(secret string) *HMACVerifier {
	return &HMACVerifier{secret: secret}
}

func (v *HMACVerifier) Verify(header string, body []byte) error {
	if v.secret == "" {
		return nil
	}

	if header == "" {
		return errors.New("no signature")
	}

	const prefix = "sha256="
	if !strings.HasPrefix(header, prefix) {
		return errors.New("bad sig format")
	}

	exp, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return err
	}

	h := hmac.New(sha256.New, []byte(v.secret))
	_, _ = h.Write(body)

	if !hmac.Equal(h.Sum(nil), exp) {
		return errors.New("signature mismatch")
	}

	return nil
}
