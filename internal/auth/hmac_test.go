package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHMACVerifier(t *testing.T) {
	body := []byte("hello")
	verifier := NewHMACVerifier("secret")
	h := hmac.New(sha256.New, []byte("secret"))
	h.Write(body)
	sig := hex.EncodeToString(h.Sum(nil))
	if !verifier.Verify(body, sig) {
		t.Fatalf("expected signature to be valid")
	}
	if verifier.Verify(body, "bad") {
		t.Fatalf("expected signature to be invalid")
	}
}

func TestHMACVerifierDisabled(t *testing.T) {
	verifier := NewHMACVerifier("")
	if !verifier.Verify([]byte("data"), "anything") {
		t.Fatalf("verification should always pass when disabled")
	}
}
