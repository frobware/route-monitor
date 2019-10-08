package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/frobware/route-monitor/pkg/controller"
	"github.com/frobware/route-monitor/pkg/metrics"
	"github.com/frobware/route-monitor/pkg/probe"
	probehttp "github.com/frobware/route-monitor/pkg/probe/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	metricsAddress = flag.String("listen-address", ":8000", "The address to listen on for metric requests.")
	kubeconfig     = flag.String("kubeconfig", "", "Path to kubeconfig file with authorization and master location information.")
)

func probeRoutes(prober probehttp.Prober, timeout time.Duration, controller *controller.RouteController, names []string) {
	for _, name := range names {
		route, err := controller.GetRoute(name)
		if err != nil || route == nil {
			log.Printf("route %q is unknown\n", name)
			metrics.SetRouteUnknown(name)
			continue
		}

		hostURL, err := url.Parse(fmt.Sprintf("https://%s", route.Host()))
		if err != nil {
			log.Printf("error parsing %q as a URL: %v", route.Host(), err)
			continue
		}

		log.Printf("probing %q\n", hostURL.String())
		result, _, _ := prober.Probe(hostURL, nil, timeout)
		switch result {
		case probe.Success, probe.Warning:
			log.Printf("route %q is reachable\n", name)
			metrics.SetRouteReachable(name)
		case probe.Failure, probe.Unknown:
			log.Printf("route %q is unreachable\n", name)
			metrics.SetRouteUnreachable(name)
		default:
			panic("unhandled probe result")
		}
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

	go func() {
		for {
			prober := probehttp.New(true)
			probeRoutes(prober, 5*time.Second, routeController, flag.Args())
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*metricsAddress, nil))
	}()

	<-stopCh
}
