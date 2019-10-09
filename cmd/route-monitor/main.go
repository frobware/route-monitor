package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/frobware/route-monitor/pkg/metrics"
	"github.com/frobware/route-monitor/pkg/probe"
	probehttp "github.com/frobware/route-monitor/pkg/probe/http"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logf.Log.WithName("route-monitor")

var (
	metricsAddress = flag.String("metrics-address", ":8081", "The address to listen on for metric requests.")
)

func probeRoutes(prober probehttp.Prober, timeout time.Duration, names []string, mgr manager.Manager) {
	log := log.WithName("prober")

	for _, name := range names {
		route := routev1.Route{}

		tokens := strings.Split(name, "/")
		if len(tokens) != 2 {
			log.Error(nil, "invalid name", "name", name)
			metrics.SetRouteUnknown(name)
			continue
		}

		key := types.NamespacedName{
			Namespace: tokens[0],
			Name:      tokens[1],
		}

		if err := mgr.GetCache().Get(context.Background(), key, &route); err != nil {
			log.Error(err, "unknown route", "name", name)
			metrics.SetRouteUnknown(name)
			continue
		}

		hostURL, err := url.Parse(fmt.Sprintf("https://%s", route.Spec.Host))
		if err != nil {
			log.Error(err, "invalid URL", "host", route.Spec.Host)
			continue
		}

		result, _, _ := prober.Probe(hostURL, nil, timeout)
		switch result {
		case probe.Success, probe.Warning:
			log.Info("reachable", "name", name, "url", hostURL)
			metrics.SetRouteReachable(name)
		case probe.Failure, probe.Unknown:
			log.Info("unreachable", "name", name, "url", hostURL)
			metrics.SetRouteUnreachable(name)
		default:
			panic("unhandled probe result")
		}
	}
}

// reconcileRoute reconciles Routes
type reconcileRoute struct {
	// client can be used to retrieve objects from the APIServer
	client client.Client
	log    logr.Logger
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &reconcileRoute{}

func (r *reconcileRoute) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log.Info("reconcile", "name", request.NamespacedName)
	return reconcile.Result{}, nil
}

func main() {
	logf.SetLogger(zap.Logger(true))

	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Error(nil, "nothing to monitor!")
		os.Exit(1)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(*metricsAddress, nil); err != nil {
			log.Error(err, "metrics listener failed", "address", *metricsAddress)
			os.Exit(1)
		}
	}()

	log.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	if err := routev1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "adding routev1 scheme failed")
		os.Exit(1)
	}

	log.Info("Setting up controller")
	_, err = controller.New("controller", mgr, controller.Options{
		Reconciler: &reconcileRoute{client: mgr.GetClient(), log: log.WithName("reconciler")},
	})

	if err != nil {
		log.Error(err, "unable to set up controller")
		os.Exit(1)
	}

	go func() {
		prober := probehttp.New(true)
		for {
			probeRoutes(prober, 5*time.Second, flag.Args(), mgr)
			time.Sleep(5 * time.Second)
		}
	}()

	log.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
