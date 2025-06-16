package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// VerifyHMAC checks the request body against the provided HMAC signature.
// When the secret is empty, the check is skipped.
func VerifyHMAC(secret, body []byte, headerSig string) error {
	if len(secret) == 0 {
		return nil
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(headerSig), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
