.PHONY: all bin/route-monitor

all: bin/route-monitor

bin/route-monitor:
	go build -gcflags=all="-N -l" -o $@ cmd/route-monitor/main.go

test:
	go test -v -race ./...

clean:
	$(RM) -r bin

ROUTES =						\
	openshift-console/console			\
	openshift-console/downloads			\

run: bin/route-monitor
	./bin/route-monitor -kubeconfig $(KUBECONFIG) openshift-console/downloads $(ROUTES) foo/does-not-exist
