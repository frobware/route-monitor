package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	UnreachableRoutes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "unreachable_routes",
			Help: "Number of unreachable routes.",
		},
		[]string{"route"},
	)
)

func init() {
	prometheus.MustRegister(UnreachableRoutes)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}
