package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/frobware/route-monitor/pkg/controller"
	"github.com/frobware/route-monitor/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type reachableStatus int
type reachableFunc func(route string, status reachableStatus)

const (
	UnknownRoute reachableStatus = iota
	ReachableRoute
	UnreachableRoute
)

var (
	defaultConnectionTimeout = 5 * time.Second
	metricsAddress           = flag.String("listen-address", ":8000", "The address to listen on for metric requests.")
	kubeconfig               = flag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
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

func connectRoutes(controller *controller.RouteController, names []string, f reachableFunc) {
	var i uint64 = 0

	for {
		// TODO(frobware) - how many of these monitoring routes do we expect?
		// TODO(frobware) - should we do them all in bulk? if so, how many?
		// TODO(frobware) - do we need timeout/cancellation

		i += 1

		for _, name := range names {
			// // Randomly fail/testing/experimentation
			// if i%2 == 0 && os.Getenv("RANDOM_RESULT") != "" {
			// 	switch n := rand.Intn(3); n {
			// 	case 0:
			// 		klog.Infof("**** Ramdomising result for %q to ReachableRoute", name)
			// 		f(name, ReachableRoute)
			// 	case 1:
			// 		klog.Infof("**** Ramdomising result for %q to UnreachableRoute", name)
			// 		f(name, UnreachableRoute)
			// 	case 2:
			// 		klog.Infof("**** Ramdomising result for %q to UnknownRoute", name)
			// 		f(name, UnknownRoute)
			// 	}
			// 	continue
			// }

			route, err := controller.GetRoute(name)
			if err != nil || route == nil {
				f(name, UnknownRoute)
				continue
			}

			url, err := route.URL()
			if err != nil {
				klog.Errorf("failed to parse URL for route %q: %v", route.Host(), err)
				f(name, UnknownRoute)
				continue
			}

			if _, err := connect(url, defaultConnectionTimeout); err != nil {
				f(name, UnreachableRoute)
				continue
			}

			f(name, ReachableRoute)
		}

		// TODO(frobware) - parameterise
		time.Sleep(5 * time.Second)
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

	go connectRoutes(routeController, flag.Args(), func(name string, status reachableStatus) {
		switch status {
		case ReachableRoute:
			klog.V(0).Infof("route %q is reachable", name)
			metrics.SetRouteReachable(name)
		case UnreachableRoute:
			klog.V(0).Infof("route %q is unreachable", name)
			metrics.SetRouteUnreachable(name)
		case UnknownRoute:
			klog.V(0).Infof("route %q is unknown", name)
			metrics.SetRouteUnknown(name)
		}
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*metricsAddress, nil))
	}()

	<-stopCh
}
