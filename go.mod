module github.com/frobware/route-monitor

go 1.12

replace (
	github.com/appscode/jsonpatch => gomodules.xyz/jsonpatch/v2 v2.0.1
	k8s.io/api => k8s.io/api v0.0.0-20191003000013-35e20aa79eb8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191003000419-f68efa97b39e
	sigs.k8s.io/controller-runtime => github.com/frobware/controller-runtime v0.2.0-beta.1.0.20191009100338-8e10fad09967
)

require (
	github.com/go-logr/logr v0.1.0
	github.com/openshift/api v3.9.1-0.20191008181517-e4fd21196097+incompatible
	github.com/prometheus/client_golang v1.1.0
	k8s.io/apimachinery v0.0.0-20191006235458-f9f2f3f8ab02
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20190920012459-5008bf6f8cd6 // indirect
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
)
