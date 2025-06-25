package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func BenchmarkVerifyHMAC(b *testing.B) {
	secret := []byte("secret")
	body := []byte("benchmark")
	sig := "" // compute once
	// compute correct signature once
	h := hmac.New(sha256.New, secret)
	h.Write(body)
	sig = hex.EncodeToString(h.Sum(nil))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := VerifyHMAC(secret, body, sig); err != nil {
			b.Fatal(err)
		}
	}
}
