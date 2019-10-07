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

type connected func(name string, reachable bool)

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

func monitorRoutes(controller *controller.RouteController, names []string, f connected) {
	for {
		// TODO(frobware) - how many of these monitoring routes do we expect?
		// TODO(frobware) - should we do them all in bulk? if so, how many?
		// TODO(frobware) - do we need timeout/cancellation

		for _, name := range names {
			route, err := controller.GetRoute(name)
			if err != nil {
				klog.Errorf("error fetching route %q: %v", name, err)
				f(name, false)
				continue
			}
			if route == nil {
				klog.Errorf("unknown route %q", name)
				f(name, false)
				continue
			}

			url, err := route.URL()
			if err != nil {
				klog.Errorf("failed to parse URL for route %q: %v", route.Host(), err)
				f(name, false)
				continue
			}

			if _, err := connect(url, defaultConnectionTimeout); err != nil {
				klog.Errorf("failed to connect to %q: %v", url.String(), err)
				f(name, false)
				continue
			}

			f(name, true)
			klog.Infof("route %q IS reachable", fmt.Sprintf("%s/%s", route.Namespace(), route.Name()))
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

	routeController, err := controller.NewController(client)
	if err != nil {
		klog.Fatalf("failed to create routeController: %v", err)
	}

	stopCh := signals.SetupSignalHandler()

	klog.Info("starting route controller")
	if err := routeController.Start(stopCh); err != nil {
		klog.Fatalf("failed to start routeController: %v", err)
	}

	go monitorRoutes(routeController, flag.Args(), func(name string, reachable bool) {
		switch reachable {
		case true:
			metrics.ReachableRoutes.With(prometheus.Labels{"route": name}).Inc()
		default:
			metrics.UnreachableRoutes.With(prometheus.Labels{"route": name}).Inc()
		}
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*metricsAddress, nil))
	}()

	<-stopCh
}
