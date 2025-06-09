package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OrderTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_total",
			Help: "Total number of orders processed",
		},
		[]string{"bot", "side"},
	)
)

func init() {
	prometheus.MustRegister(OrderTotal)
}
