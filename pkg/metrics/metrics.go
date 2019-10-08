package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	routesReachableGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "route_state",
			Help: "Current state of a route.",
		},
		[]string{"name", "reachable", "unreachable", "unknown"},
	)

	routesReachableHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "route_reachability",
			Help: "Number of reachable routes.",
		},
		[]string{"name", "status"},
	)

	routesReachableCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "routes_reachable_total",
			Help: "Number of reachable routes.",
		},
		[]string{"name"},
	)

	routesUnReachableCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "routes_unreachable_total",
			Help: "Number of unreachable routes.",
		},
		[]string{"name"},
	)

	routesUnknownCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "routes_unknown_total",
			Help: "Number of unknown routes.",
		},
		[]string{"name"},
	)
)

func SetRoutesReachableGauge(name string) {
	routesReachableGauge.With(prometheus.Labels{"name": name, "reachable": "1", "unreachable": "0", "unknown": "0"}).Set(1)
}

func SetRoutesUnreachableGauge(name string) {
	routesReachableGauge.With(prometheus.Labels{"name": name, "reachable": "0", "unreachable": "1", "unknown": "0"}).Set(1)
}

func IncRoutesReachableCounter(name string) {
	routesReachableCounter.With(prometheus.Labels{"name": name}).Inc()
}

func IncRoutesUnreachableCounter(name string) {
	routesUnReachableCounter.With(prometheus.Labels{"name": name}).Inc()
}

func IncRoutesUnknownCounter(name string) {
	routesUnknownCounter.With(prometheus.Labels{"name": name}).Inc()
}

func SetRouteStatus(name string, status string) {
	routesReachableHistogram.WithLabelValues(name, status).Observe(1.0)
}

func SetRouteReachable(name string) {
	// Experimenting with metric types; what makes most sense here?

	// Historgram?
	SetRouteStatus(name, "200")

	// Counter?
	IncRoutesReachableCounter(name)

	// Gauge?
	SetRoutesReachableGauge(name)
}

func SetRouteUnreachable(name string) {
	// Experimenting with metric types; what makes most sense here?

	// Histogram?
	SetRouteStatus(name, "404")

	// Counter?
	IncRoutesUnreachableCounter(name)

	// Gauge?
	SetRoutesUnreachableGauge(name)
}

func SetRouteUnknown(name string) {
	routesReachableGauge.With(prometheus.Labels{"name": name, "reachable": "0", "unreachable": "0", "unknown": "1"}).Set(1)
}

func init() {
	prometheus.MustRegister(routesReachableCounter)
	prometheus.MustRegister(routesUnReachableCounter)
	prometheus.MustRegister(routesUnknownCounter)
	prometheus.MustRegister(routesReachableGauge)
	prometheus.MustRegister(routesReachableHistogram)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}
