Version := $(shell date "+%Y%m%d%H%M")
GitCommit := $(shell git rev-parse HEAD)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"

run-dev: build 
	./bin/redis-gli -h 192.168.1.230

run: build 
	./bin/redis-gli

build:
	go build -race -ldflags $(LDFLAGS) -o bin/redis-gli *.go

release:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o release/redis-gli-darwin *.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS) -o release/redis-gli-win.exe *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o release/redis-gli-linux *.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags $(LDFLAGS) -o release/redis-gli-linux-arm *.go

.PHONY: run build release

