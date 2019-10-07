.PHONY: all bin/route-monitor

all: bin/route-monitor

bin/route-monitor:
	go build -gcflags=all="-N -l" -o $@ cmd/route-monitor/main.go

test:
	go test -race ./...

clean:
	$(RM) -r bin

run: bin/route-monitor
	./bin/route-monitor -kubeconfig $(KUBECONFIG) openshift-console/downloads openshift-console/console foo/does-not-exist
