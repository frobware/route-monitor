module github.com/frobware/route-monitor

go 1.12

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190425172711-65184652c889

require (
	github.com/prometheus/client_golang v1.1.0
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.2
)
