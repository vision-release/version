APP := version
GOCACHE := $(CURDIR)/.cache/go-build

.PHONY: build test fmt

build:
	GOCACHE=$(GOCACHE) go -C src build -o ../bin/$(APP) .

test:
	GOCACHE=$(GOCACHE) go -C src test ./...

fmt:
	GOCACHE=$(GOCACHE) go -C src fmt ./...
