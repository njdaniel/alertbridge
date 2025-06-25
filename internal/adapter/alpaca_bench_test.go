package adapter

import "testing"

func BenchmarkIsCrypto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = isCrypto("ETHUSD")
	}
}
