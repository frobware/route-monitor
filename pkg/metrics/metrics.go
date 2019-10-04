package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	routesReachableGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "route_state",
			Help: "Reachability state of a route.",
		},
		[]string{"name", "reachable", "unreachable", "unknown"},
	)
)

func SetRouteReachable(name string) {
	routesReachableGauge.With(prometheus.Labels{"name": name, "reachable": "1", "unreachable": "0", "unknown": "0"}).Set(1)
}

func SetRouteUnreachable(name string) {
	routesReachableGauge.With(prometheus.Labels{"name": name, "reachable": "0", "unreachable": "1", "unknown": "0"}).Set(1)
}

func SetRouteUnknown(name string) {
	routesReachableGauge.With(prometheus.Labels{"name": name, "reachable": "0", "unreachable": "0", "unknown": "1"}).Set(1)
}

func init() {
	prometheus.MustRegister(routesReachableGauge)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}
