package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/frobware/route-monitor/pkg/controller"
	"github.com/frobware/route-monitor/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	defaultConnectionTimeout = 5 * time.Second
	metricsAddress           = flag.String("listen-address", ":8000", "The address to listen on for metric requests.")
	kubeconfig               = flag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
)

type route struct {
	scheme    string
	host      string
	namespace string
}

// TODO(frobware) take these from the command line or the environment
var routes = []route{{
	scheme: "https",
	host:   "foo.com",
}, {
	scheme: "https",
	host:   "bar.com",
}, {
	scheme:    "https",
	host:      "console-openshift-console.apps.amcdermo.devcluster.openshift.com",
	namespace: "openshift-console",
}, {
	scheme:    "https",
	host:      "console-openshift-console.apps.amcdermo.devcluster.openshift.com",
	namespace: "openshift-console",
},
}

func connect(url *url.URL, timeout time.Duration) (bool, error) {
	// construct the client and request. The HTTP client timeout
	// is independent of the context timeout.
	client := http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return false, err
	}

	// We initialize the context, and specify a context timeout.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

type connected func(r route, reachable bool)

func monitorRoutes(controller *controller.RouteController, routes []route, f connected) {
	for {
		existingRoutes := controller.GetRoutesIndexedByNamespace()
		for k, v := range existingRoutes {
			klog.Infof("existing route: namespace: %q, host: %q", k, v)
		}

		// TODO(frobware) - how many of these monitoring routes do we expect?
		// TODO(frobware) - should we do them all in bulk? if so, how many?
		// TODO(frobware) - do we need timeout/cancellation

		for _, r := range routes {
			if _, ok := existingRoutes[r.namespace]; !ok {
				klog.Errorf("route %q does not exist", r.host)
				f(r, false)
				continue
			}

			// TODO(frobware) - infer scheme from cluster route object
			rawurl := fmt.Sprintf("%s://%s", r.scheme, r.host)
			route, err := url.Parse(rawurl)
			if err != nil {
				klog.Errorf("failed to parse URL %q: %v", rawurl, err)
				f(r, false)
				continue
			}

			if _, err := connect(route, defaultConnectionTimeout); err != nil {
				klog.Errorf("failed to connect to %q: %v", route.String(), err)
				f(r, false)
				continue
			}

			f(r, true)
			klog.Infof("route %q IS reachable", route.String())
		}

		// TODO(frobware) - parameterise
		time.Sleep(1 * time.Second)
	}
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatalf("failed to build config: %v", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.Fatalf("failed to create config: %v", err)
	}

	stopCh := signals.SetupSignalHandler()

	routeController, err := controller.NewController(client)
	if err != nil {
		klog.Fatalf("failed to create routeController: %v", err)
	}

	klog.Info("starting route routeController")
	if err := routeController.Start(stopCh); err != nil {
		klog.Fatalf("failed to start routeController: %v", err)
	}

	go monitorRoutes(routeController, routes, func(r route, reachable bool) {
		if !reachable {
			metrics.UnreachableHosts.With(prometheus.Labels{"host": r.host, "namespace": r.namespace}).Inc()
		}
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*metricsAddress, nil))
	}()

	<-stopCh
}
