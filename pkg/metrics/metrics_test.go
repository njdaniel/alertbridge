package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestOrderTotalRegistered(t *testing.T) {
	OrderTotal.WithLabelValues("test", "buy").Add(0)
	count, err := testutil.GatherAndCount(prometheus.DefaultGatherer, "order_total")
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	if count == 0 {
		t.Fatal("order_total metric not registered")
	}
}

func TestOrderTotalIncrement(t *testing.T) {
	lblBot := "metrics_bot"
	lblSide := "buy"
	before := testutil.ToFloat64(OrderTotal.WithLabelValues(lblBot, lblSide))
	OrderTotal.WithLabelValues(lblBot, lblSide).Inc()
	after := testutil.ToFloat64(OrderTotal.WithLabelValues(lblBot, lblSide))
	if diff := after - before; diff != 1 {
		t.Fatalf("expected increment by 1, got %v", diff)
	}
}
