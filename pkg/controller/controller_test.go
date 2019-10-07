package controller

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

type testControllerShutdownFunc func()

func newTestController(t *testing.T, routes ...string) (*RouteController, testControllerShutdownFunc) {
	t.Helper()
	routeObjects := make([]runtime.Object, 0)

	for _, route := range routes {
		obj, err := newUnstructuredFromRoute(route, "test", route)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		routeObjects = append(routeObjects, obj)
	}

	clientSet := fake.NewSimpleDynamicClient(runtime.NewScheme(), routeObjects...)
	controller, err := NewController(clientSet)
	if err != nil {
		t.Fatal("failed to create test controller")
	}

	stopCh := make(chan struct{})
	if err := controller.Start(stopCh); err != nil {
		t.Fatalf("failed to start controller: %v", err)
	}

	return controller, func() {
		close(stopCh)
	}
}

func newUnstructuredFromRoute(name, namespace, host string) (*unstructured.Unstructured, error) {
	u := unstructured.Unstructured{}
	u.SetAPIVersion("route.openshift.io/v1")
	u.SetKind("Route")
	u.SetName(name)
	u.SetNamespace(namespace)
	if err := unstructured.SetNestedField(u.Object, host, "spec", "host"); err != nil {
		return nil, err
	}
	return &u, nil
}

func TestNewController(t *testing.T) {
	c, cleanup := newTestController(t)
	defer cleanup()

	routes := c.GetRoutes(v1.NamespaceAll)
	if actual, expected := len(routes), 0; expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}

	newRoute := "foo.bar.com"
	obj, err := newUnstructuredFromRoute(newRoute, "test", newRoute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.routeInformer.Informer().GetStore().Add(obj); err != nil {
		t.Fatalf("failed to add new route: %v", err)
	}

	routes = c.GetRoutes(v1.NamespaceAll)
	if actual, expected := len(routes), 1; expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}

	if routes[0] != newRoute {
		t.Errorf("expected %q, got %q", newRoute, routes[0])
	}

	routesIndexedByNamespace := c.GetRoutesIndexedByNamespace()
	val, ok := routesIndexedByNamespace["test"]
	if !ok {
		t.Fatalf("expected to find a route in namespace %q", "test")
	}
	if val != newRoute {
		t.Errorf("expected %q, got %q", newRoute, val)
	}
}
