package adapter

import "testing"

var result bool // Package-level variable to prevent compiler optimization

func BenchmarkIsCrypto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result = isCrypto("ETHUSD")
	}
}
