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

	ReachableRoutes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reachable_routes",
			Help: "Number of reachable routes.",
		},
		[]string{"route"},
	)
)

func init() {
	prometheus.MustRegister(ReachableRoutes)
	prometheus.MustRegister(UnreachableRoutes)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}
