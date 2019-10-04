package route_monitor

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"k8s.io/klog"
)

type Logger interface {
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

type Connector interface {
	Connect(route Route) (bool, error)
}

var defaultConnectionTimeout = time.Duration(5 * time.Second)

type Route struct {
	Host    string
	Timeout *time.Duration
}

type RouteMonitor struct {
	config RouteMonitorConfig
	routes []Route
}

type RouteMonitorConfig struct {
	Connector Connector
	Logger    Logger
}

type defaultLogger struct{}
type defaultConnector struct{}

// defaultLogger is a Logger
var _ Logger = (*defaultLogger)(nil)

// defaultConnector is a Connector
var _ Connector = (*defaultConnector)(nil)

func NewRouteMonitor(routes []Route, config RouteMonitorConfig) *RouteMonitor {
	return &RouteMonitor{
		config: config,
		routes: routes,
	}
}

func (m RouteMonitor) VerifyRoutes(ctx context.Context, stopCh <-chan struct{}) {
	for i := range m.routes {
		m.config.Logger.Infof("[%d/%d] verifying route connectivity for %q", i, len(m.routes), m.routes[i].Host)
		fmt.Println(m.config.Connector.Connect(m.routes[i]))
	}
}

func (r Route) GetTimeout() time.Duration {
	if r.Timeout != nil {
		return *r.Timeout
	}
	return defaultConnectionTimeout
}

func (d defaultConnector) Connect(route Route) (bool, error) {
	// construct the client and request. The HTTP client timeout
	// is independent of the context timeout.
	client := http.Client{Timeout: route.GetTimeout()}
	req, err := http.NewRequest(http.MethodGet, route.Host, nil)

	// We initialize the context, and specify a context timeout.
	ctx, cancel := context.WithTimeout(context.Background(), route.GetTimeout())
	defer cancel() // cancel() is a hook to cancel the deadline

	// We attach the initialized context to the request, and
	// execute a request with it.
	reqWithDeadline := req.WithContext(ctx)
	response, clientErr := client.Do(reqWithDeadline)
	if clientErr != nil {
		return false, err
	}

	defer response.Body.Close()

	_, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return false, readErr
	}

	return true, nil
}

func (d defaultLogger) Infof(format string, a ...interface{}) {
	klog.Infof(format, a...)
}

func (d defaultLogger) Debugf(format string, a ...interface{}) {
	klog.V(4).Infof(format, a...)
}

func (d defaultLogger) Errorf(format string, a ...interface{}) {
	klog.Errorf(format, a...)
}

func DefaultConnector() defaultConnector {
	return defaultConnector{}
}

func DefaultLogger() defaultLogger {
	return defaultLogger{}
}
