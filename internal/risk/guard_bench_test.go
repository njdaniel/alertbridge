package risk

import "testing"

func BenchmarkGuardCheck(b *testing.B) {
	b.Setenv("PROM_URL", "")
	b.Setenv("PNL_MAX", "")
	b.Setenv("PNL_MIN", "")
	g := NewGuard("0")
	for i := 0; i < b.N; i++ {
		if err := g.Check("bot"); err != nil {
			b.Fatal(err)
		}
	}
}
