package controller

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// TODO(frobware) parameterize this
const routesV1 = "routes.v1.route.openshift.io"

type RouteController struct {
	client          dynamic.Interface
	informerFactory dynamicinformer.DynamicSharedInformerFactory
	routeInformer   informers.GenericInformer
	routeResource   *schema.GroupVersionResource
}

// NewController constructs a new RouteController that watches for routes
// as they are added, updated and deleted on the cluster.
func NewController(client dynamic.Interface) (*RouteController, error) {
	informerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil)
	routeResource, _ := schema.ParseResourceArg(routesV1)

	if routeResource == nil {
		return nil, fmt.Errorf("failed to parse resource: %q", routesV1)
	}

	routeInformer := informerFactory.ForResource(*routeResource)
	routeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})

	return &RouteController{
		client:          client,
		informerFactory: informerFactory,
		routeInformer:   routeInformer,
		routeResource:   routeResource,
	}, nil
}

// Start starts the route informer
func (c *RouteController) Start(stopCh <-chan struct{}) error {
	c.informerFactory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.routeInformer.Informer().HasSynced) {
		return fmt.Errorf("syncing caches failed")
	}
	return nil
}

// GetRoutes returns all the routes in namespace
func (c *RouteController) GetRoutes(namespace string) []string {
	var routes []string

	for _, x := range c.routeInformer.Informer().GetStore().List() {
		u := x.(*unstructured.Unstructured).DeepCopy()
		if host := newRouteHostFromUnstructured(u); host != "" {
			routes = append(routes, host)
		}
	}

	return routes
}

// GetRoutesIndexedByNamespace returns all the routes mapped by their
// namespace.
func (c *RouteController) GetRoutesIndexedByNamespace() map[string]string {
	routes := map[string]string{}

	for _, x := range c.routeInformer.Informer().GetStore().List() {
		u := x.(*unstructured.Unstructured).DeepCopy()
		if host := newRouteHostFromUnstructured(u); host != "" {
			routes[u.GetNamespace()] = host
		}
	}

	return routes
}

func newRouteHostFromUnstructured(u *unstructured.Unstructured) string {
	host, _, _ := unstructured.NestedString(u.Object, "spec", "host")
	return host
}
