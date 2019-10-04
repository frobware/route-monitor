package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	UnreachableHosts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "unreachable_hosts",
			Help: "Number of unreachable hosts.",
		},
		[]string{"host", "namespace"},
	)
)

func init() {
	prometheus.MustRegister(UnreachableHosts)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}
