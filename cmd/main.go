package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/frobware/route-monitor/pkg/controller"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	defaultConnectionTimeout = time.Duration(5 * time.Second)
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
	scheme:    "https",
	host:      "downloads-openshift-console.apps.amcdermo.devcluster.openshift.com",
	namespace: "openshift-console",
},
}

func connect(url *url.URL, timeout time.Duration) (bool, error) {
	// construct the client and request. The HTTP client timeout
	// is independent of the context timeout.
	client := http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)

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

func main() {
	klog.InitFlags(nil)

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

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

	controller, err := controller.NewController(client)
	if err != nil {
		klog.Fatalf("failed to create controller: %v", err)
	}

	klog.Info("starting route controller")
	if err := controller.Start(stopCh); err != nil {
		klog.Fatalf("failed to start controller: %v", err)
	}

	go func() {
		for {
			currentRoutes := controller.GetRoutesIndexedByNamespace()
			for k, v := range currentRoutes {
				klog.Infof("existing route: namespace: %q, host: %q", k, v)
			}

			// TODO(frobware) - how many of these monitoring routes do we expect?
			// TODO(frobware) - should we do them all in bulk? if so, how many?
			// TODO(frobware) - do we need timeout/cancellation

			for _, r := range routes {
				if _, ok := currentRoutes[r.namespace]; !ok {
					klog.Errorf("route %q does not exist", r.host)
					continue
				}

				// TODO(frobware) - infer scheme from cluster route object
				url, err := url.Parse(fmt.Sprintf("%s://%s", r.scheme, r.host))
				if err != nil {
					klog.Errorf("failed to construct URL from %s/%s: %v", r.scheme, r.host, err)
					continue
				}

				if _, err := connect(url, defaultConnectionTimeout); err != nil {
					klog.Errorf("failed to connect to %q: %v", url.String(), err)
					continue
				}

				klog.Infof("route %q IS reachable", url.String())
			}

			// TODO(frobware) - parameterise
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-stopCh:
		break
	}
}
