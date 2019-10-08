package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
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

func dial(host, port string, timeout time.Duration) (bool, error) {
	log.Printf("dialling %q\n", fmt.Sprintf("%s:%s", host, port))

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	dialer := net.Dialer{
		Timeout: timeout,
	}

	conn, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("%s:%s", host, port), conf)
	if err != nil {
		log.Println(err)
		return false, err
	}

	defer conn.Close()
	log.Printf("connection established: %s => %s\n", conn.LocalAddr().String(), conn.RemoteAddr().String())
	return true, nil
}

func dialRoutes(timeout time.Duration, controller *controller.RouteController, names []string, f reachableFunc) {
	for {
		// TODO(frobware) - how many of these monitoring routes do we expect?
		// TODO(frobware) - should we do them all in bulk? if so, how many?
		// TODO(frobware) - do we need timeout/cancellation
		for _, name := range names {
			route, err := controller.GetRoute(name)
			if err != nil || route == nil {
				f(name, UnknownRoute)
				continue
			}

			if _, err := dial(route.Host(), route.Port(), timeout); err != nil {
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

	go dialRoutes(defaultConnectionTimeout, routeController, flag.Args(), func(name string, status reachableStatus) {
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
