.PHONY: all bin/route-monitor

all: bin/route-monitor

bin/route-monitor:
	go build -gcflags=all="-N -l" -o $@ cmd/main.go
