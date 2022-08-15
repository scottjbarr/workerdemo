GO ?= go

# command to build and run on the local OS.
GO_BUILD = go build

# command to compiling the distributable. Specify GOOS and GOARCH for the target OS.
GO_DIST = go build # GOOS=linux GOARCH=amd64 go build

.PHONY: dist

all: clean prepare dist

clean:
	rm -rf dist

prepare:
	mkdir -p dist

run:
	go run cmd/worker-demo/main.go

enqueue:
	./scripts/send.sh

install:
	$(GO) install ./cmd/...

dist: worker-demo

worker-demo:
	$(GO_DIST) -o dist/worker-demo cmd/worker-demo/main.go
