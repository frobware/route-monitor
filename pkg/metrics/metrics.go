package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

var (
	routesReachableGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "routes_reachable",
			Help: "Number of reachable routes.",
		},
		[]string{"name"},
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
	klog.V(0).Infof("route %q is reachable", name)
	routesReachableGauge.With(prometheus.Labels{"name": name}).Set(1)
}

func SetRoutesUnreachableGauge(name string) {
	klog.V(0).Infof("route %q is unreachable", name)
	routesReachableGauge.With(prometheus.Labels{"name": name}).Set(0)
}

func IncRoutesReachableCounter(name string) {
	klog.V(0).Infof("route %q is reachable", name)
	routesReachableCounter.With(prometheus.Labels{"name": name}).Inc()
}

func IncRoutesUnreachableCounter(name string) {
	klog.V(0).Infof("route %q is NOT reachable", name)
	routesUnReachableCounter.With(prometheus.Labels{"name": name}).Inc()
}

func IncRoutesUnknownCounter(name string) {
	klog.V(0).Infof("route %q is unknown", name)
	routesUnknownCounter.With(prometheus.Labels{"name": name}).Inc()
}

func SetRouteStatus(name string, status string) {
	routesReachableHistogram.WithLabelValues(name, status).Observe(1.0)
}

func SetRouteReachable(name string) {
	// Experimenting with metric types; what makes most sense here?

	// Historgram?
	// SetRouteStatus(name, "200")

	// Counter?
	// IncRoutesReachableCounter(name)

	// Gauge?
	SetRoutesReachableGauge(name)
}

func SetRouteUnreachable(name string) {
	// Experimenting with metric types; what makes most sense here?

	// Histogram?
	// SetRouteStatus(name, "404")

	// Counter?
	// IncRoutesUnreachableCounter(name)

	// Gauge?
	SetRoutesUnreachableGauge(name)
}

func SetRouteUnknown(name string) {
}

func init() {
	prometheus.MustRegister(routesReachableCounter)
	prometheus.MustRegister(routesUnReachableCounter)
	prometheus.MustRegister(routesUnknownCounter)
	prometheus.MustRegister(routesReachableGauge)
	prometheus.MustRegister(routesReachableHistogram)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}
