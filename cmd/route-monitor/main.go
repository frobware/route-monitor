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
			route, err := controller.GetRoute(name)
			if err != nil || route == nil {
				f(name, UnknownRoute)
				continue
			}

			hostURL, err := route.URL()
			if err != nil {
				log.Printf("failed to parse URL for route %q: %v", route.Host(), err)
				f(name, UnknownRoute)
				continue
			}

			if _, err := connect(hostURL, defaultConnectionTimeout); err != nil {
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
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("failed to build config: %v\n", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create config: %v\n", err)
	}

	routeController, err := controller.NewController(client)
	if err != nil {
		log.Fatalf("failed to create routeController: %v\n", err)
	}

	stopCh := signals.SetupSignalHandler()

	log.Println("starting route controller")
	if err := routeController.Start(stopCh); err != nil {
		log.Fatalf("failed to start routeController: %v\n", err)
	}

	go connectRoutes(routeController, flag.Args(), func(name string, status reachableStatus) {
		switch status {
		case ReachableRoute:
			log.Printf("route %q is reachable\n", name)
			metrics.SetRouteReachable(name)
		case UnreachableRoute:
			log.Printf("route %q is unreachable\n", name)
			metrics.SetRouteUnreachable(name)
		case UnknownRoute:
			log.Printf("route %q is unknown\n", name)
			metrics.SetRouteUnknown(name)
		}
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*metricsAddress, nil))
	}()

	<-stopCh
}
