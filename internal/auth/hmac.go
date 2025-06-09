package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type HMACVerifier struct {
	secret string
}

func NewHMACVerifier(secret string) *HMACVerifier {
	return &HMACVerifier{
		secret: secret,
	}
}

func (v *HMACVerifier) IsEnabled() bool {
	return v.secret != ""
}

func (v *HMACVerifier) Verify(bodyBytes []byte, signature string) bool {
	if !v.IsEnabled() {
		return true
	}

	// Calculate HMAC
	h := hmac.New(sha256.New, []byte(v.secret))
	h.Write(bodyBytes)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
