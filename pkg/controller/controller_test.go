package controller

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
)

type testControllerShutdownFunc func()

func newTestController(t *testing.T, objects ...runtime.Object) (*RouteController, testControllerShutdownFunc) {
	t.Helper()

	clientSet := fake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
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

func newUnstructuredRoute(namespace, name, host string) (*unstructured.Unstructured, error) {
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

func mustCreateRoute(t *testing.T, namespace, name, host string) runtime.Object {
	obj, err := newUnstructuredRoute(namespace, name, host)
	if err != nil {
		t.Fatal(err)
	}
	return runtime.Object(obj)
}

func TestControllerNoRoutes(t *testing.T) {
	c, cleanup := newTestController(t)
	defer cleanup()

	routes, err := c.AllRoutes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if actual, expected := len(routes), 0; expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestControllerKnownRoutes(t *testing.T) {
	testCases := []struct {
		namespace string
		name      string
		host      string
	}{{
		namespace: "openshift-console",
		name:      "downloads",
		host:      "host-a",
	}, {
		namespace: "openshift-console",
		name:      "console",
		host:      "host-b",
	}}

	var expected []runtime.Object

	for _, tc := range testCases {
		expected = append(expected, mustCreateRoute(t, tc.namespace, tc.name, tc.host))
	}

	c, cleanup := newTestController(t, expected...)
	defer cleanup()

	actual, err := c.AllRoutes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(actual) != len(expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}

	for _, tc := range testCases {
		route, err := c.GetRoute(fmt.Sprintf("%s/%s", tc.namespace, tc.name))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if route == nil {
			t.Fatalf("expected non-nil route")
		}

		if route.Namespace() != tc.namespace {
			t.Errorf("expected %q, got %q", tc.namespace, route.Namespace())
		}

		if route.Name() != tc.name {
			t.Errorf("expected %q, got %q", tc.name, route.Name())
		}

		if route.Host() != tc.host {
			t.Errorf("expected %q, got %q", tc.host, route.Host())
		}

		if a, e := route.String(), fmt.Sprintf("%s/%s", tc.namespace, tc.name); a != e {
			t.Errorf("expected %q, got %q", e, a)
		}
	}
}

func TestControllerUnknownRoutes(t *testing.T) {
	c, cleanup := newTestController(t, mustCreateRoute(t, "foo", "bar", "foo.com"))
	defer cleanup()

	route, err := c.GetRoute("should/not-exist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if route != nil {
		t.Fatalf("expected nil route, got %v", route)
	}
}
