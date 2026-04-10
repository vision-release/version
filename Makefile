APP := version
GOCACHE := $(CURDIR)/.cache/go-build

.PHONY: build test integration-test fmt

build:
	GOCACHE=$(GOCACHE) go -C src build -o ../bin/$(APP) .

test:
	GOCACHE=$(GOCACHE) go -C src test ./...

integration-test:
	./scripts/integration_test.sh

fmt:
	GOCACHE=$(GOCACHE) go -C src fmt ./...
