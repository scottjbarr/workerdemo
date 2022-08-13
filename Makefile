GO ?= go

# command to build and run on the local OS.
GO_BUILD = go build

# command to compiling the distributable. Specify GOOS and GOARCH for the target OS.
GO_DIST = go build # GOOS=linux GOARCH=amd64 go build

DB_DRIVER = postgres
DB_URL ?= "host=/var/run/postgresql user=scott dbname=dividends_test sslmode=disable"
GOOSE = goose -dir ./db $(DB_DRIVER) "$(DB_URL)"

.PHONY: dist

all: clean prepare dist

clean:
	rm -rf dist

prepare:
	mkdir -p dist

run:
	go run cmd/workflow-queue-example/main.go

enqueue:
	./scripts/send.sh

install:
	$(GO) install ./cmd/...

dist: workflow-queue-example

workflow-queue-example:
	$(GO_DIST) -o dist/workflow-queue-example cmd/workflow-queue-example/main.go
