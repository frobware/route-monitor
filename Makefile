.PHONY: all bin/route-monitor clean test

all: bin/route-monitor

binary: bin/route-monitor

bin/route-monitor:
	GO111MODULE=on go build -gcflags=all="-N -l" -o $@ cmd/route-monitor/main.go

test:
	GO111MODULE=on go test -v -race ./...

clean:
	$(RM) -r bin

ROUTES =						\
	openshift-console/console			\
	openshift-console/downloads			\

run: bin/route-monitor
	./bin/route-monitor -kubeconfig $(KUBECONFIG) openshift-console/downloads $(ROUTES) foo/does-not-exist
